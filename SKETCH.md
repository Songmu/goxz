## goxz

Just do cross building and archiving go tools conventionally

- provides `goxc` subset
  - only provides cross building and archiving
- no older Go support
- no complicated configuration and behaivors
  - convention over configuration

## usage

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
- `-build` (not implemented)
   - specify build constarints
   - os/arch と build はそれぞれは独立して動くで良い？
     - 良さそう(ダメかも)
   - buildが指定がされていて、 os/arch 設定が明にされていない場合は、デフォルトビルドはおこなわない
- `-pv` for version specification
- `-build-ldflags` / `-build-tags`
- `-o` output filename
  - not compatible with multiple package building
