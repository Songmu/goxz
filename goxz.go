package goxz

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"text/template"
	"time"

	"golang.org/x/sync/errgroup"
)

// Run the goxz
func Run(ctx context.Context, args []string, outStream, errStream io.Writer) error {
	return (&cli{outStream: outStream, errStream: errStream}).run(args)
}

type goxz struct {
	os, arch                                    string
	name, version                               string
	dest                                        string
	include                                     string
	output                                      string
	buildLdFlags, buildTags, buildInstallSuffix string
	zipAlways                                   bool
	pkgs                                        []string
	static                                      bool
	work                                        bool
	trimpath                                    bool
	checksum                                    checksumFlag

	platforms           []*platform
	projDir             string
	workDir             string
	resources           []string
	executableResources map[string]struct{}
	archiveTimestamp    time.Time
	checksumFileName    string
}

const defaultChecksumFileName = "SHA256SUMS"

type checksumFlag struct {
	enabled bool
	pattern string
}

func (f *checksumFlag) Set(value string) error {
	switch value {
	case "true":
		f.enabled = true
		f.pattern = defaultChecksumFileName
	case "false":
		f.enabled = false
		f.pattern = ""
	default:
		f.enabled = true
		f.pattern = value
	}
	return nil
}

func (f *checksumFlag) String() string {
	if !f.enabled {
		return "false"
	}
	if f.pattern == defaultChecksumFileName {
		return "true"
	}
	return f.pattern
}

func (f *checksumFlag) IsBoolFlag() bool {
	return true
}

func (gx *goxz) run() error {
	err := gx.init()
	if err != nil {
		return err
	}

	gx.workDir, err = os.MkdirTemp(gx.dest, ".goxz-")
	if err != nil {
		return err
	}
	defer func() {
		if !gx.work {
			os.RemoveAll(gx.workDir)
		}
	}()
	if gx.work {
		log.Printf("working dir: %s\n", gx.workDir)
	}
	wd, err := filepath.Abs(".")
	if err != nil {
		return err
	}
	if wd != gx.projDir {
		if err := os.Chdir(gx.projDir); err != nil {
			return err
		}
		defer os.Chdir(wd)
	}
	archivePaths, err := gx.buildAll()
	if err != nil {
		return err
	}
	if gx.checksum.enabled {
		if err := gx.writeChecksumManifest(archivePaths); err != nil {
			return err
		}
	}
	log.Println("Success!")
	return nil
}

func (gx *goxz) init() error {
	log.Println("Initializing...")
	if len(gx.pkgs) == 0 {
		gx.pkgs = append(gx.pkgs, ".")
	}
	if len(gx.pkgs) > 1 && gx.output != "" {
		return errors.New("When building multiple packages, output(`-o`) doesn't work")
	}

	if gx.projDir == "" {
		var err error
		gx.projDir, err = filepath.Abs(".")
		if err != nil {
			return err
		}
	} else if !filepath.IsAbs(gx.projDir) {
		p, err := filepath.Abs(gx.projDir)
		if err != nil {
			return err
		}
		gx.projDir = p
	}

	if gx.name == "" {
		gx.name = filepath.Base(gx.projDir)
	}
	if err := gx.initChecksumFileName(); err != nil {
		return err
	}

	if err := gx.initDest(); err != nil {
		return err
	}
	err := os.MkdirAll(gx.dest, 0755)
	if err != nil {
		return err
	}

	// fill the defaults
	if gx.os == "" {
		gx.os = "linux darwin windows"
	}
	if gx.arch == "" {
		gx.arch = "amd64 arm64"
	}
	gx.platforms, err = resolvePlatforms(gx.os, gx.arch)
	if err != nil {
		return err
	}

	gx.resources, err = gx.gatherResources()
	if err != nil {
		return err
	}
	gx.executableResources, err = gitExecutableResources(gx.projDir, gx.resources)
	if err != nil {
		return err
	}
	gx.archiveTimestamp, err = archiveTimestamp(gx.projDir)
	if err != nil {
		return err
	}
	rBaseNames := make([]string, len(gx.resources))
	for i, r := range gx.resources {
		rBaseNames[i], _ = filepath.Rel(gx.projDir, r)
	}
	log.Printf("Resources to include: [%s]\n", strings.Join(rBaseNames, " "))
	return nil
}

var separateReg = regexp.MustCompile(`\s*(?:\s+|,)\s*`)

