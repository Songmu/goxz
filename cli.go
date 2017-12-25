package goxz

import (
	"flag"
	"io"
	"log"
)

type cli struct {
	outStream, errStream io.Writer
}

func (cl *cli) run(args []string) error {
	log.SetOutput(cl.errStream)
	log.SetPrefix("[goxz] ")
	log.SetFlags(0)

	gx, err := cl.parseArgs(args)
	if err != nil {
		return err
	}
	return gx.run()
}

func (cl *cli) parseArgs(args []string) (*goxz, error) {
	gx := &goxz{}
	fs := flag.NewFlagSet("goxz", flag.ContinueOnError)
	fs.SetOutput(cl.errStream)

	fs.StringVar(&gx.name, "n", "", "Application name. By default this is the directory name.")
	fs.StringVar(&gx.dest, "d", "goxz", "Destination directory")
	fs.StringVar(&gx.version, "pv", "", "Package version")
	fs.StringVar(&gx.output, "o", "", "output")
	fs.StringVar(&gx.os, "os", "", "Specify OS (default is 'linux darwin windows')")
	fs.StringVar(&gx.arch, "arch", "", "Specify Arch (default is 'amd64')")
	fs.StringVar(&gx.buildLdFlags, "build-ldflags", "", "arguments to pass on each go tool link invocation")
	fs.StringVar(&gx.buildTags, "build-tags", "", "a space-separated list of build `tags`")
	fs.BoolVar(&gx.zipAlways, "zip", false, "zip always")

	fs.StringVar(&gx.projDir, "C", "", "[for debug] change directory")
	fs.BoolVar(&gx.work, "work", false, "[for debug] print the name of the temporary work directory and do not delete it when exiting.")

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
