package goxz

import (
	"context"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/sync/errgroup"
)

// Run the goxz
func Run(ctx context.Context, args []string, outStream, errStream io.Writer) error {
	return (&cli{outStream: outStream, errStream: errStream}).run(args)
}

type goxz struct {
	os, arch                                    string
	name, version                               string
	dest                                        string
	include                                     string
	output                                      string
	buildLdFlags, buildTags, buildInstallSuffix string
	zipAlways                                   bool
	pkgs                                        []string
	static                                      bool
	work                                        bool
	trimpath                                    bool

	platforms []*platform
	projDir   string
	workDir   string
	resources []string
}

func (gx *goxz) run() error {
	err := gx.init()
	if err != nil {
		return err
	}

	gx.workDir, err = os.MkdirTemp(gx.dest, ".goxz-")
	if err != nil {
		return err
	}
	defer func() {
		if !gx.work {
			os.RemoveAll(gx.workDir)
		}
	}()
	if gx.work {
		log.Printf("working dir: %s\n", gx.workDir)
	}
	wd, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	if wd != gx.projDir {
		if err := os.Chdir(gx.projDir); err != nil {
			return err
		}
		defer os.Chdir(wd)
	}
	err = gx.buildAll()
	if err == nil {
		log.Println("Success!")
	}
	return err
}

func (gx *goxz) init() error {
	log.Println("Initializing...")
	if len(gx.pkgs) == 0 {
		gx.pkgs = append(gx.pkgs, ".")
	}
	if len(gx.pkgs) > 1 && gx.output != "" {
		return errors.New("When building multiple packages, output(`-o`) doesn't work")
	}

	if gx.projDir == "" {
		var err error
		gx.projDir, err = filepath.Abs(".")
		if err != nil {
			return err
		}
	} else if !filepath.IsAbs(gx.projDir) {
		p, err := filepath.Abs(gx.projDir)
		if err != nil {
			return err
		}
		gx.projDir = p
	}

	if gx.name == "" {
		gx.name = filepath.Base(gx.projDir)
	}

	if err := gx.initDest(); err != nil {
		return err
	}
	err := os.MkdirAll(gx.dest, 0755)
	if err != nil {
		return err
	}

	// fill the defaults
	if gx.os == "" {
		gx.os = "linux darwin windows"
	}
	if gx.arch == "" {
		gx.arch = "amd64 arm64"
	}
	gx.platforms, err = resolvePlatforms(gx.os, gx.arch)
	if err != nil {
		return err
	}

	gx.resources, err = gx.gatherResources()
	if err != nil {
		return err
	}
	rBaseNames := make([]string, len(gx.resources))
	for i, r := range gx.resources {
		rBaseNames[i], _ = filepath.Rel(gx.projDir, r)
	}
	log.Printf("Resources to include: [%s]\n", strings.Join(rBaseNames, " "))
	return nil
}

var separateReg = regexp.MustCompile(`\s*(?:\s+|,)\s*`)

func resolvePlatforms(os, arch string) ([]*platform, error) {
	platforms := []*platform{}
	osTargets := separateReg.Split(os, -1)
	archTargets := separateReg.Split(arch, -1)
	for _, os := range osTargets {
		if strings.TrimSpace(os) == "" {
			continue
		}
		for _, arch := range archTargets {
			if strings.TrimSpace(arch) == "" {
				continue
			}
			platforms = append(platforms, &platform{os: os, arch: arch})
		}
	}
	uniqPlatforms := []*platform{}
	seen := make(map[string]struct{})
	for _, pf := range platforms {
		key := pf.os + ":" + pf.arch
		_, ok := seen[key]
		if !ok {
			seen[key] = struct{}{}
			uniqPlatforms = append(uniqPlatforms, pf)
		}
	}
	return uniqPlatforms, nil
}

func (gx *goxz) initDest() error {
	if gx.dest == "" {
		gx.dest = "goxz"
	}
	if !filepath.IsAbs(gx.dest) {
		var err error
		gx.dest, err = filepath.Abs(gx.dest)
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	resourceReg = regexp.MustCompile(`(?i)^(?:readme|licen[sc]e|credits?|install|changelog)(?:\.|$)`)
	execExtReg  = regexp.MustCompile(`(?i)\.(?:[a-z]*sh|p[ly]|rb|exe|go)$`)
)

func (gx *goxz) gatherResources() ([]string, error) {
	dir := gx.projDir

	var ret []string
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !f.Type().IsRegular() {
			continue
		}
		n := f.Name()
		if resourceReg.MatchString(n) && !execExtReg.MatchString(n) {
			ret = append(ret, filepath.Join(dir, n))
		}
	}

	if gx.include != "" {
		for _, inc := range separateReg.Split(gx.include, -1) {
			if !filepath.IsAbs(inc) {
				inc = filepath.Join(dir, inc)
			}
			files, err := filepath.Glob(inc)
			if err != nil {
				return nil, err
			}
			for _, f := range files {
				if !filepath.IsAbs(f) {
					var err error
					f, err = filepath.Abs(f)
					if err != nil {
						return nil, err
					}
				}
				ret = append(ret, f)
			}
		}
	}

	seen := make(map[string]struct{})
	ret2 := make([]string, 0, len(ret))
	for _, p := range ret {
		_, ok := seen[p]
		if !ok {
			seen[p] = struct{}{}
			ret2 = append(ret2, p)
		}
	}
	return ret2, nil
}

func (gx *goxz) buildAll() error {
	eg := errgroup.Group{}
	for _, bdr := range gx.builders() {
		bdr := bdr
		eg.Go(func() error {
			archivePath, err := bdr.build()
			if err != nil {
				return err
			}
			installPath := filepath.Join(gx.dest, filepath.Base(archivePath))
			err = os.Rename(archivePath, installPath)
			if err != nil {
				return err
			}
			log.Printf("Artifact archived to %s\n", installPath)
			return nil
		})
	}
	return eg.Wait()
}

func (gx *goxz) builders() []*builder {
	builders := make([]*builder, len(gx.platforms))
	for i, pf := range gx.platforms {
		builders[i] = &builder{
			platform:           pf,
			name:               gx.name,
			version:            gx.version,
			output:             gx.output,
			buildLdFlags:       gx.buildLdFlags,
			buildTags:          gx.buildTags,
			buildInstallSuffix: gx.buildInstallSuffix,
			pkgs:               gx.pkgs,
			zipAlways:          gx.zipAlways,
			static:             gx.static,
			workDirBase:        gx.workDir,
			trimpath:           gx.trimpath,
			resources:          gx.resources,
			projDir:            gx.projDir,
		}
	}
	return builders
}
