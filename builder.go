package goxz

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mholt/archiver"
	"github.com/pkg/errors"
)

type builder struct {
	name, version                               string
	platform                                    *platform
	output                                      string
	buildLdFlags, buildTags, buildInstallSuffix string
	pkgs                                        []string
	workDirBase                                 string
	zipAlways                                   bool
	resources                                   []string
	projDir                                     string
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
		var stdout, stderr bytes.Buffer
		cmd := exec.Command("go", "list", "-f", "{{.Name}}", pkg)
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			return "", errors.Errorf("go list failed with following output: %q", stderr.String())
		}
		pkgName := strings.TrimSpace(stdout.String())
		if pkgName != "main" {
			return "", errors.Errorf("can't build artifact for non main package: %q", pkgName)
		}
		output := bdr.output
		if output == "" {
			output = filepath.Base(pkg)
			if output == "." {
				wd, err := os.Getwd()
				if err != nil {
					return "", err
				}
				output = filepath.Base(wd)
			}
			if bdr.platform.os == "windows" {
				output += ".exe"
			}
		}
		cmdArgs := []string{"build", "-o", filepath.Join(workDir, output)}
		if bdr.buildLdFlags != "" {
			cmdArgs = append(cmdArgs, "-ldflags", bdr.buildLdFlags)
		}
		if bdr.buildTags != "" {
			cmdArgs = append(cmdArgs, "-tags", bdr.buildTags)
		}
		if bdr.buildInstallSuffix != "" {
			cmdArgs = append(cmdArgs, "-installsuffix", bdr.buildInstallSuffix)
		}
		cmdArgs = append(cmdArgs, pkg)

		cmd = exec.Command("go", cmdArgs...)
		cmd.Env = append(os.Environ(), "GOOS="+bdr.platform.os, "GOARCH="+bdr.platform.arch)
		bs, err := cmd.CombinedOutput()
		if err != nil {
			return "", errors.Wrapf(err,
				"go build failed while building %q for %s/%s with following output:\n%s",
				pkg, bdr.platform.os, bdr.platform.arch, string(bs))
		}
	}
	files, err := ioutil.ReadDir(workDir)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", errors.Errorf("No binaries are built from [%s] for %s/%s",
			strings.Join(bdr.pkgs, " "), bdr.platform.os, bdr.platform.arch)
	}

	for _, rc := range bdr.resources {
		rel, _ := filepath.Rel(bdr.projDir, rc)
		dest := filepath.Join(workDir, rel)
		if err := os.MkdirAll(filepath.Dir(dest), 0755); err != nil {
			return "", err
		}
		if err := os.Link(rc, dest); err != nil {
			return "", err
		}
	}

	var arch archiver.Archiver = &archiver.Zip{
		CompressionLevel:     flate.DefaultCompression,
		MkdirAll:             true,
		SelectiveCompression: true,
	}
	archiveFilePath := workDir + ".zip"
	if !bdr.zipAlways && bdr.platform.os != "windows" && bdr.platform.os != "darwin" {
		arch = &archiver.TarGz{
			CompressionLevel: gzip.DefaultCompression,
			Tar: &archiver.Tar{
				MkdirAll: true,
			},
		}
		archiveFilePath = workDir + ".tar.gz"
	}
	log.Printf("Archiving %s\n", filepath.Base(archiveFilePath))
	err = arch.Archive([]string{workDir}, archiveFilePath)
	if err != nil {
		return "", err
	}
	return archiveFilePath, nil
}
