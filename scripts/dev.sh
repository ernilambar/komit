#!/usr/bin/env bash
# Local development helper. Run from anywhere: scripts/dev.sh <command> [args]
set -euo pipefail

cd "$(dirname "$0")/.."

if [[ "$(uname -s)" != "Darwin" || "$(uname -m)" != "arm64" ]]; then
	echo "komit requires macOS 15 or later on Apple Silicon" >&2
	exit 1
fi

# Load KOMIT_* settings from .env (gitignored) if present.
if [[ -f .env ]]; then
	set -a
	# shellcheck disable=SC1091
	source .env
	set +a
fi

version="$(git describe --tags --always --dirty 2>/dev/null || echo dev)"
ldflags="-X main.version=${version} -macos 15.0"

usage() {
	cat <<EOF
Usage: scripts/dev.sh <command> [args]

Commands:
  build          Build ./komit with version ${version}
  run [args]     Run from source in dry-run mode (e.g. run -n 4)
  status         Run 'komit status' from source
  check          gofmt, go vet, go test (same as CI)
  install        Install to \$(go env GOPATH)/bin
EOF
}

cmd="${1:-}"
[[ $# -gt 0 ]] && shift

case "$cmd" in
build)
	go build -trimpath -ldflags="$ldflags" -o komit .
	./komit --version
	;;
run)
	go run -ldflags="$ldflags" . --dry-run "$@"
	;;
status)
	go run -ldflags="$ldflags" . status
	;;
check)
	unformatted="$(gofmt -l .)"
	if [[ -n "$unformatted" ]]; then
		echo "Run gofmt on:"
		echo "$unformatted"
		exit 1
	fi
	go vet ./...
	go test -race -cover ./...
	;;
install)
	go install -trimpath -ldflags="$ldflags" .
	;;
*)
	usage
	[[ -z "$cmd" || "$cmd" == "-h" || "$cmd" == "--help" ]] || exit 1
	;;
esac
