# Changelog

## [v0.10.1](https://github.com/Songmu/goxz/compare/v0.10.0...v0.10.1) - 2025-06-18
- drop mholt/archiver dependency by @Songmu in https://github.com/Songmu/goxz/pull/47
- drop pkg/errors dependency by @Songmu in https://github.com/Songmu/goxz/pull/48

## [v0.10.0](https://github.com/Songmu/goxz/compare/v0.9.1...v0.10.0) - 2025-06-18
- replace ioutil by @yulog in https://github.com/Songmu/goxz/pull/39
- docs: add the installation guide with aqua by @suzuki-shunsuke in https://github.com/Songmu/goxz/pull/38
- update Go and dependencies to latest by @Songmu in https://github.com/Songmu/goxz/pull/41
- introduce tagpr by @Songmu in https://github.com/Songmu/goxz/pull/42

## [v0.9.1](https://github.com/Songmu/goxz/compare/v0.9.0...v0.9.1) (2022-08-18)

* refine resource gathering [#36](https://github.com/Songmu/goxz/pull/36) ([Songmu](https://github.com/Songmu))

## [v0.9.0](https://github.com/Songmu/goxz/compare/v0.8.2...v0.9.0) (2022-08-18)

* update deps [#35](https://github.com/Songmu/goxz/pull/35) ([Songmu](https://github.com/Songmu))
* skip executables from auto collected resources [#34](https://github.com/Songmu/goxz/pull/34) ([Songmu](https://github.com/Songmu))

## [v0.8.2](https://github.com/Songmu/goxz/compare/v0.8.1...v0.8.2) (2022-05-06)

* Use deflate algorithm for zip archives [#33](https://github.com/Songmu/goxz/pull/33) ([itchyny](https://github.com/itchyny))

## [v0.8.1](https://github.com/Songmu/goxz/compare/v0.8.0...v0.8.1) (2021-12-26)

* update mholt/archiver/v3 [#31](https://github.com/Songmu/goxz/pull/31) ([Songmu](https://github.com/Songmu))

## [v0.8.0](https://github.com/Songmu/goxz/compare/v0.7.0...v0.8.0) (2021-12-26)

* build arm64 by default in addition to amd64 [#30](https://github.com/Songmu/goxz/pull/30) ([Songmu](https://github.com/Songmu))
* go1.17 and update deps [#29](https://github.com/Songmu/goxz/pull/29) ([Songmu](https://github.com/Songmu))
* fix synopsis text [#28](https://github.com/Songmu/goxz/pull/28) ([yulog](https://github.com/yulog))

## [v0.7.0](https://github.com/Songmu/goxz/compare/v0.6.0...v0.7.0) (2021-04-05)

* make -trimpath true by default [#27](https://github.com/Songmu/goxz/pull/27) ([Songmu](https://github.com/Songmu))
* Add -trimpath option [#24](https://github.com/Songmu/goxz/pull/24) ([hirose31](https://github.com/hirose31))
* migrate CIs to GitHub Actions [#25](https://github.com/Songmu/goxz/pull/25) ([Songmu](https://github.com/Songmu))
* Do not use -Hwindowsgui [#21](https://github.com/Songmu/goxz/pull/21) ([mattn](https://github.com/mattn))

## [v0.6.0](https://github.com/Songmu/goxz/compare/v0.5.0...v0.6.0) (2020-01-17)

* implement -static flag to build static binary [#20](https://github.com/Songmu/goxz/pull/20) ([Songmu](https://github.com/Songmu))

## [v0.5.0](https://github.com/Songmu/goxz/compare/v0.4.1...v0.5.0) (2019-11-19)

* change interface of func Run [#19](https://github.com/Songmu/goxz/pull/19) ([Songmu](https://github.com/Songmu))
* add -build-installsuffix option to support -installsuffix [#17](https://github.com/Songmu/goxz/pull/17) ([Songmu](https://github.com/Songmu))

## [v0.4.1](https://github.com/Songmu/goxz/compare/v0.4.0...v0.4.1) (2019-05-01)

* [bugfix] fix output path in windows [#16](https://github.com/Songmu/goxz/pull/16) ([Songmu](https://github.com/Songmu))

## [v0.4.0](https://github.com/Songmu/goxz/compare/v0.3.3...v0.4.0) (2019-04-27)

* [incompatible] keep directory structure for included resources [#15](https://github.com/Songmu/goxz/pull/15) ([Songmu](https://github.com/Songmu))

## [v0.3.3](https://github.com/Songmu/goxz/compare/v0.3.2...v0.3.3) (2019-04-08)

* Capture stdout and stderr of go list separately [#14](https://github.com/Songmu/goxz/pull/14) ([itchyny](https://github.com/itchyny))

## [v0.3.2](https://github.com/Songmu/goxz/compare/v0.3.1...v0.3.2) (2019-04-03)

* Make sure to create the destination directory [#13](https://github.com/Songmu/goxz/pull/13) ([Songmu](https://github.com/Songmu))

## [v0.3.1](https://github.com/Songmu/goxz/compare/v0.3.0...v0.3.1) (2019-04-02)

* Fix -C option behavior [#12](https://github.com/Songmu/goxz/pull/12) ([Songmu](https://github.com/Songmu))

## [v0.3.0](https://github.com/Songmu/goxz/compare/v0.2.0...v0.3.0) (2019-03-29)

* update deps [#10](https://github.com/Songmu/goxz/pull/10) ([Songmu](https://github.com/Songmu))
* To work outside GOPATH [#9](https://github.com/Songmu/goxz/pull/9) ([Songmu](https://github.com/Songmu))

## [v0.2.0](https://github.com/Songmu/goxz/compare/v0.1.1...v0.2.0) (2019-02-18)

* introduce go modules [#7](https://github.com/Songmu/goxz/pull/7) ([Songmu](https://github.com/Songmu))

## [v0.1.1](https://github.com/Songmu/goxz/compare/v0.1.0...v0.1.1) (2018-11-09)

* use github.com/mholt/archiver v3 [#6](https://github.com/Songmu/goxz/pull/6) ([astj](https://github.com/astj))

## [v0.1.0](https://github.com/Songmu/goxz/compare/v0.0.2...v0.1.0) (2017-12-31)

* add -Include option for additional resources inclusion [#3](https://github.com/Songmu/goxz/pull/3) ([Songmu](https://github.com/Songmu))

## [v0.0.2](https://github.com/Songmu/goxz/compare/v0.0.1...v0.0.2) (2017-12-28)

* output usage to stdout and normally exit with `-h` option [#2](https://github.com/Songmu/goxz/pull/2) ([Songmu](https://github.com/Songmu))

## [v0.0.1](https://github.com/Songmu/goxz/compare/3fde63a0...v0.0.1) (2017-12-26)

* Initial implement [#1](https://github.com/Songmu/goxz/pull/1) ([Songmu](https://github.com/Songmu))
