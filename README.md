# tgbotspec

[![build](https://github.com/metalagman/tgbotspec/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/metalagman/tgbotspec/actions/workflows/golangci-lint.yml)
[![codecov](https://codecov.io/github/metalagman/tgbotspec/graph/badge.svg?token=LRNA4STCO7)](https://codecov.io/github/metalagman/tgbotspec)

tgbotspec turns the official Telegram Bot API reference into an OpenAPI 3.0
specification. Use the generated file with SDK generators, API explorers, or
your own tooling.

## Install

Choose the option that suits your workflow:

- **Go users:** `go install github.com/metalagman/tgbotspec/cmd/tgbotspec@latest`
- **Binary download:** grab the latest tagged release for Linux from the
  [GitHub releases](https://github.com/metalagman/tgbotspec/releases) page.
- **Run without installing:** `go run ./cmd/tgbotspec`

All commands require Go 1.24+ when building locally.

## Generate the spec

```bash
tgbotspec > openapi.generated.yaml
```

The tool prints the OpenAPI document to standard output so you can redirect it
to a file or pipe it into another process.

By default the scraper works online. If you need to run it offline, place a
cached copy of the Telegram docs in `spec_cache.html` before executing the
command.
