package goxz

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func setup(t *testing.T) string {
	tmpd, err := os.MkdirTemp("", "goxz-")
	if err != nil {
		t.Fatal(err)
	}
	return tmpd
}

func TestCliRun(t *testing.T) {
	testCases := []struct {
		name   string
		input  []string
		files  []string
		errStr string
	}{
		{
			name:  "simple",
			input: []string{"./testdata/hello"},
			files: []string{
				"goxz_darwin_amd64.zip",
				"goxz_darwin_arm64.zip",
				"goxz_linux_amd64.tar.gz",
				"goxz_linux_arm64.tar.gz",
				"goxz_windows_amd64.zip",
				"goxz_windows_arm64.zip",
			},
		},
		{
			name:  "zip always and specify multi arch",
			input: []string{"-z", "-os=freebsd,linux", "-arch=386 amd64", "./testdata/hello"},
			files: []string{
				"goxz_freebsd_amd64.zip",
				"goxz_freebsd_386.zip",
				"goxz_linux_amd64.zip",
				"goxz_linux_386.zip",
			},
		},
		{
			name:  "zip always, static and specify multi os",
			input: []string{"-z", "-static", "-os=darwin,linux,freebsd,windows", "./testdata/hello"},
			files: []string{
				"goxz_darwin_amd64.zip",
				"goxz_darwin_arm64.zip",
				"goxz_linux_amd64.zip",
				"goxz_linux_arm64.zip",
				"goxz_windows_amd64.zip",
				"goxz_windows_arm64.zip",
				"goxz_freebsd_amd64.zip",
				"goxz_freebsd_arm64.zip",
			},
		},
		{
			name:  "build multiple pakcages with app name",
			input: []string{"-n=abc", "-os=linux", "-arch=amd64", "./testdata/hello", "./cmd/goxz"},
			files: []string{"abc_linux_amd64.tar.gz"},
		},
		{
			name:  "output option with version",
			input: []string{"-o=abc", "-C=.", "-pv=0.1.1", "-os=freebsd", "./testdata/hello"},
			files: []string{
				"goxz_0.1.1_freebsd_amd64.tar.gz",
				"goxz_0.1.1_freebsd_arm64.tar.gz",
			},
		},
		{
			name:   "[error] no resulting object",
			input:  []string{}, // same as []string{"."}
			errStr: `can't build artifact for non main package: "goxz"`,
		},
		{
			name:   "[error] multiple packages and -o flag are not compatible",
			input:  []string{"-o=hoge", "./testdata/hello", "./cmd/goxz"},
			errStr: "When building multiple packages",
		},
		{
			name:   "[error] package not exists",
			input:  []string{"-work", "--checksum", "./testdata/hello___"},
			errStr: "go list failed with following output",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cl := &cli{outStream: io.Discard, errStream: io.Discard}
			tmpd := setup(t)
			defer os.RemoveAll(tmpd)
			args := append([]string{"-d=" + tmpd}, tc.input...)
			err := cl.run(args)
			if tc.errStr == "" {
				if err != nil {
					t.Errorf("%s: error should be nil but: %s", tc.name, err)
				}
			} else {
				if err == nil {
					t.Errorf("%s: error should be occured but nil", tc.name)
				} else if !strings.Contains(err.Error(), tc.errStr) {
					t.Errorf("%s: error should be contains %q, but %q", tc.name, tc.errStr, err)
				}
			}
			files, err := os.ReadDir(tmpd)
			if err != nil {
				t.Fatal(err)
			}
			var outs []string
			for _, f := range files {
				if !f.IsDir() {
					outs = append(outs, f.Name())
				}
			}
			sort.Strings(tc.files)
			sort.Strings(outs)
			if !reflect.DeepEqual(tc.files, outs) {
				t.Errorf("%s: files are not built correctly\n   out: %v\nexpect: %v",
					tc.name, outs, tc.files)
			}
		})
	}
}

