# Utilities for processing Wikipedia dumps in Go

[![pkg.go.dev](https://pkg.go.dev/badge/gitlab.com/tozd/go/mediawiki)](https://pkg.go.dev/gitlab.com/tozd/go/mediawiki)
[![Go Report Card](https://goreportcard.com/badge/gitlab.com/tozd/go/mediawiki)](https://goreportcard.com/report/gitlab.com/tozd/go/mediawiki)
[![pipeline status](https://gitlab.com/tozd/go/mediawiki/badges/main/pipeline.svg?ignore_skipped=true)](https://gitlab.com/tozd/go/mediawiki/-/pipelines)
[![coverage report](https://gitlab.com/tozd/go/mediawiki/badges/main/coverage.svg)](https://gitlab.com/tozd/go/mediawiki/-/graphs/main/charts)

A Go package providing utilities for processing Wikipedia dumps.

Features:

* Supports [Wikidata entities JSON dumps](https://dumps.wikimedia.org/wikidatawiki/entities/).
* Supports [Wikimedia Enterprise HTML dumps](https://dumps.wikimedia.org/other/enterprise_html/).
* Decompression and JSON decoding is parallelized for maximum throughput on a single machine.
* Parses into idiomatic Go structs.
* Can download and process a dump at the same time.
* Caches downloaded files locally.

## Installation

This is a Go package. You can add it to your project using `go get`:

```sh
go get gitlab.com/tozd/go/mediawiki
```

There is also a [read-only GitHub mirror available](https://github.com/tozd/go-errors),
if you need to fork the project there.

## Usage

See full package documentation with examples on [pkg.go.dev](https://pkg.go.dev/gitlab.com/tozd/go/mediawiki#section-documentation).
