#!/usr/bin/env bash
# Builds sidecars/hamix-draft-agent, writes the repo-root launcher, and sets
# HAMIX_DRAFT_AGENT_BIN so taskapi/desktop can spawn it without PATH luck.
# Required: missing sidecar dir or a failed build exits 1 (fail-boot).
# Safe to source from scripts/dev.sh and scripts/dev-desktop.sh.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SIDECAR="$ROOT/sidecars/hamix-draft-agent"
if [[ ! -d "$SIDECAR" ]]; then
  echo "sidecars/hamix-draft-agent is missing; draft-assist cannot start without it." >&2
  exit 1
fi

(
  cd "$SIDECAR"
  if command -v pnpm >/dev/null 2>&1; then
    pnpm install --silent
    pnpm run build
  else
    npm install --silent
    npm run build
  fi
)

BUNDLE="$SIDECAR/dist/hamix-draft-agent.js"
if [[ ! -f "$BUNDLE" ]]; then
  echo "sidecar build did not produce $BUNDLE" >&2
  exit 1
fi

cp "$BUNDLE" "$ROOT/hamix-draft-agent.js"
cat > "$ROOT/hamix-draft-agent" <<'LAUNCHER'
#!/usr/bin/env sh
exec node "$(dirname "$0")/hamix-draft-agent.js" "$@"
LAUNCHER
chmod +x "$ROOT/hamix-draft-agent"
export HAMIX_DRAFT_AGENT_BIN="$ROOT/hamix-draft-agent"
