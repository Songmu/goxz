package goxz

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type builder struct {
	name, version                               string
	platform                                    *platform
	output                                      string
	buildLdFlags, buildTags, buildInstallSuffix string
	pkgs                                        []string
	workDirBase                                 string
	zipAlways, static, trimpath                 bool
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
			return "", fmt.Errorf("go list failed with following output: %q", stderr.String())
		}
		pkgName := strings.TrimSpace(stdout.String())
		if pkgName != "main" {
			return "", fmt.Errorf("can't build artifact for non main package: %q", pkgName)
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
		// ref. https://github.com/golang/go/issues/26492#issuecomment-435462350
		if bdr.buildLdFlags != "" || bdr.static {
			var flags string
			if bdr.static {
				switch bdr.platform.os {
				case "freebsd", "netbsd", "linux", "windows":
					flags = `-extldflags "-static"`
				case "darwin":
					flags = `-s -extldflags "-sectcreate __TEXT __info_plist Info.plist"`
				case "android":
					flags = `-s`
				}
			}
			if bdr.buildLdFlags != "" {
				if flags == "" {
					flags = bdr.buildLdFlags
				} else {
					flags += " " + bdr.buildLdFlags
				}
			}
			if flags != "" {
				cmdArgs = append(cmdArgs, "-ldflags", flags)
			}
		}
		if bdr.buildTags != "" || bdr.static {
			var tags string
			if bdr.static {
				switch bdr.platform.os {
				case "windows", "freebsd", "netbsd":
					tags = "netgo"
				case "linux":
					tags = "netgo osusergo"
				}
			}
			if bdr.buildTags != "" {
				if tags == "" {
					tags = bdr.buildTags
				} else {
					tags += " " + bdr.buildTags
				}
			}
			if tags != "" {
				cmdArgs = append(cmdArgs, "-tags", tags)
			}
		}
		if bdr.trimpath {
			cmdArgs = append(cmdArgs, "-trimpath")
		}
		if bdr.buildInstallSuffix != "" {
			cmdArgs = append(cmdArgs, "-installsuffix", bdr.buildInstallSuffix)
		}
		cmdArgs = append(cmdArgs, pkg)

		cmd = exec.Command("go", cmdArgs...)
		cmd.Env = append(os.Environ(), "GOOS="+bdr.platform.os, "GOARCH="+bdr.platform.arch)
		bs, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf(
				"go build failed while building %q for %s/%s with following output:\n%s: %v",
				pkg, bdr.platform.os, bdr.platform.arch, string(bs), err)
		}
	}
	files, err := os.ReadDir(workDir)
	if err != nil {
		return "", err
	}
	if len(files) == 0 {
		return "", fmt.Errorf("no binaries are built from [%s] for %s/%s",
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

	var archiveFilePath string
	var archive func(sourceDir string, w io.Writer) error
	if bdr.zipAlways || bdr.platform.os == "windows" || bdr.platform.os == "darwin" {
		archiveFilePath = workDir + ".zip"
		archive = archiveZip
	} else {
		archiveFilePath = workDir + ".tar.gz"
		archive = archiveTarGz
	}

	f, err := os.Create(archiveFilePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	log.Printf("Archiving %s\n", filepath.Base(archiveFilePath))
	if err := archive(workDir, f); err != nil {
		return "", err
	}
	return archiveFilePath, nil
}

func archiveZip(sourceDir string, w io.Writer) error {
	zipWriter := zip.NewWriter(w)
	defer zipWriter.Close()

	baseName := filepath.Base(sourceDir)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		archivePath := filepath.Join(baseName, relPath)

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(archivePath)
		header.Method = zip.Deflate

		if info.IsDir() {
			header.Name += "/"
			header.Method = zip.Store
		}

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(writer, file)
		return err
	})
}

func archiveTarGz(sourceDir string, w io.Writer) error {
	gzipWriter := gzip.NewWriter(w)
	defer gzipWriter.Close()

	tarWriter := tar.NewWriter(gzipWriter)
	defer tarWriter.Close()

	baseName := filepath.Base(sourceDir)

	return filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, path)
		if err != nil {
			return err
		}

		archivePath := filepath.Join(baseName, relPath)

		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = filepath.ToSlash(archivePath)

		if err := tarWriter.WriteHeader(header); err != nil {
			return err
		}

		if info.IsDir() {
			return nil
		}

		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		_, err = io.Copy(tarWriter, file)
		return err
	})
}
