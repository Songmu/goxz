package goxz

import (
	"context"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestRun_help(t *testing.T) {
	err := Run(context.Background(), []string{"-h"}, io.Discard, io.Discard)
	if err != flag.ErrHelp {
		t.Errorf("somthing went wrong: %s", err)
	}
}

func TestResolvePlatforms(t *testing.T) {
	testCases := []struct {
		name   string
		inOS   string
		inArch string
		expect []platform
	}{
		{
			name:   "simple",
			inOS:   "linux",
			inArch: "amd64",
			expect: []platform{{"linux", "amd64"}},
		},
		{
			name:   "comma separated 2 os and whitespece separated 2 arch",
			inOS:   "linux,windows",
			inArch: "amd64 386",
			expect: []platform{
				{"linux", "amd64"},
				{"linux", "386"},
				{"windows", "amd64"},
				{"windows", "386"},
			},
		},
		{
			name:   "empty OS",
			inOS:   "",
			inArch: "amd64 386",
			expect: []platform{},
		},
		{
			name:   "empty Arch",
			inOS:   "linux",
			inArch: "",
			expect: []platform{},
		},
		{
			name:   "mixed separators",
			inOS:   "linux ,windows darwin ",
			inArch: "amd64  386,     arm",
			expect: []platform{
				{"linux", "amd64"},
				{"linux", "386"},
				{"linux", "arm"},
				{"windows", "amd64"},
				{"windows", "386"},
				{"windows", "arm"},
				{"darwin", "amd64"},
				{"darwin", "386"},
				{"darwin", "arm"},
			},
		},
	}
	for _, tc := range testCases {
		o, err := resolvePlatforms(tc.inOS, tc.inArch)
		if err != nil {
			t.Errorf("error should be nil but: %s", err)
		}
		out := []platform{}
		for _, pf := range o {
			out = append(out, *pf)
		}
		if !reflect.DeepEqual(out, tc.expect) {
			t.Errorf("wrong resolvePlatform (%s)\n  out: %v\nexpect: %v", tc.name, out, tc.expect)
		}
	}
}

func TestGatherResources(t *testing.T) {
	projDir, _ := filepath.Abs("./testdata")
	gx := &goxz{
		projDir: projDir,
		include: "sample.*",
	}
	files, err := gx.gatherResources()
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, len(files))
	for i, r := range files {
		out[i], _ = filepath.Rel(projDir, r)
	}
	expect := []string{"CREDITS", "LICENSE.txt", "README.md", "sample.conf"}
	sort.Strings(expect)
	sort.Strings(out)
	if !reflect.DeepEqual(out, expect) {
		t.Errorf("something went wrong:\n  out: %v\nexpect: %v", out, expect)
	}
}

func TestInitChecksumFileName(t *testing.T) {
	tests := []struct {
		name    string
		pattern string
		want    string
		wantErr bool
	}{
		{
			name:    "default",
			pattern: defaultChecksumFileName,
			want:    defaultChecksumFileName,
		},
		{
			name:    "template",
			pattern: "{{.Name}}_{{.Version}}_checksums.txt",
			want:    "goxz_v1.2.3_checksums.txt",
		},
		{name: "malformed template", pattern: "{{", wantErr: true},
		{name: "unknown field", pattern: "{{.Missing}}", wantErr: true},
		{name: "empty filename", pattern: "", wantErr: true},
		{name: "parent path", pattern: "../SHA256SUMS", wantErr: true},
		{name: "subdirectory", pattern: "checksums/SHA256SUMS", wantErr: true},
		{name: "windows subdirectory", pattern: `checksums\SHA256SUMS`, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gx := &goxz{
				name:     "goxz",
				version:  "v1.2.3",
				checksum: checksumFlag{enabled: true, pattern: tt.pattern},
			}
			err := gx.initChecksumFileName()
			if tt.wantErr {
				if err == nil {
					t.Fatal("initChecksumFileName() error = nil")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if gx.checksumFileName != tt.want {
				t.Errorf("checksumFileName = %q, want %q", gx.checksumFileName, tt.want)
			}
		})
	}
}

func TestWriteChecksumManifest(t *testing.T) {
	dest := t.TempDir()
	archives := []struct {
		name    string
		content string
	}{
		{name: "z.zip", content: "zip"},
		{name: "a.tar.gz", content: "tar"},
	}
	var paths []string
	for _, archive := range archives {
		path := filepath.Join(dest, archive.name)
		if err := os.WriteFile(path, []byte(archive.content), 0644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, path)
	}
	if err := os.WriteFile(filepath.Join(dest, "stale.txt"), []byte("stale"), 0644); err != nil {
		t.Fatal(err)
	}

	gx := &goxz{dest: dest, checksumFileName: defaultChecksumFileName}
	if err := gx.writeChecksumManifest(paths); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dest, defaultChecksumFileName))
	if err != nil {
		t.Fatal(err)
	}
	want := fmt.Sprintf(
		"%x  a.tar.gz\n%x  z.zip\n",
		sha256.Sum256([]byte("tar")),
		sha256.Sum256([]byte("zip")),
	)
	if string(got) != want {
		t.Errorf("checksum manifest = %q, want %q", got, want)
	}
}

func TestGitExecutableResources(t *testing.T) {
	projDir := t.TempDir()
	for _, name := range []string{"script[1].sh", "script1.sh", ":script.sh", "README", "untracked.sh"} {
		if err := os.WriteFile(filepath.Join(projDir, name), []byte(name+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{
		{"init", "--quiet"},
		{"--literal-pathspecs", "add", "--", "script[1].sh", "script1.sh", ":script.sh", "README"},
		{"--literal-pathspecs", "update-index", "--chmod=+x", "--", "script[1].sh", ":script.sh"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = projDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, output)
		}
	}

	resources := []string{
		filepath.Join(projDir, "script[1].sh"),
		filepath.Join(projDir, "script1.sh"),
		filepath.Join(projDir, ":script.sh"),
		filepath.Join(projDir, "README"),
		filepath.Join(projDir, "untracked.sh"),
	}
	executable, err := gitExecutableResources(projDir, resources)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"script[1].sh", ":script.sh"} {
		if _, ok := executable[filepath.Join(projDir, name)]; !ok {
			t.Errorf("%s is not executable", name)
		}
	}
	for _, name := range []string{"script1.sh", "README", "untracked.sh"} {
		if _, ok := executable[filepath.Join(projDir, name)]; ok {
			t.Errorf("%s is executable", name)
		}
	}
}
