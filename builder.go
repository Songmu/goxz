package goxz

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

type builder struct {
	name, version                   string
	platform                        *platform
	output, buildLdFlags, buildTags string
	pkgs                            []string
	workDirBase                     string
	zipAlways                       bool
	resources                       []string
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
		log.Printf("Building %s for %s/%s\n", pkg, bdr.platform.os, bdr.platform.arch)
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
		bs, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrapf(err,
				"go build failed while building %s for %s/%s with following output:\n%s",
				pkg, bdr.platform.os, bdr.platform.arch, string(bs))
		}
	}
	// TODO: build check. If the binaries are under workDir or not.

	for _, rc := range bdr.resources {
		dest := filepath.Join(workDir, filepath.Base(rc))
		if err := os.Link(rc, dest); err != nil {
			return "", err
		}
	}

	archiveFn := archiver.Zip.Make
	archiveFilePath := workDir + ".zip"
	if !bdr.zipAlways && bdr.platform.os != "windows" && bdr.platform.os != "darwin" {
		archiveFn = archiver.TarGz.Make
		archiveFilePath = workDir + ".tar.gz"
	}
	log.Printf("Archiving %s\n", filepath.Base(archiveFilePath))
	err := archiveFn(archiveFilePath, []string{workDir})
	if err != nil {
		return "", nil
	}
	return archiveFilePath, nil
}
