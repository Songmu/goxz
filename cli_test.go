package goxz

import (
	"io/ioutil"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func setup(t *testing.T) string {
	tmpd, err := ioutil.TempDir("", "goxz-")
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
				"goxz_linux_amd64.tar.gz",
				"goxz_windows_amd64.zip",
			},
		},
		{
			name:  "zip always and specify multi arch",
			input: []string{"-z", "-os=darwin,linux", "-arch=386 amd64", "./testdata/hello"},
			files: []string{
				"goxz_darwin_amd64.zip",
				"goxz_darwin_386.zip",
				"goxz_linux_amd64.zip",
				"goxz_linux_386.zip",
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
			files: []string{"goxz_0.1.1_freebsd_amd64.tar.gz"},
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
			input:  []string{"-work", "./testdata/hello___"},
			errStr: "can't load package",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cl := &cli{outStream: ioutil.Discard, errStream: ioutil.Discard}
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
			files, err := ioutil.ReadDir(tmpd)
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
	builtFiles := []string{"goxz_0.1.1_freebsd_amd64.tar.gz"}

	cl := &cli{outStream: ioutil.Discard, errStream: ioutil.Discard}
	tmpd := setup(t)
	defer os.RemoveAll(tmpd)
	args := append([]string{"-d=" + tmpd}, input...)
	err := cl.run(args)

	if err != nil {

		t.Errorf("error should be nil but: %s", err)
	}

	files, err := ioutil.ReadDir(tmpd)
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
		t.Errorf("files are not built correctly\n   out: %v\nexpect: %v", outs, files)
	}

}
