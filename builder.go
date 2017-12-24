package goxz

type builder struct {
	platform                        *platform
	name, version                   string
	output, buildLdFlags, buildTags string
	pkgs                            []string
	projDir                         string
	zipAlways                       bool
}

func (bdr *builder) build() (string, error) {
	return "", nil
}
