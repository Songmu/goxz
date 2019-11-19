package goxz

import (
	"context"
	"flag"
	"io/ioutil"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

func TestRun_help(t *testing.T) {
	err := Run(context.Background(), []string{"-h"}, ioutil.Discard, ioutil.Discard)
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
	expect := []string{"LICENSE.txt", "README.md", "sample.conf"}
	sort.Strings(expect)
	sort.Strings(out)
	if !reflect.DeepEqual(out, expect) {
		t.Errorf("something went wrong:\n  out: %v\nexpect: %v", out, expect)
	}
}
