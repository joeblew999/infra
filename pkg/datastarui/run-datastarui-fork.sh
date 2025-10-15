#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "$SCRIPT_DIR/../.." && pwd)"
FORK_DIR="$ROOT_DIR/.src/datastarui/fork/datastarui"

if [[ ! -d "$FORK_DIR" ]]; then
  echo "DatastarUI fork not found at $FORK_DIR" >&2
  exit 1
fi

if [[ $# -lt 1 ]]; then
  cat <<USAGE >&2
Usage: $(basename "$0") <codegen|playwright|test|serve-test> [args...]
  codegen       Run the fork's code generation helper (passes remaining args).
  playwright    Run the fork's Playwright helper (passes remaining args).
  test          Run Go tests in the fork (default ./...).
  serve-test    Run Go tests, then launch the fork's server for manual inspection.
USAGE
  exit 1
fi

command="$1"
shift

case "$command" in
  codegen)
    (cd "$FORK_DIR" && GOWORK=off go run ./cmd/codegen "$@")
    ;;
  playwright)
    (cd "$FORK_DIR" && GOWORK=off go run ./cmd/playwright "$@")
    ;;
  test)
    if [[ $# -eq 0 ]]; then
      set -- ./...
    fi
    (cd "$FORK_DIR" && GOWORK=off go test "$@")
    ;;
  serve-test)
    if [[ $# -eq 0 ]]; then
      set -- ./...
    fi
    (cd "$FORK_DIR" && GOWORK=off go test "$@")
    echo
    echo "Tests passed. Starting DatastarUI fork server (Ctrl+C to stop)"
    cd "$FORK_DIR"
    exec env GOWORK=off go run .
    ;;
  *)
    echo "Unknown command: $command" >&2
    exit 1
    ;;
esac
