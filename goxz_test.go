package goxz

import (
	"reflect"
	"testing"
)

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
