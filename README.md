# tgbotspec

[![Go Report Card](https://goreportcard.com/badge/github.com/metalagman/tgbotspec)](https://goreportcard.com/report/github.com/metalagman/tgbotspec)
[![build](https://github.com/metalagman/tgbotspec/actions/workflows/golangci-lint.yml/badge.svg)](https://github.com/metalagman/tgbotspec/actions/workflows/golangci-lint.yml)
[![tests](https://github.com/metalagman/tgbotspec/actions/workflows/codecov.yml/badge.svg?label=tests)](https://github.com/metalagman/tgbotspec/actions/workflows/codecov.yml)
[![codecov](https://codecov.io/github/metalagman/tgbotspec/graph/badge.svg?token=LRNA4STCO7)](https://codecov.io/github/metalagman/tgbotspec)
[![version](https://img.shields.io/github/v/release/metalagman/tgbotspec?sort=semver)](https://github.com/metalagman/tgbotspec/releases)
[![license](https://img.shields.io/github/license/metalagman/tgbotspec)](LICENSE)

**tgbotspec** turns the official Telegram Bot API reference into an OpenAPI 3.0
specification. Use the generated file with SDK generators, API explorers, or
your own tooling.

## Features

- Methods: all Telegram Bot API methods are included as OpenAPI paths with proper HTTP verbs and parameters.
- Objects: all Bot API objects are generated as reusable component schemas.
- Any‑of/one‑of types: union types from the docs are modeled with OpenAPI `anyOf`/`oneOf` (and refs) so generators can produce correct sum types.
- Authorization: bearer token (`TelegramBotToken`) with server URL `https://api.telegram.org/bot{botToken}`.

## Run in Docker

You can run the tool in a container and capture the generated OpenAPI to a file.

- Using the published image from GHCR:

```bash
docker run --rm ghcr.io/metalagman/tgbotspec:latest > openapi.generated.yaml
```

## Run from binary

Download the latest release for your platform from
[GitHub Releases](https://github.com/metalagman/tgbotspec/releases) and run:

```bash
./tgbotspec > openapi.generated.yaml
```

## Build and run locally

All commands require Go 1.24+ when building locally.

- Install the CLI into your GOPATH/bin and run it:

```bash
go install github.com/metalagman/tgbotspec/cmd/tgbotspec@latest
tgbotspec > openapi.generated.yaml
```

- Or build a local binary from this repo and run it:

```bash
go build -o tgbotspec ./cmd/tgbotspec
./tgbotspec > openapi.generated.yaml
```


## Links

- Telegram Bot API: https://core.telegram.org/bots/api
- OpenAPI Specification: https://learn.openapis.org

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

