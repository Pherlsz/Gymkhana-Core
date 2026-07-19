#!/usr/bin/env bash
set -Eeuo pipefail

MODE="${1:-quick}"
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

log() { printf '\n==> %s\n' "$*"; }
need() { command -v "$1" >/dev/null 2>&1 || { echo "missing required command: $1" >&2; exit 1; }; }

need go
need make

case "$MODE" in
  quick)
    log "Formatting, vet and unit tests"
    go mod download
    make format-check vet test
    ;;
  full)
    log "Complete local quality gate"
    go mod download
    make setup-quality
    make format-check vet lint test test-race fuzz-smoke
    ;;
  security)
    log "Security scans"
    go mod download
    make security
    ;;
  *)
    echo "usage: scripts/verify-local.sh [quick|full|security]" >&2
    exit 2
    ;;
esac

log "Local verification completed"
