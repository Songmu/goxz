package goxz

import (
	"io"
	"log"
	"os"
)

type cli struct {
	outStream, errStream io.Writer
}

const (
	exitCodeOK = iota
	exitCodeErr
)

// Run the goxz
func Run(args []string) int {
	err := (&cli{outStream: os.Stdout, errStream: os.Stderr}).run(args)
	if err != nil {
		log.Println(err)
		return exitCodeErr
	}
	return exitCodeOK
}

func (cl *cli) run(args []string) error {
	log.SetOutput(cl.errStream)
	return nil
}
