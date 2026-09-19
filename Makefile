VERSION = $(shell godzil show-version)
CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-X github.com/Songmu/goxz.revision=$(CURRENT_REVISION)"
u := $(if $(update),-u)

.PHONY: deps
deps:
	go get ${u}
	go mod tidy

.PHONY: devel-deps
devel-deps: build
	go install github.com/Songmu/godzil/cmd/godzil@latest

.PHONY: test
test: deps
	go test

.PHONY: build
build: deps
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/goxz

CREDITS: deps devel-deps go.sum
	godzil credits -w

.PHONY: CREDITS crossbuild
crossbuild: devel-deps
	./goxz -pv=v$(VERSION) -static -build-ldflags=$(BUILD_LDFLAGS) \
        -d=./dist/v$(VERSION) ./cmd/goxz
	cd ./dist/v$(VERSION) && \
		for file in *; do \
			if [ -f "$$file" ] && [ "$$file" != SHA256SUMS ]; then \
				shasum -a 256 "$$file"; \
			fi; \
		done > SHA256SUMS
