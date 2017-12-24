package goxz

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type builder struct {
	name, version                   string
	platform                        *platform
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

	for _, pkg := range bdr.pkgs {
		cmdArgs := []string{"build"}
		if bdr.output != "" {
			cmdArgs = append(cmdArgs, "-o", bdr.output)
		}
		if bdr.buildLdFlags != "" {
			cmdArgs = append(cmdArgs, "-ldflags", bdr.buildLdFlags)
		}
		if bdr.buildTags != "" {
			cmdArgs = append(cmdArgs, "-tags", bdr.buildTags)
		}
		cmdArgs = append(cmdArgs, pkg)

		cmd := exec.Command("go", cmdArgs...)
		cmd.Dir = workDir
		cmd.Env = append(os.Environ(), "GOOS="+bdr.platform.os, "GOARCH="+bdr.platform.arch)
		err := cmd.Run()
		if err != nil {
			return "", err
		}
	}

	return "", nil
}
