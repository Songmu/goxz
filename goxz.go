package goxz

import (
	"flag"
	"io"
	"log"
	"os"
	"regexp"
)

type cli struct {
	outStream, errStream io.Writer
}

const (
	exitCodeOK = iota
	exitFlagErr
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
	version, dest, output, os, arch, buildLdFlags, buildTags string
	pkgs                                                     []string
	platforms                                                []platform
}

func (cl *cli) run(args []string) error {
	log.SetOutput(cl.errStream)
	log.SetPrefix("[goxz] ")
	log.SetFlags(0)

	gx, err := cl.parseArgs(args)
	if err != nil {
		return err
	}
	err = gx.init()
	if err != nil {
		return err
	}
	return nil
}

func (cl *cli) parseArgs(args []string) (*goxz, error) {
	gx := &goxz{}
	fs := flag.NewFlagSet("goxz", flag.ContinueOnError)
	fs.SetOutput(cl.errStream)

	fs.StringVar(&gx.version, "pv", "", "Package version")
	fs.StringVar(&gx.dest, "d", "dist", "Destination directory")
	fs.StringVar(&gx.output, "o", "", "output")
	fs.StringVar(&gx.os, "os", "", "Specify OS (default is 'linux darwin windows')")
	fs.StringVar(&gx.arch, "arch", "", "Specify Arch (default is 'amd64')")
	fs.StringVar(&gx.buildLdFlags, "build-ldflags", "", "arguments to pass on each go tool link invocation")
	fs.StringVar(&gx.buildTags, "build-tags", "", "a space-separated list of build `tags`")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	gx.pkgs = fs.Args()
	if len(gx.pkgs) == 0 {
		gx.pkgs = append(gx.pkgs, ".")
	}
	return gx, nil
}

var separateReg = regexp.MustCompile(`\s*,?\s*`)

func (gx *goxz) init() error {
	platforms := []platform{}
	if gx.os == "" {
		gx.os = "linux darwin windows"
	}
	if gx.arch == "" {
		gx.arch = "amd64"
	}

	osTargets := separateReg.Split(gx.os, -1)
	archTargets := separateReg.Split(gx.os, -1)
	for _, os := range osTargets {
		for _, arch := range archTargets {
			platforms = append(platforms, platform{os: os, arch: arch})
		}
	}

	// uniq and assign
	seen := make(map[string]struct{})
	for _, pf := range platforms {
		key := pf.os + ":" + pf.arch
		_, ok := seen[key]
		if !ok {
			seen[key] = struct{}{}
			gx.platforms = append(gx.platforms, pf)
		}
	}
	return nil
}
