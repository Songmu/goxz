package main

import (
	"os"

	"github.com/Songmu/goxz"
)

func main() {
	os.Exit(goxz.Run(os.Args[1:]))
}
