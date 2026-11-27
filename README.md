# komit

Generate a commit subject from your staged changes with an OpenAI-compatible model, pick one, then commit and push.

Subjects follow one style: plain imperative mood, single line, at most 72 characters, no trailing period, no Conventional Commit prefix, no body.

Requires macOS 15 (Sequoia) or later on Apple Silicon, [git](https://git-scm.com) on `PATH`, and an OpenAI-compatible chat-completions endpoint.

## Install

```sh
brew install ernilambar/tap/komit
```

Prebuilt binaries and installing from source: [docs/install.md](docs/install.md).

## Configuration

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

Stage changes, then run `komit`. It sends the staged diff to the model, shows the candidates, commits the one you pick, and pushes the branch (`git push -u origin HEAD` if it has no upstream).

The menu offers the generated subjects, `↻ Regenerate`, and `✎ Write my own` (submit empty to go back). Press `e` to edit the highlighted subject, or `r` to regenerate. `Esc` or `Ctrl+C` cancels without committing.

### Flags

| Flag | Description |
|---|---|
| `-a`, `--all` | Stage all changes first, including untracked files (`git add -A`) |
| `-y`, `--yes` | Commit the first suggestion without prompting |
| `-n`, `--count <n>` | Number of candidates, 1–5 (default 2). Cannot be combined with `-y` unless `-n 1` |
| `--dry-run` | Print the candidates and exit without committing |
| `--no-push` | Commit without pushing |
| `--model <name>` | Override `KOMIT_MODEL` |
| `-V`, `--version` | Print version and exit |
| `-h`, `--help` | Print help and exit |

Short flags can be combined, e.g. `-ay`.

### Examples

```sh
komit                  # pick from 2 candidates, commit, push
komit -a -n 4          # stage all changes, pick from 4 candidates
komit -ay              # stage, commit the first suggestion, push
komit --dry-run        # preview candidates only
```

The interactive menu needs a terminal. In scripts or with piped stdin, use `-y` or `--dry-run`.

### Status

`komit status` prints the configuration and checks the server:

```
$ komit status
Base URL   http://127.0.0.1:11434/v1
Model      gemma3:4b
API key    not set

✓ Server reachable        42ms
✓ Model available
✓ Generation works        810ms
```

Server and model problems are warnings; missing required config or a failed generation check exits 1.

## What gets sent

- `git diff --cached --stat`, always in full.
- The staged diff, capped at 12 KB. Changes to `*.lock`, `package-lock.json`, `go.sum` and `*.min.*` files are left out.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Error: unsupported platform, invalid flags, missing config, nothing staged, server unreachable, commit or push failed |
| `130` | Cancelled in the interactive prompt |

If the push fails, the commit is kept.

## License

[MIT](LICENSE) © [Nilambar Sharma](https://www.nilambar.net)
