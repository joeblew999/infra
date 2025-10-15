#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
APP_DIR="$SCRIPT_DIR/sampleapp"

if [[ ! -d "$APP_DIR" ]]; then
  echo "Sample app not found at $APP_DIR" >&2
  exit 1
fi

if [[ $# -lt 1 ]]; then
  cat <<USAGE >&2
Usage: $(basename "$0") <codegen|playwright|test|serve-test> [args...]
  codegen       Run code generation against the sample app (--src injected).
  playwright    Run the Playwright helper against the sample app (--src injected).
  test          Run Go tests for the sample app (default ./sampleapp).
  serve-test    Run Go tests, then launch the sample app server for manual inspection.
USAGE
  exit 1
fi

command="$1"
shift

case "$command" in
  codegen)
    (cd "$SCRIPT_DIR" && GOWORK=off go run ./cmd/codegen --src "$APP_DIR" "$@")
    ;;
  playwright)
    (cd "$SCRIPT_DIR" && GOWORK=off go run ./cmd/playwright --src "$APP_DIR" "$@")
    ;;
  test)
    if [[ $# -eq 0 ]]; then
      set -- ./sampleapp
    fi
    (cd "$SCRIPT_DIR" && GOWORK=off go test "$@")
    ;;
  serve-test)
    if [[ $# -eq 0 ]]; then
      set -- ./sampleapp
    fi
    (cd "$SCRIPT_DIR" && GOWORK=off go test "$@")
    echo
    echo "Tests passed. Starting sample app at http://localhost:4242 (Ctrl+C to stop)"
    cd "$SCRIPT_DIR"
    exec env GOWORK=off go run ./sampleapp
    ;;
  *)
    echo "Unknown command: $command" >&2
    exit 1
    ;;
esac
