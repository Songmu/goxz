goxz
=======

[![Build Status](https://travis-ci.org/Songmu/goxz.png?branch=master)][travis]
[![Coverage Status](https://coveralls.io/repos/Songmu/goxz/badge.png?branch=master)][coveralls]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![GoDoc](https://godoc.org/github.com/Songmu/goxz?status.svg)][godoc]

[travis]: https://travis-ci.org/Songmu/goxz
[coveralls]: https://coveralls.io/r/Songmu/goxz?branch=master
[license]: https://github.com/Songmu/goxz/blob/master/LICENSE
[godoc]: https://godoc.org/github.com/Songmu/goxz

## Description

Just do cross building and archiving go tools conventionally

## Installation

    % go get github.com/Songmu/goxz/cmd/goxz

Built binaries are available on gihub releases.
<https://github.com/Songmu/goxz/releases>

## Concept

- Simple and Lightweight
- Provides `goxc` subset
  - only provides cross building and archiving
- No older Go support
- No complicated configuration and behaivors
  - convention over configuration

## Synopsis

```console
# in your repository
% gozx -v 0.0.1 -os=linux,darwin -arch=amd64 ./cmd/mytool [...]

# archives are built into `./goxz` directory
%  tree ./goxz
goxz/
├── yourapp_0.0.1_darwin_amd64.zip
└── ...
```

## Included resources

following files are included to archives automatically.

- `LICENSE*`
- `README*`
- `INSTALL*`
- `CREDIT*`
- `CHANGELOG*`

Custumizable file gathering rules may be provided in future.

## Archive naming specification

`{{Package}}_{{Version}}_{{OS}}_{{Arch}}.{{Ext}}`
or
`{{Package}}_{{OS}}_{{Arch}}.{{Ext}}`

- `{{Package}}`
  - directory name of the project by default
  - you can specify it with `-n` option
- `{{Version}}`
  - When the version is specified by `-pv` option, that is contained in archive name
- `{{Ext}}`
  - `.zip` is by default on "windows" and "darwin", `.tar.gz` is by default on other os.
  - use `-z` option to use zip always to compress.
- No file naming notations are available yet
  - ref. goxc: `{{.ExeName}}_{{.Version}}_{{.Os}}_{{.Arch}}{{.Ext}}`

## Options

- `os`
   - os: linux,darwin and windows by default
- `arch`
   - arc: arm64 only by default
- `-pv` for version specification
- `-build-ldflags` / `-build-tags`
- `-o` output filename
  - not compatible with multiple package building

## Author

[Songmu](https://github.com/Songmu)
