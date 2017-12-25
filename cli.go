package goxz

import (
	"flag"
	"fmt"
	"io"
	"log"
	"runtime"
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
	fs.Usage = func() {
		fmt.Fprintf(cl.errStream, `goxz - Just do cross building and archiving go tools conventionally

Version: %s (rev: %s/%s)

Synopsis:
    %% gozx -v 0.0.1 -os=linux,darwin -arch=amd64 ./cmd/mytool [...]

Options:
`, version, revision, runtime.Version())
		fs.PrintDefaults()
	}

	fs.StringVar(&gx.name, "n", "", "Application name. By default this is the directory name.")
	fs.StringVar(&gx.dest, "d", "goxz", "Destination directory")
	fs.StringVar(&gx.version, "pv", "", "Package version (optional)")
	fs.StringVar(&gx.output, "o", "", "output (optional)")
	fs.StringVar(&gx.os, "os", "", "Specify OS (default is 'linux darwin windows')")
	fs.StringVar(&gx.arch, "arch", "", "Specify Arch (default is 'amd64')")
	// TODO: fs.StringVar(&gx.buildConstraints, "build", "", "Specify build constraints")
	fs.StringVar(&gx.buildLdFlags, "build-ldflags", "", "arguments to pass on each go tool link invocation")
	fs.StringVar(&gx.buildTags, "build-tags", "", "a space-separated list of build `tags`")
	fs.BoolVar(&gx.zipAlways, "z", false, "zip always")
	fs.StringVar(&gx.projDir, "C", "", "specify the project directory. cwd by default")

	fs.BoolVar(&gx.work, "work", false, "[for debug] print the name of the temporary work directory and do not delete it when exiting.")

	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	gx.pkgs = fs.Args()
	return gx, nil
}
