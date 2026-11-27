# Contributing

## Setup

Requires macOS 15 (Sequoia) or later on Apple Silicon (`arm64`) and Go 1.27.3 or later (see `go.mod`).

```sh
git clone https://github.com/ernilambar/komit.git
cd komit
scripts/dev.sh build
```

To try changes against a model, run any OpenAI-compatible server and set `KOMIT_BASE_URL` and `KOMIT_MODEL` for it, either exported or in a `.env` file (gitignored, loaded by `scripts/dev.sh`). Check the connection with:

```sh
scripts/dev.sh status
```

`scripts/dev.sh run` always passes `--dry-run`, so nothing gets committed while iterating:

```sh
scripts/dev.sh run -n 4
```

Run `scripts/dev.sh` with no arguments to list all commands.

## Before opening a pull request

```sh
scripts/dev.sh check   # gofmt, go vet, go test -race (same as CI)
```

## Project layout

Single `package main`, one file per concern:

| File | Responsibility |
|---|---|
| `main.go` | Flag parsing, validation, top-level flow, exit codes |
| `config.go` | `KOMIT_*` environment variables and validation |
| `git.go` | Staging, staged diff, commit, push |
| `ai.go` | Chat-completions client and system prompt |
| `clean.go` | Subject cleanup rules (tested in `clean_test.go`) |
| `prompt.go` | Interactive selection |
| `status.go` | `komit status` checks |

## Guidelines

- Keep dependencies minimal. Use the standard library unless a dependency clearly earns its place.
- Add or update table cases in `clean_test.go` when changing cleanup rules.
- Keep changes focused. One concern per pull request.

## Commit messages

Use the same style komit produces:

- Plain imperative mood: `Add endpoint validation`, `Fix nil pointer in parser`.
- Single line, at most 72 characters, no trailing period.
- No Conventional Commit prefixes (`feat:`, `fix:`, …) and no body.
