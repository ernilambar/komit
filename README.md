# komit

Generate a commit subject from your staged changes with an OpenAI-compatible model, pick one, then commit and push.

Subjects follow one style: plain imperative mood, single line, at most 72 characters, no trailing period, no Conventional Commit prefix, no body.

Requires macOS 15 (Sequoia) or later on Apple Silicon, [git](https://git-scm.com) on `PATH`, and an OpenAI-compatible chat-completions endpoint.

## Install

```sh
brew install ernilambar/tap/komit
```

Prebuilt binaries and installing from source: [docs/install.md](docs/install.md).

## Usage

```sh
komit
```

Stage your changes, then run `komit`. It sends the staged diff to the model, shows the candidates, commits the one you pick, and pushes the branch. The interactive menu needs a terminal; in scripts or with piped stdin, use `-y` or `--dry-run`.

Configuration, flags, examples, status checks, exit codes, and what gets sent: [docs/usage.md](docs/usage.md).

## License

[MIT](LICENSE) © [Nilambar Sharma](https://www.nilambar.net)
