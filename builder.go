package goxz

import (
	"os"
	"path/filepath"
	"strings"
)

type builder struct {
	platform                        *platform
	name, version                   string
	output, buildLdFlags, buildTags string
	pkgs                            []string
	projDir                         string
	workDirBase                     string
	zipAlways                       bool
}

func (bdr *builder) build() (string, error) {
	dirStuff := []string{bdr.name}
	if bdr.version != "" {
		dirStuff = append(dirStuff, bdr.version)
	}
	dirStuff = append(dirStuff, bdr.platform.os, bdr.platform.arch)
	dirname := strings.Join(dirStuff, "_")
	workDir := filepath.Join(bdr.workDirBase, dirname)
	if err := os.Mkdir(workDir, 0755); err != nil {
		return "", err
	}

	return "", nil
}
