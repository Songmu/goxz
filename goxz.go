package goxz

import (
	"flag"
	"go/build"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

const (
	exitCodeOK = iota
	exitCodeErr
)

// Run the goxz
func Run(args []string) int {
	err := (&cli{outStream: os.Stdout, errStream: os.Stderr}).run(args)
	if err != nil {
		if err == flag.ErrHelp {
			return exitCodeOK
		}
		log.Printf("[!!ERROR!!] %s\n", err)
		return exitCodeErr
	}
	return exitCodeOK
}

type goxz struct {
	os, arch                        string
	name, version                   string
	dest                            string
	include                         string
	output, buildLdFlags, buildTags string
	zipAlways                       bool
	pkgs                            []string
	work                            bool

	absPkgs   []string
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

	gx.workDir, err = ioutil.TempDir(gx.getDest(), ".goxz-")
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

	err := os.MkdirAll(gx.getDest(), 0755)
	if err != nil {
		return err
	}

	// fill the defaults
	if gx.os == "" {
		gx.os = "linux darwin windows"
	}
	if gx.arch == "" {
		gx.arch = "amd64"
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

	gx.absPkgs, err = goAbsPkgs(gx.pkgs, gx.projDir)
	if err != nil {
		return err
	}
	log.Printf("Package to build: [%s]\n", strings.Join(gx.absPkgs, " "))
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

func (gx *goxz) getDest() string {
	if gx.dest == "" {
		gx.dest = "goxz"
	}
	return gx.dest
}

func goAbsPkgs(pkgs []string, projDir string) ([]string, error) {
	var gosrcs []string
	for _, gopath := range filepath.SplitList(build.Default.GOPATH) {
		gosrcs = append(gosrcs, filepath.Join(filepath.Clean(gopath), "src"))
	}
	stuff := make([]string, len(pkgs))
	for i, pkg := range pkgs {
		if strings.HasPrefix(pkg, ".") {
			absPath := filepath.Clean(filepath.Join(projDir, pkg))
			for _, gosrc := range gosrcs {
				if strings.HasPrefix(absPath, gosrc) {
					p, err := filepath.Rel(gosrc, absPath)
					if err != nil {
						return nil, err
					}
					pkg = p
					break
				}
			}
		}
		stuff[i] = pkg
	}
	return stuff, nil
}

var resourceReg = regexp.MustCompile(`(?i)^(?:readme|license|credit|install|changelog)`)

func (gx *goxz) gatherResources() ([]string, error) {
	dir := gx.projDir

	var ret []string
	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !f.Mode().IsRegular() {
			continue
		}
		n := f.Name()
		if resourceReg.MatchString(n) && !strings.HasSuffix(n, ".go") {
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
			platform:     pf,
			name:         gx.name,
			version:      gx.version,
			output:       gx.output,
			buildLdFlags: gx.buildLdFlags,
			buildTags:    gx.buildTags,
			pkgs:         gx.absPkgs,
			zipAlways:    gx.zipAlways,
			workDirBase:  gx.workDir,
			resources:    gx.resources,
		}
	}
	return builders
}
