CURRENT_REVISION = $(shell git rev-parse --short HEAD)
BUILD_LDFLAGS = "-X github.com/Songmu/goxz.revision=$(CURRENT_REVISION)"
ifdef update
  u=-u
endif

deps:
	go get ${u} github.com/golang/dep/cmd/dep
	dep ensure

devel-deps: deps
	go get ${u} github.com/golang/lint/golint
	go get ${u} github.com/mattn/goveralls
	go get ${u} github.com/motemen/gobump
	go get ${u} github.com/Songmu/goxz
	go get ${u} github.com/Songmu/ghch
	go get ${u} github.com/tcnksm/ghr

test: deps
	go test

lint: devel-deps
	go vet
	golint -set_exit_status

cover: devel-deps
	goveralls

build: deps
	go build -ldflags=$(BUILD_LDFLAGS) ./cmd/goxz

crossbuild: devel-deps
	goxz -pv=v$(shell gobump show -r) -build-ldflags=$(BUILD_LDFLAGS) \
	  -d=./dist/v$(shell gobump show -r)

release:
	_tools/releng
	_tools/upload_artifacts

.PHONY: test deps devel-deps lint cover crossbuild release
