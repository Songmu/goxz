package goxz

import (
	"context"
	"flag"
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

func TestGitExecutableResources(t *testing.T) {
	projDir := t.TempDir()
	for _, name := range []string{"script.sh", "README", "untracked.sh"} {
		if err := os.WriteFile(filepath.Join(projDir, name), []byte(name+"\n"), 0600); err != nil {
			t.Fatal(err)
		}
	}
	for _, args := range [][]string{
		{"init", "--quiet"},
		{"add", "script.sh", "README", "untracked.sh"},
		{"update-index", "--chmod=+x", "script.sh"},
		{"update-index", "--chmod=-x", "README"},
		{"reset", "--quiet", "untracked.sh"},
	} {
		cmd := exec.Command("git", args...)
		cmd.Dir = projDir
		if output, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v\n%s", args, err, output)
		}
	}

	resources := []string{
		filepath.Join(projDir, "script.sh"),
		filepath.Join(projDir, "README"),
		filepath.Join(projDir, "untracked.sh"),
	}
	executable, err := gitExecutableResources(projDir, resources)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := executable[filepath.Join(projDir, "script.sh")]; !ok {
		t.Error("script.sh is not executable")
	}
	for _, name := range []string{"README", "untracked.sh"} {
		if _, ok := executable[filepath.Join(projDir, name)]; ok {
			t.Errorf("%s is executable", name)
		}
	}
}