func resolvePlatforms(os, arch string) ([]*platform, error) {
	platforms := []*platform{}
	osTargets := separateReg.Split(os, -1)
	archTargets := separateReg.Split(arch, -1)
	for _, os := range osTargets {
		if strings.TrimSpace(os) == "" {
			continue
		}
		for _, arch := range archTargets {
			if strings.TrimSpace(arch) == "" {
				continue
			}
			platforms = append(platforms, &platform{os: os, arch: arch})
		}
	}
	uniqPlatforms := []*platform{}
	seen := make(map[string]struct{})
	for _, pf := range platforms {
		key := pf.os + ":" + pf.arch
		_, ok := seen[key]
		if !ok {
			seen[key] = struct{}{}
			uniqPlatforms = append(uniqPlatforms, pf)
		}
	}
	return uniqPlatforms, nil
}

func (gx *goxz) initDest() error {
	if gx.dest == "" {
		gx.dest = "goxz"
	}
	if !filepath.IsAbs(gx.dest) {
		var err error
		gx.dest, err = filepath.Abs(gx.dest)
		if err != nil {
			return err
		}
	}
	return nil
}

var (
	resourceReg = regexp.MustCompile(`(?i)^(?:readme|licen[sc]e|credits?|install|changelog)(?:\.|$)`)
	execExtReg  = regexp.MustCompile(`(?i)\.(?:[a-z]*sh|p[ly]|rb|exe|go)$`)
)

func (gx *goxz) gatherResources() ([]string, error) {
	dir := gx.projDir

	var ret []string
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}
	for _, f := range files {
		if !f.Type().IsRegular() {
			continue
		}
		n := f.Name()
		if resourceReg.MatchString(n) && !execExtReg.MatchString(n) {
			ret = append(ret, filepath.Join(dir, n))
		}
	}

	if gx.include != "" {
		for _, inc := range separateReg.Split(gx.include, -1) {
			if !filepath.IsAbs(inc) {
				inc = filepath.Join(dir, inc)
			}
			files, err := filepath.Glob(inc)
			if err != nil {
				return nil, err
			}
			for _, f := range files {
				if !filepath.IsAbs(f) {
					var err error
					f, err = filepath.Abs(f)
					if err != nil {
						return nil, err
					}
				}
				ret = append(ret, f)
			}
		}
	}

	seen := make(map[string]struct{})
	ret2 := make([]string, 0, len(ret))
	for _, p := range ret {
		_, ok := seen[p]
		if !ok {
			seen[p] = struct{}{}
			ret2 = append(ret2, p)
		}
	}
	return ret2, nil
}

func (gx *goxz) buildAll() ([]string, error) {
	eg := errgroup.Group{}
	builders := gx.builders()
	archivePaths := make([]string, len(builders))
	for i, bdr := range builders {
		i, bdr := i, bdr
		eg.Go(func() error {
			archivePath, err := bdr.build()
			if err != nil {
				return err
			}
			installPath := filepath.Join(gx.dest, filepath.Base(archivePath))
			err = os.Rename(archivePath, installPath)
			if err != nil {
				return err
			}
			archivePaths[i] = installPath
			log.Printf("Artifact archived to %s\n", installPath)
			return nil
		})
	}
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	return archivePaths, nil
}

