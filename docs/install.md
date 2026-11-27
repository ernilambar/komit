# Install

komit runs on macOS 15 (Sequoia) or later on Apple Silicon (`arm64`).

## Homebrew (recommended)

```sh
brew install ernilambar/tap/komit
```

The formula lives in [ernilambar/homebrew-tap](https://github.com/ernilambar/homebrew-tap). No separate `brew tap` is needed: the fully qualified name taps it on first use.

Upgrade:

```sh
brew upgrade komit
```

Uninstall:

```sh
brew uninstall komit
brew untap ernilambar/tap   # optional
```

## Prebuilt binary

Each release ships a `komit-v<version>-darwin-arm64.tar.gz` archive on the [releases page](https://github.com/ernilambar/komit/releases). Download, extract, and move `komit` onto your `PATH`:

```sh
curl -LO https://github.com/ernilambar/komit/releases/download/v0.1.0/komit-v0.1.0-darwin-arm64.tar.gz
tar -xzf komit-v0.1.0-darwin-arm64.tar.gz
sudo mv komit /usr/local/bin/
```

Replace `v0.1.0` with the release you want.

## From source

Requires Go 1.27.3 or later:

```sh
go install github.com/ernilambar/komit@latest
```

`go install` drops the binary in `$(go env GOPATH)/bin`, which must be on your `PATH`.

## Verify

```sh
komit --version
```

Next: [configure a model](../README.md#configuration).
