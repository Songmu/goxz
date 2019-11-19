package main

import (
	"context"
	"flag"
	"log"
	"os"

	"github.com/Songmu/goxz"
)

func main() {
	err := goxz.Run(context.Background(), os.Args[1:], os.Stdout, os.Stderr)
	if err != nil && err != flag.ErrHelp {
		log.Printf("[!!ERROR!!] %s\n", err)
		os.Exit(1)
	}
}
