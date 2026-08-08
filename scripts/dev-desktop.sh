#!/usr/bin/env bash
# hamix-desktop from repo root: ./scripts/dev-desktop.sh  (needs .env / DATABASE_URL)
# Schema migrate is a separate step: ./scripts/migrate.sh
# Browser/API path remains: ./scripts/dev.sh
#
# Usage: ./scripts/dev-desktop.sh [--migrate] [--live] [--help]
#
# Flags:
#   --migrate   Run ./scripts/migrate.sh first (convenience sugar)
#   --live      Use `wails dev` (Vite HMR) instead of a production SPA embed
#   --help, -h  Show options
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"

ENV_FILE="$ROOT/.env"
if [[ ! -f "$ENV_FILE" ]]; then
  echo ".env not found at: $ENV_FILE" >&2
  echo "Copy .env.example to .env and set DATABASE_URL:" >&2
  echo "  cp .env.example .env" >&2
  echo "See CONTRIBUTING.md for setup." >&2
  exit 1
fi

RUN_MIGRATE=0
LIVE=0

show_help() {
  sed -n '2,12p' "$0" | sed 's/^# \{0,1\}//'
  cat <<'EOF'

Default (no --live): npm build → copy into cmd/hamix-desktop/frontend/dist →
  ./scripts/build-desktop.sh (go build -tags desktop,production) → run binary.
DSN: DATABASE_URL in .env (or first-run / Settings UI). See docs/adr/ADR-0095-desktop-wails-host.md.
EOF
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --migrate)
      RUN_MIGRATE=1
      shift
      ;;
    --live)
      LIVE=1
      shift
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
DESKTOP_EXE="$ROOT/hamix-desktop-dev"
MCP_EXE="$ROOT/hamix-agent-mcp"
[ "$GOOS" = "windows" ] && DESKTOP_EXE="$ROOT/hamix-desktop-dev.exe" && MCP_EXE="$ROOT/hamix-agent-mcp.exe"
DESKTOP_DIR="$ROOT/cmd/hamix-desktop"
COPY_SCRIPT="$DESKTOP_DIR/scripts/copy-web-dist.mjs"

if [[ "$RUN_MIGRATE" -eq 1 ]]; then
  "$ROOT/scripts/migrate.sh"
fi

go mod download
( cd "$ROOT/web" && npm install )
go build -o "$MCP_EXE" ./cmd/hamix-agent-mcp
export PATH="$ROOT${PATH:+:$PATH}"

if [[ "$LIVE" -eq 1 ]]; then
  if ! command -v wails >/dev/null 2>&1; then
    echo "wails CLI not found (required for --live)." >&2
    echo "Install once: go install github.com/wailsapp/wails/v2/cmd/wails@latest" >&2
    echo "Or omit --live to build the SPA and run the embedded desktop binary." >&2
    exit 1
  fi
  cd "$DESKTOP_DIR"
  exec wails dev
fi

( cd "$ROOT/web" && npm run build )
node "$COPY_SCRIPT"
# Wails tags live only in build-desktop.* (ADR-0095 / manual builds guide).
"$ROOT/scripts/build-desktop.sh" --out "$DESKTOP_EXE"

echo "starting $DESKTOP_EXE (Ctrl+C quits)" >&2
exec "$DESKTOP_EXE"
