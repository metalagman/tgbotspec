# AGENTS Guidelines for This Repository

This project scrapes the official Telegram Bot API reference and renders an OpenAPI
3.0 spec via `./cmd/tgbotspec`. The entrypoints live under `internal/` with the
OpenAPI template in `internal/openapi/openapi.yaml.gotmpl`.

## Branching & Commits
- Branch from `master` for new work.
- Keep commits focused on one logical change and use present-tense, imperative
  messages (e.g. `Add coverage for scraper`).
- Reference relevant issue IDs in commits or PRs when applicable (e.g. `Fix #42`).

## Preferred Workflow
- Use Go **1.24+**. When running the generator inside sandboxes, set
  `GOCACHE=$(pwd)/.gocache go run ./cmd/tgbotspec` to avoid permission issues with
  the default build cache.
- A `Taskfile.yml` is available: `task`, `task parse`, and `task parse:save` all run
  the generator; the latter writes to `openapi.generated.yaml`.
- The scraper relies on `spec_cache.html` to work offline. If the network is blocked,
  `touch spec_cache.html` before running the generator so the cached copy is used.

## Code Style
- Go sources must compile with the pinned `go1.24` toolchain.
- Run `gofmt`, `goimports`, and any project formatters before committing.
- Respect `.golangci.yml`; do not disable linters locally to achieve a clean run.
- Favor explicit, self-documenting code. Add concise comments only when logic is
  non-obvious or tied to Telegram documentation quirks.

## Linting & Tests
- Lint locally with:
  ```bash
  GOCACHE=$(pwd)/.cache/go-build golangci-lint run ./...
  ```
- Run the full test suite with coverage before pushing:
  ```bash
  mkdir -p .cache/go-build
  GOCACHE=$(pwd)/.cache/go-build go test -coverprofile=coverage.out -covermode=atomic ./...
  ```
- Maintain or improve coverage; avoid introducing untested files unless impossible,
  and call it out in the PR.
- `openapi.generated.yaml` is intentionally git-ignored; regenerate it for manual
  verification but do not commit it.

## Pull Requests & Documentation
- Include a concise summary, testing notes, and risk assessment in PRs.
- Document any migrations (schema updates, cache invalidations, etc.) and request
  review from at least one maintainer.
- Update `README.md` or other docs whenever behavior changes so the docs match the
  code.
- Keep changelog notes inside the PR description; they are harvested for releases.

## CI & Releases
- Ensure GitHub Actions (lint + Codecov) stay green; fix regressions rather than
  relaxing thresholds.
- Add new workflows only after verifying the required tokens or secrets exist.
- Maintainers tag releases from `master` once CI and Codecov pass, and should run
  `task parse:save` when publishing.

## File Conventions
- Repository files now standardize on LF line endings. If you spot a CRLF file,
  run `dos2unix <file>` before committing.
- Keep changes focused—avoid reverting unrelated user edits or deleting cache files.
