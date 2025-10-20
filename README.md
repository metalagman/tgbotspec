# tgbotspec

tgbotspec scrapes the official Telegram Bot API reference and emits an
OpenAPI 3.0 specification. The generated document is suitable for feeding
into code generators, client SDKs, or documentation tooling.

## Quick start

```bash
go run ./cmd/tgbotspec > openapi.generated.yaml
```

The command prints the complete OpenAPI spec to standard output. Redirect it
to a file (as shown above) or pipe it into other tooling as needed.

## Tasks

A Taskfile is provided for common workflows:

```bash
task          # runs the scraper and prints the spec
task parse    # same as the default task
task parse:save
```

The `parse:save` task writes the generated spec to `openapi.generated.yaml`.

## Development

Run the test suite to validate changes:

```bash
go test ./...
```

The project uses [Cobra](https://github.com/spf13/cobra) for the CLI entrypoint
and relies on Go 1.24 or newer.