func TestCliRun_projDir(t *testing.T) {
	if err := os.Chdir("./testdata"); err != nil {
		t.Fatal(err)
	}
	defer os.Chdir("../")

	input := []string{"-o=abc", "-C=../", "-pv=0.1.1", "-os=freebsd", "./testdata/hello"}
	builtFiles := []string{
		"goxz_0.1.1_freebsd_amd64.tar.gz",
		"goxz_0.1.1_freebsd_arm64.tar.gz",
	}

	cl := &cli{outStream: io.Discard, errStream: io.Discard}
	tmpd := setup(t)
	// This deletion process is performed to check whether goxz itself creates
	// a directory correctly
	if err := os.RemoveAll(tmpd); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpd)
	args := append([]string{"-d=" + tmpd}, input...)
	err := cl.run(args)

	if err != nil {

		t.Errorf("error should be nil but: %s", err)
	}

	files, err := os.ReadDir(tmpd)
	if err != nil {
		t.Fatal(err)
	}
	var outs []string
	for _, f := range files {
		if !f.IsDir() {
			outs = append(outs, f.Name())
		}
	}
	if !reflect.DeepEqual(builtFiles, outs) {
		t.Errorf("files are not built correctly\n   out: %v\nexpect: %v", outs, builtFiles)
	}

}

func TestParseArgs_checksum(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		enabled bool
		pattern string
	}{
		{name: "omitted"},
		{name: "bare", args: []string{"--checksum"}, enabled: true, pattern: defaultChecksumFileName},
		{name: "true", args: []string{"--checksum=true"}, enabled: true, pattern: defaultChecksumFileName},
		{name: "false", args: []string{"--checksum=false"}},
		{
			name:    "template",
			args:    []string{"--checksum={{.Name}}_{{.Version}}_checksums.txt"},
			enabled: true,
			pattern: "{{.Name}}_{{.Version}}_checksums.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cl := &cli{outStream: io.Discard, errStream: io.Discard}
			gx, err := cl.parseArgs(tt.args)
			if err != nil {
				t.Fatal(err)
			}
			if gx.checksum.enabled != tt.enabled || gx.checksum.pattern != tt.pattern {
				t.Errorf("checksum = %#v, want enabled=%t pattern=%q", gx.checksum, tt.enabled, tt.pattern)
			}
		})
	}
}

func TestCliRun_checksum(t *testing.T) {
	tests := []struct {
		name         string
		checksumArg  string
		version      string
		manifestName string
	}{
		{
			name:         "default filename",
			checksumArg:  "--checksum",
			manifestName: defaultChecksumFileName,
		},
		{
			name:         "filename template",
			checksumArg:  "--checksum={{.Name}}_{{.Version}}_checksums.txt",
			version:      "v1.2.3",
			manifestName: "hello_v1.2.3_checksums.txt",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dest := t.TempDir()
			if err := os.WriteFile(filepath.Join(dest, "stale.txt"), []byte("stale"), 0644); err != nil {
				t.Fatal(err)
			}

			cl := &cli{outStream: io.Discard, errStream: io.Discard}
			err := cl.run([]string{
				"-d", dest,
				"-n", "hello",
				"-pv", tt.version,
				"-os", "linux",
				"-arch", "amd64",
				tt.checksumArg,
				"./testdata/hello",
			})
			if err != nil {
				t.Fatal(err)
			}

			archiveName := "hello_linux_amd64.tar.gz"
			if tt.version != "" {
				archiveName = "hello_" + tt.version + "_linux_amd64.tar.gz"
			}
			archive, err := os.ReadFile(filepath.Join(dest, archiveName))
			if err != nil {
				t.Fatal(err)
			}
			want := fmt.Sprintf("%x  %s\n", sha256.Sum256(archive), archiveName)
			got, err := os.ReadFile(filepath.Join(dest, tt.manifestName))
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != want {
				t.Errorf("checksum manifest = %q, want %q", got, want)
			}
		})
	}
}
