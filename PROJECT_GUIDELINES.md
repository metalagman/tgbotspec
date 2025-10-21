# Project Guidelines

These guidelines capture the expectations for contributing to the `tgbotspec`
codebase. Follow them for every change unless the maintainers agree on an
exception.

---

## 1. Branching & Commits
- Create feature branches off `master`.
- Keep commits focused on one logical change; split large work into reviewable
  chunks.
- Use present‑tense, imperative commit messages (e.g. `Add coverage for scraper`).
- Reference issue IDs in commits/PRs when relevant (e.g. `Fix #42`).

## 2. Code Style
- Go source must compile with `go1.24` (toolchain pinned in `go.mod`).
- Run `gofmt`, `goimports`, and project formatters before committing.
- Respect the linters configured in `.golangci.yml`; **never** disable them
  locally to get a clean run.
- Prefer explicit, self‑documenting code. Add brief comments only when logic
  is non‑obvious or relies on Telegram doc quirks.

## 3. Linting & Tests
- Lint locally with:
  ```bash
  GOCACHE=$(pwd)/.cache/go-build golangci-lint run ./...
  ```
- Run the full test suite (with coverage) before pushing:
  ```bash
  mkdir -p .cache/go-build
  GOCACHE=$(pwd)/.cache/go-build go test -coverprofile=coverage.out -covermode=atomic ./...
  ```
- Keep or improve package coverage. Do not introduce new files without tests
  unless impossible (explain in PR).

## 4. GitHub Actions
- CI must be green (lint + Codecov workflows).
- For coverage regressions, extend tests rather than relaxing thresholds.
- Add new workflows only after confirming the project token/secret needs are
  satisfied.

## 5. Pull Requests
- Include a short summary, testing notes, and risk assessment.
- Document any migration steps (schema updates, cache invalidations, etc.).
- Request review from at least one maintainer.

## 6. Documentation
- Update `README.md` or other docs when behavior changes.
- Keep changelog notes inside PR descriptions; they are harvested for releases.

## 7. Release Process (Maintainers)
- Tag releases from `master` after CI + Codecov pass.
- Regenerate OpenAPI artifacts with `task parse:save` when publishing.

---

A lightweight checklist for every change:
1. Lint.
2. Test with coverage.
3. Ensure docs reflect code.
4. Push branch, open PR, monitor CI & Codecov.
