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
		if err != flag.ErrHelp {
			log.Println(err)
		}
		return exitCodeErr
	}
	return exitCodeOK
}

type goxz struct {
	os, arch                        string
	name, version                   string
	dest                            string
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

	err = gx.prepareWorkdir()
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
	return gx.buildAll()
}

func (gx *goxz) init() error {
	log.Println("Initializing...")
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

	// TODO: implement build constraints
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
	gx.resources, err = gatherResources(gx.projDir)
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
	if len(gx.absPkgs) == 1 {
		log.Printf("Package to build: %s\n", gx.absPkgs[0])
	} else {
		log.Printf("Packages to build:\n  - %s\n", strings.Join(gx.absPkgs, "\n  - "))
	}
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

var resourceReg = regexp.MustCompile(`(?i)^(?:readme|license|credit|install)`)

func gatherResources(dir string) ([]string, error) {
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
	return ret, nil
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
			return os.Rename(archivePath, filepath.Join(gx.dest, filepath.Base(archivePath)))
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

func (gx *goxz) prepareWorkdir() error {
	tmpd, err := ioutil.TempDir(gx.getDest(), ".goxz-")
	gx.workDir = tmpd
	return err
}
