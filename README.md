goxz
=======

[![Build Status](https://github.com/Songmu/goxz/workflows/test/badge.svg?branch=main)][actions]
[![Coverage Status](https://codecov.io/gh/Songmu/goxz/branch/main/graph/badge.svg)][codecov]
[![MIT License](http://img.shields.io/badge/license-MIT-blue.svg?style=flat-square)][license]
[![GoDev](https://pkg.go.dev/badge/github.com/Songmu/goxz)][godev]

[actions]: https://github.com//Songmu/goxz/actions?workflow=test
[codecov]: https://codecov.io/gh/Songmu/goxz
[license]: https://github.com/Songmu/goxz/blob/main/LICENSE
[godev]: https://pkg.go.dev/github.com/Songmu/goxz

GO + X (crossbuild) + Z (zip)

## Description

Just do cross building and archiving go tools conventionally

## Installation

    % go install github.com/Songmu/goxz/cmd/goxz@latest

Built binaries are available on gihub releases.
<https://github.com/Songmu/goxz/releases>

You can also install goxz with [aqua](https://aquaproj.github.io/).

    % aqua g -i Songmu/goxz

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
% goxz -pv 0.0.1 -os=linux,darwin -arch=amd64 ./cmd/mytool [...]

# archives are built into `./goxz` directory by default (configurable by `-d` option)
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

You can specify additional resources by using `-include` option.

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

- `-d` destination directory (`./goxz` by default)
- `-n` application name. by default the directory name is used.
- `-os` linux,darwin and windows by default
- `-arch` amd64,arm64 by default
- `-pv` for speicifing version (optional)
- `-o` output filename
  - not available with multiple package building
- `-z` to use zip always to compless
  - by default, zip is used on "windows and "darwin", tar.gz is used on other OS
- `-include` Include additional resources in archives
- `-build-ldflags` / `-build-tags`

## Author

[Songmu](https://github.com/Songmu)
