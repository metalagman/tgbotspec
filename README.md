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

## Run with Docker

You can run the tool in a container and capture the generated OpenAPI to a file.

- Using the published image from GHCR:

```bash
docker run --rm ghcr.io/metalagman/tgbotspec:latest > openapi.generated.yaml
```

- Or using a locally built image (from this repo):

```bash
# Build the image
docker build -t tgbotspec:local .

# Run and save output
docker run --rm tgbotspec:local > openapi.generated.yaml
```


## Generate the Spec

```bash
tgbotspec > openapi.generated.yaml
```

The tool prints the OpenAPI document to standard output so you can redirect it
to a file or pipe it into another process.

By default the scraper works online.
