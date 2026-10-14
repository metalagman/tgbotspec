# Bot API refresh and release runbook

Use this runbook when Telegram updates the [official Bot API reference](https://core.telegram.org/bots/api). The generator reads that page, caches its HTML in `spec_cache.html`, and writes an OpenAPI 3.0 document. The cache, `openapi.generated.yaml`, and `coverage.out` are local artifacts; do not commit them.

If the repository still uses `master`, follow [Migrate from master to main once](#migrate-from-master-to-main-once) before the routine release steps below.

## Prepare the release branch

Use Go 1.26.4 or newer, Task, Docker, Git, and an authenticated GitHub CLI. Confirm that you can push to this repository and change its default branch if this is the first release from `main`. A maintainer must review the pull request, and CI and Codecov must pass before tagging.

```bash
git fetch origin main --tags
git switch main
git pull --ff-only origin main
git switch -c release/vX.Y.Z
```

Choose the next unused patch version by checking `git tag --sort=-version:refname` and `gh release list`. Keep release notes in the pull request description, with a summary, testing notes, risk assessment, and any migration details.

## Refresh and inspect the Bot API

Expire the existing cache without deleting it, then regenerate. Check the log for `fetcher: fetched spec`; `fetcher: using cached spec` means this run did not fetch the current page. If the network is unavailable, preserve the cache and postpone the release. Touching the cache to force offline use is useful for development, but it does not prove that a release includes Telegram's latest API.

```bash
if [ -e spec_cache.html ]; then touch -d '2 days ago' spec_cache.html; fi
task parse:save
sed -n '1,6p' openapi.generated.yaml
```

Compare `info.version` with the version at the top of Telegram's official page. Review its recent changes and check representative new methods, component schemas, and fields in `openapi.generated.yaml`. If documented elements are missing, add focused parser tests and fix the parser before proceeding. The generator reads the live page again in release CI, so recheck the published artifact's version after the tag workflow finishes.

## Validate and review

```bash
go test -coverprofile=coverage.out -covermode=atomic ./...
go tool golangci-lint run ./...
task validate
git diff --check
git status --short
git check-ignore spec_cache.html openapi.generated.yaml
```

`task validate` requires Docker and validates `openapi.generated.yaml` against the OpenAPI 3.0 schema. Review the generated document manually even when validation passes. Format changed Go files with `gofmt` and `goimports`; normalize any edited CRLF files with `dos2unix` before committing. Do not stage the ignored cache or generated spec. Push the release branch, open a pull request into `main`, request at least one maintainer review, and wait for lint, tests, Codecov, and any required checks to pass.

## Migrate from master to main once

If the remote still uses `master`, first review and merge the branch-reference and runbook changes through a pull request targeting `master`. Then create `main` at that reviewed commit. Check branch protection, rulesets, and any external integrations that name `master` before changing the default branch.

```bash
git fetch origin master --tags
git switch master
git pull --ff-only origin master
git branch -m master main
git push -u origin main
gh repo edit metalagman/tgbotspec --default-branch main
gh repo view metalagman/tgbotspec --json defaultBranchRef --jq '.defaultBranchRef.name'
git remote set-head origin -a
```

Wait for lint, tests, and Codecov on the new `main` branch. Confirm that the image workflow publishes `latest` from `main`. Retire the remote `master` branch only after branch protection and integrations have been migrated and the maintainer agrees it is no longer needed.

## Tag and check the release

After the reviewed release changes are on `main` and its required checks pass, fetch again and confirm the selected tag is still unused. Tag the exact checked commit and push that tag; the `v*` workflows build the Linux binary and `openapi.yaml`, publish the container image, and publish Omnidist npm and uv packages.

```bash
git switch main
git pull --ff-only origin main
git tag --sort=-version:refname | head
git tag -a vX.Y.Z -m 'Release vX.Y.Z'
git push origin vX.Y.Z
gh release view vX.Y.Z --json url,assets
gh run list --limit 20
```

Inspect the `release-linux`, `publish-image`, and `omnidist-release` workflow runs for that tag. Confirm the GitHub release has `openapi.yaml`, the Linux binary, and its checksum; inspect `openapi.yaml` for the reviewed Bot API version. Check the versioned GHCR image and the npm and uv/PyPI publications. npm uses trusted publishing; uv/PyPI uses the repository's `UV_PUBLISH_TOKEN` secret. A green GitHub release job does not prove that the independent package jobs succeeded.

## Recover from a failed step

- Fetch, generation, validation, test, lint, or review failure: stop before tagging, repair the release branch, and rerun the checks.
- No GitHub write or settings access: leave the reviewed branch intact and have an authorized maintainer finish the branch migration or release.
- A failed tag workflow: inspect its logs and fix the cause. A published tag and release must not be silently moved; disclose partial publication and use a new corrective release when needed.
- If the live Bot API page changes between review and release CI, review and validate the newly generated artifact before treating the release as current.
