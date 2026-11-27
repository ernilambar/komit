# Contributing

## Setup

Requires Go 1.27.3 or later (see `go.mod`).

```sh
git clone https://github.com/ernilambar/komit.git
cd komit
go build .
```

To try changes against a model, run any OpenAI-compatible server and point `KOMIT_BASE_URL` at it. Check the connection with:

```sh
go run . status
```

Use `--dry-run` while iterating so nothing gets committed:

```sh
go run . --dry-run -n 4
```

## Before opening a pull request

```sh
gofmt -l .        # should print nothing
go vet ./...
go test ./...
```

## Project layout

Single `package main`, one file per concern:

| File | Responsibility |
|---|---|
| `main.go` | Flag parsing, validation, top-level flow, exit codes |
| `config.go` | `KOMIT_*` environment variables and defaults |
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
