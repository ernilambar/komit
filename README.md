# komit

Generate a commit subject from your staged changes with an OpenAI-compatible model, pick one, then commit and push.

Subjects follow one style: plain imperative mood, single line, at most 72 characters, no trailing period, no Conventional Commit prefix, no body.

## Install

```sh
go install github.com/ernilambar/komit@latest
```

## Configuration

Set through environment variables:

| Variable | Required | Description |
|---|---|---|
| `KOMIT_BASE_URL` | Yes | OpenAI-compatible API base URL, e.g. `http://127.0.0.1:11434/v1` |
| `KOMIT_MODEL` | Yes, unless `--model` is passed | Model name |
| `KOMIT_API_KEY` | No | Sent as `Authorization: Bearer` when set |

komit exits with an error if a required variable is missing.

## Usage

```sh
komit [flags]
komit status
```

Stage changes, then run `komit`. It sends the staged diff to the model, shows the candidates, and commits the one you pick. Then it pushes the current branch. If the branch has no upstream, it runs `git push -u origin HEAD`.

The selection menu:

- AI-generated subjects
- `↻ Regenerate`: ask the model for new candidates
- `✎ Write my own`: edit a subject, prefilled with the first suggestion. Submit empty to go back to the list.

`Esc` or `Ctrl+C` cancels without committing.

### Flags

| Flag | Description |
|---|---|
| `-a`, `--all` | Stage modified and deleted tracked files first (`git add -u`) |
| `-y`, `--yes` | Commit the first suggestion without prompting |
| `-n`, `--count <n>` | Number of candidates, 1–5 (default 2). Cannot be combined with `-y` unless `-n 1` |
| `--dry-run` | Print the candidates and exit without committing |
| `--no-push` | Commit without pushing |
| `--model <name>` | Override `KOMIT_MODEL` |

Short flags can be combined, e.g. `-ay`.

### Examples

```sh
komit                  # pick from 2 candidates, commit, push
komit -a -n 4          # stage tracked changes, pick from 4 candidates
komit -ay              # stage, commit the first suggestion, push
komit --dry-run        # preview candidates only
komit -y --no-push     # commit without prompting, skip push
```

The interactive menu needs a terminal. In scripts or with piped stdin, use `-y` or `--dry-run`.

### Status

`komit status` prints the configuration and checks the server:

```
$ komit status
Base URL   http://127.0.0.1:9090/v1
Model      gemma-4-e4b-it-4bit
API key    not set

✓ Server reachable        42ms
✓ Model available
✓ Generation works        810ms
```

Server and model problems are warnings. It exits 1 if a required variable is missing or the generation check fails.

## What gets sent

- `git diff --cached --stat`, always in full.
- The staged diff, capped at 12 KB. Changes to `*.lock`, `package-lock.json`, `go.sum` and `*.min.*` files are left out.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Error: invalid flags, missing config, nothing staged, server unreachable, commit or push failed |
| `130` | Cancelled in the interactive prompt |

If the push fails, the commit is kept.

## License

[MIT](LICENSE) © [Nilambar Sharma](https://www.nilambar.net)
