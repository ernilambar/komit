# AGENTS.md

Guidance for AI coding agents working in this repository.

## Overview

komit is a macOS CLI that sends the staged git diff to an OpenAI-compatible model, lets the
user pick a generated commit subject, then commits and pushes. Supported platform: macOS 15
(Sequoia) or later on Apple Silicon (`arm64`). Single `package main`; UI uses
`charm.land/huh` and `bubbles`, flags use `github.com/spf13/pflag`.

## Setup

Requires macOS 15 (Sequoia) or later on Apple Silicon (`arm64`), Go 1.27.3 or later (see
`go.mod`), and `git` on `PATH`.

```sh
go mod download
export KOMIT_BASE_URL=http://127.0.0.1:11434/v1
export KOMIT_MODEL=gemma3:4b
export KOMIT_API_KEY=            # optional; sent as Authorization: Bearer
export KOMIT_REASONING_EFFORT=   # optional; sent as reasoning_effort, e.g. none
```

`scripts/dev.sh` also loads `KOMIT_*` variables from a gitignored `.env` file. No build
step is required before running; use `scripts/dev.sh build` for a versioned binary.

## Commands

```sh
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w -macos 15.0" -o komit .
go test -race -cover ./...
gofmt -l .
go vet ./...
golangci-lint run
go mod tidy
```

There is no separate typecheck; `go vet ./...` and `go build ./...` cover compilation and
static analysis.

## Conventions

- **One file per concern, single `package main`.** Keep the existing split: `main.go`
  (flags, flow, exit codes), `config.go` (env), `git.go` (git operations), `ai.go`
  (chat-completions client), `clean.go` (subject cleanup), `prompt.go` (interactive
  selection), `status.go` (`komit status`). Add new behavior to the matching file rather
  than creating a new package.
- **Standard library first.** Do not add a dependency unless it clearly earns its place
  (see `CONTRIBUTING.md`). Keep `go.mod` tidy and minimal.
- **Subject rules live only in `clean.go`.** Model output goes through
  `cleanSubjects`/`cleanLine` (prefixes stripped); user-typed subjects go through
  `cleanInput` (prefixes rejected): plain imperative, one line, ≤72 characters, no
  trailing period, no Conventional Commit prefix, no body. When changing these rules, add
  or update table cases in `clean_test.go`. Repo commit messages use the same style, e.g.
  `Add scripts`.
- **Target macOS 15+ on Apple Silicon.** Keep the guard at the top of `run()`, and keep
  `GOARCH=arm64` with `-macos 15.0` in release builds.
- **Exit codes are a contract.** Return `int` from `run()` and use exit codes `0` success,
  `1` error, `130` cancelled (see README). Never call `os.Exit` outside `main`. A failed
  push must keep the commit.
- **Changelog.** `CHANGELOG.md` is maintained with
  [`@nilambar/flatlog`](https://github.com/ernilambar/flatlog). Record user-facing changes
  with `npx @nilambar/flatlog bullet "Added: ..."` (allowed prefixes: `Added:`, `Changed:`,
  `Fixed:`), promote the unreleased block with `npx @nilambar/flatlog release <version>`
  before tagging, and run `npx @nilambar/flatlog validate <version>` to verify the result.

## Releases

1. Add pending entries, then promote them: `npx @nilambar/flatlog release <version>`.
2. Validate the changelog: `npx @nilambar/flatlog validate <version>`. Validation fails
   while an unreleased (`X.X.X`) block exists, so run it only after promoting the release.

## Quality gate

Run before declaring any task complete; every command must exit 0.

```sh
gofmt -l .            # must print nothing
go vet ./...
go test -race -cover ./...
GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w -macos 15.0" -o /tmp/komit .
/tmp/komit --version
```

CI lints on Ubuntu (static analysis only) and runs tests and the build smoke test on
macOS 15 (Apple Silicon). Run `scripts/dev.sh check` for the gofmt/vet/test subset locally.
