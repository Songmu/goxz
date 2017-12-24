package goxz

import (
	"flag"
	"io"
	"log"
	"os"
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
	version, dest, output, os, arch, buildConstraints, buildLdFlags, buildTags string
	pkgs []string
}

func (cl *cli) run(args []string) error {
	log.SetOutput(cl.errStream)
	log.SetPrefix("[goxz] ")
	log.SetFlags(0)

	gx := &goxz{}
	fs := flag.NewFlagSet("goxz", flag.ContinueOnError)
	fs.SetOutput(cl.errStream)

	fs.StringVar(&gx.version, "pv", "", "Package version")
	fs.StringVar(&gx.dest, "d", "dist", "Destination directory")
	fs.StringVar(&gx.output, "o", "", "output")
	fs.StringVar(&gx.os, "os", "", "Specify OS (default is 'linux darwin windows')")
	fs.StringVar(&gx.arch, "arch", "", "Specify Arch (default is 'amd64')")
	fs.StringVar(&gx.buildConstraints, "bc", "", "Specify build constraints (e.g. 'linux,arm windows')")
	fs.StringVar(&gx.buildLdFlags, "build-ldflags", "", "arguments to pass on each go tool link invocation")
	fs.StringVar(&gx.buildTags, "build-tags", "", "a space-separated list of build `tags`")

	err := fs.Parse(args)
	if err != nil {
		return err
	}
	gx.pkgs = fs.Args()
	if len(gx.pkgs) == 0 {
		gx.pkgs = append(gx.pkgs, ".")
	}

	return nil
}
