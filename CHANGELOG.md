# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Do not handle signals inside `Process`. It should be done outside of it.

## [0.2.0] - 2022-01-23

### Added

- `Amount` implements `fmt.Stringer` interface.

### Fixed

- Always format `Amount` to string precisely if possible.

## [0.1.0] - 2022-01-09

### Added

- First public release.

[Unreleased]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.2.0...main
[0.2.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.1.0...v0.2.0
[0.1.0]: https://gitlab.com/tozd/go/mediawiki/-/tags/v0.1.0

<!-- markdownlint-disable-file MD024 -->
