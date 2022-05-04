# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Changed

- Utility functions to determine the latest dump's URL accept `context.Context` parameter.

## [0.9.0] - 2022-04-17

### Changed

- Implementation now uses Go generics and Go 1.18 or newer is now required.
  `ProcessConfig` has no more `Item` field and instead has a type parameter.

## [0.8.1] - 2022-03-01

### Fixed

- All string data is normalized to Unicode NFC.

## [0.8.0] - 2022-02-17

### Changed

- `MainEntity` of `Article` struct is a pointer now because it might be missing.

## [0.7.0] - 2022-02-16

### Changed

- Stale download timeout has been removed because it can lead to false positives
  when processing is slower than downloading.

## [0.6.0] - 2022-02-16

### Added

- Utility functions to determine the latest dump's URL:
  `LatestCommonsEntitiesRun`, `LatestCommonsImageMetadataRun`,
  `LatestWikidataEntitiesRun`, `LatestWikipediaImageMetadataRun`,
  `LatestWikipediaRun`.

### Changed

- High-level function do not anymore automatically determine the latest dump's URL.
  This logic is now in separate utility functions.
- Low-level `Process` function and high-level functions do not deal with cache location
  anymore but accept `URL` and `Path` arguments which directly control what is downloaded
  and where it is stored (or used from if already stored).

## [0.5.0] - 2022-02-04

### Added

- Support for Wikimedia Commons entities dumps.
- Support for SQL dumps.

### Changed

- `JSONDecodeThreads` field in `ProcessDumpConfig` has been renamed to `DecodingThreads`.
  Similarly in `ProcessWikipediaDump` and `ProcessWikidataDump`.
- `EntityType` enumeration has been extended with `MediaInfo`.
- `FileType` enumeration has been extended with `SQLDump`.

## [0.4.0] - 2022-01-26

### Changed

- Remove `UserAgent` parameter. Provided HTTP client should set it instead.

## [0.3.0] - 2022-01-24

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

[unreleased]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.9.0...main
[0.9.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.8.1...v0.9.0
[0.8.1]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.8.0...v0.8.1
[0.8.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.7.0...v0.8.0
[0.7.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.6.0...v0.7.0
[0.6.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.5.0...v0.6.0
[0.5.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.4.0...v0.5.0
[0.4.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.3.0...v0.4.0
[0.3.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.2.0...v0.3.0
[0.2.0]: https://gitlab.com/tozd/go/mediawiki/-/compare/v0.1.0...v0.2.0
[0.1.0]: https://gitlab.com/tozd/go/mediawiki/-/tags/v0.1.0

<!-- markdownlint-disable-file MD024 -->