func (gx *goxz) initChecksumFileName() error {
	if !gx.checksum.enabled {
		return nil
	}
	tmpl, err := template.New("checksum filename").Option("missingkey=error").Parse(gx.checksum.pattern)
	if err != nil {
		return fmt.Errorf("invalid checksum filename template: %w", err)
	}
	var name strings.Builder
	if err := tmpl.Execute(&name, struct {
		Name    string
		Version string
	}{
		Name:    gx.name,
		Version: gx.version,
	}); err != nil {
		return fmt.Errorf("render checksum filename template: %w", err)
	}
	gx.checksumFileName = name.String()
	if gx.checksumFileName == "" {
		return errors.New("checksum filename must not be empty")
	}
	if gx.checksumFileName == "." || gx.checksumFileName == ".." ||
		filepath.IsAbs(gx.checksumFileName) ||
		strings.ContainsAny(gx.checksumFileName, `/\`) ||
		strings.ContainsRune(gx.checksumFileName, 0) {
		return fmt.Errorf("checksum filename must be a basename: %q", gx.checksumFileName)
	}
	return nil
}

func (gx *goxz) writeChecksumManifest(archivePaths []string) error {
	sort.Slice(archivePaths, func(i, j int) bool {
		return filepath.Base(archivePaths[i]) < filepath.Base(archivePaths[j])
	})
	for _, archivePath := range archivePaths {
		if filepath.Base(archivePath) == gx.checksumFileName {
			return fmt.Errorf("checksum filename conflicts with archive: %q", gx.checksumFileName)
		}
	}

	tmp, err := os.CreateTemp(gx.dest, "."+gx.checksumFileName+".tmp-*")
	if err != nil {
		return fmt.Errorf("create checksum manifest: %w", err)
	}
	tmpPath := tmp.Name()
	renamed := false
	defer func() {
		if !renamed {
			os.Remove(tmpPath)
		}
	}()

	writer := bufio.NewWriter(tmp)
	for _, archivePath := range archivePaths {
		file, err := os.Open(archivePath)
		if err != nil {
			tmp.Close()
			return fmt.Errorf("open archive for checksum %q: %w", archivePath, err)
		}
		hash := sha256.New()
		_, copyErr := io.Copy(hash, file)
		closeErr := file.Close()
		if copyErr != nil {
			tmp.Close()
			return fmt.Errorf("hash archive %q: %w", archivePath, copyErr)
		}
		if closeErr != nil {
			tmp.Close()
			return fmt.Errorf("close archive %q: %w", archivePath, closeErr)
		}
		if _, err := fmt.Fprintf(writer, "%x  %s\n", hash.Sum(nil), filepath.Base(archivePath)); err != nil {
			tmp.Close()
			return fmt.Errorf("write checksum manifest: %w", err)
		}
	}
	if err := writer.Flush(); err != nil {
		tmp.Close()
		return fmt.Errorf("flush checksum manifest: %w", err)
	}
	if err := tmp.Chmod(0644); err != nil {
		tmp.Close()
		return fmt.Errorf("set checksum manifest mode: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("close checksum manifest: %w", err)
	}

	manifestPath := filepath.Join(gx.dest, gx.checksumFileName)
	if err := os.Rename(tmpPath, manifestPath); err != nil {
		return fmt.Errorf("install checksum manifest: %w", err)
	}
	renamed = true
	log.Printf("Checksum manifest written to %s\n", manifestPath)
	return nil
}

func archiveTimestamp(projDir string) (time.Time, error) {
	if sourceDateEpoch := os.Getenv("SOURCE_DATE_EPOCH"); sourceDateEpoch != "" {
		seconds, err := strconv.ParseInt(sourceDateEpoch, 10, 64)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid SOURCE_DATE_EPOCH %q: %w", sourceDateEpoch, err)
		}
		return time.Unix(seconds, 0).UTC(), nil
	}

	cmd := exec.Command("git", "-C", projDir, "log", "-1", "--format=%ct")
	output, err := cmd.Output()
	if err != nil {
		return time.Time{}, nil
	}
	seconds, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil {
		return time.Time{}, nil
	}
	return time.Unix(seconds, 0).UTC(), nil
}

func gitExecutableResources(projDir string, resources []string) (map[string]struct{}, error) {
	executableResources := make(map[string]struct{})
	args := []string{"-C", projDir, "--literal-pathspecs", "ls-files", "--stage", "-z", "--"}
	for _, resource := range resources {
		rel, err := filepath.Rel(projDir, resource)
		if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			continue
		}
		args = append(args, filepath.ToSlash(rel))
	}
	if len(args) == 7 {
		return executableResources, nil
	}

	output, err := exec.Command("git", args...).Output()
	if err != nil {
		return executableResources, nil
	}
	for _, entry := range bytes.Split(output, []byte{0}) {
		if len(entry) == 0 {
			continue
		}
		parts := bytes.SplitN(entry, []byte{'\t'}, 2)
		fields := bytes.Fields(parts[0])
		if len(parts) != 2 || len(fields) != 3 {
			return nil, fmt.Errorf("unexpected git ls-files output: %q", entry)
		}
		if string(fields[0]) == "100755" {
			path := filepath.Join(projDir, filepath.FromSlash(string(parts[1])))
			executableResources[filepath.Clean(path)] = struct{}{}
		}
	}
	return executableResources, nil
}

func (gx *goxz) builders() []*builder {
	builders := make([]*builder, len(gx.platforms))
	for i, pf := range gx.platforms {
		builders[i] = &builder{
			platform:            pf,
			name:                gx.name,
			version:             gx.version,
			output:              gx.output,
			buildLdFlags:        gx.buildLdFlags,
			buildTags:           gx.buildTags,
			buildInstallSuffix:  gx.buildInstallSuffix,
			pkgs:                gx.pkgs,
			zipAlways:           gx.zipAlways,
			static:              gx.static,
			workDirBase:         gx.workDir,
			trimpath:            gx.trimpath,
			resources:           gx.resources,
			executableResources: gx.executableResources,
			projDir:             gx.projDir,
			archiveTimestamp:    gx.archiveTimestamp,
		}
	}
	return builders
}
