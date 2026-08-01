#!/usr/bin/env bash
# Build hamix-desktop with required Wails tags (SSOT for go build flags).
# See https://wails.io/docs/guides/manual-builds/ and ADR-0095.
#
# Usage: ./scripts/build-desktop.sh [--out path] [--help]
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

OUT=""
show_help() {
  cat <<'EOF'
Build cmd/hamix-desktop with Wails production tags.

Usage: ./scripts/build-desktop.sh [--out path] [--help]

  --out path   Output binary (default: repo-root hamix-desktop-dev[.exe])
  --help, -h   Show this help

Always uses: go build -tags desktop,production
Do not use plain `go build ./cmd/hamix-desktop` — untagged builds are a Wails stub.
Dev console binary intentionally omits -H windowsgui (keeps Ctrl+C / logs).
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --out)
      OUT="$2"
      shift 2
      ;;
    --help|-h)
      show_help
      exit 0
      ;;
    *)
      echo "unknown flag: $1 (try --help)" >&2
      exit 2
      ;;
  esac
done

GOOS="$(go env GOOS)"
if [[ -z "$OUT" ]]; then
  OUT="$ROOT/hamix-desktop-dev"
  [[ "$GOOS" = "windows" ]] && OUT="$ROOT/hamix-desktop-dev.exe"
fi

go build -tags desktop,production -o "$OUT" ./cmd/hamix-desktop
echo "built $OUT (-tags desktop,production)" >&2
