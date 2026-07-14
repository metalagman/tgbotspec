# tgbotspec

[![lint](https://github.com/metalagman/tgbotspec/actions/workflows/lint.yml/badge.svg)](https://github.com/metalagman/tgbotspec/actions/workflows/lint.yml)
[![test](https://github.com/metalagman/tgbotspec/actions/workflows/test.yml/badge.svg)](https://github.com/metalagman/tgbotspec/actions/workflows/test.yml)
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

## Examples

- [TgBotKit Client OpenAPI (latest)](https://github.com/tgbotkit/client/releases/latest/download/openapi.yaml) (published by the client project)
- [TgBotKit Client](https://github.com/tgbotkit/client) (golang, generated with [oapi-codegen](https://github.com/oapi-codegen/oapi-codegen) from this spec)


## Run in Docker

You can run the tool in a container and capture the generated OpenAPI to a file.

- Using the published image from GHCR:

```bash
docker run --rm ghcr.io/metalagman/tgbotspec:latest > openapi.yaml
```

## Run from binary

Download the latest release for your platform from
[GitHub Releases](https://github.com/metalagman/tgbotspec/releases) and run:

```bash
./tgbotspec -o openapi.yaml
```

## Build and run locally

All commands require Go 1.26.4+ when building locally.

- (Recommended) Add the CLI as a tool to your project and run it:

```bash
go get -tool github.com/metalagman/tgbotspec/cmd/tgbotspec@latest
go tool tgbotspec -o openapi.yaml
```

- Install the CLI into your GOPATH/bin and run it:

```bash
go install github.com/metalagman/tgbotspec/cmd/tgbotspec@latest
tgbotspec -o openapi.yaml
```

- Or build a local binary from this repo and run it:

```bash
go build -o tgbotspec ./cmd/tgbotspec
./tgbotspec -o openapi.yaml
```

## Multi-platform Distribution (Omnidist)

This repository includes Omnidist configuration in `.omnidist/omnidist.yaml` and
a generated GitHub Actions workflow at
`.github/workflows/omnidist-release.yml`.

Local Omnidist commands (using npm package `@omnidist/omnidist@latest`):

```bash
task omnidist:build
task omnidist:stage
task omnidist:verify
```

One-time Omnidist bootstrap in this repo:

```bash
task omnidist:init
task omnidist:ci
```

For tag-based release publishing in GitHub Actions, set repository secrets:
`NPM_PUBLISH_TOKEN` and `UV_PUBLISH_TOKEN`.


## Links

- Telegram Bot API: https://core.telegram.org/bots/api
- OpenAPI Specification: https://learn.openapis.org

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
