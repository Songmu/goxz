CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-X github.com/Songmu/goxz.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps:
	go install github.com/Songmu/gocredits/cmd/gocredits@v0.5.0

.PHONY: test
test: deps
	go test

.PHONY: build
build:
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/goxz

.PHONY: prepare-release
prepare-release: devel-deps
	go mod tidy
	gocredits . > CREDITS
	git update-index --add --remove -- go.mod go.sum CREDITS

.PHONY: crossbuild
crossbuild:
	go mod tidy -diff
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/goxz
	./goxz -pv=$(PACKAGE_VERSION) -static -build-ldflags=$(BUILD_LDFLAGS) \
		-d=./dist ./cmd/goxz
	cd ./dist && \
		shasum -a 256 -- * > SHA256SUMS
