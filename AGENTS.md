# AGENTS Guidelines for This Repository

This project scrapes the official Telegram Bot API reference and renders an OpenAPI
3.0 spec via `./cmd/tgbotspec`. The entrypoints live under `internal/` with the
OpenAPI template in `internal/openapi/openapi.yaml.gotmpl`.

## Preferred Workflow
- Use Go **1.24+**. When running the generator inside sandboxes, set
  `GOCACHE=$(pwd)/.gocache go run ./cmd/tgbotspec` to avoid permission issues with
  the default build cache.
- A `Taskfile.yml` is available: `task`, `task parse`, and `task parse:save` all run
  the generator; the latter writes to `openapi.generated.yaml`.
- The scraper relies on `spec_cache.html` to work offline. If the network is blocked,
  `touch spec_cache.html` before running the generator so the cached copy is used.

## Validation
- Run `go test ./...` before finishing significant work.
- `golangci-lint run ./...` is the expected lint command (installed separately).
- `openapi.generated.yaml` is intentionally git-ignored; regenerate it for manual
  verification but do not commit it.

## File Conventions
- `internal/openapi/openapi.yaml.gotmpl` currently uses CRLF line endings; preserve
  them when editing to keep diffs minimal.
- Keep changes focused—avoid reverting unrelated user edits or deleting cache files.
