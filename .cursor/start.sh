#!/usr/bin/env bash
# Cloud agent launch step — runs from the repo root on every agent run.
set -euo pipefail

sudo service postgresql start

if [[ ! -f .env ]]; then
  echo ".env missing — the environment Build did not complete; re-run it from the Cursor environment settings" >&2
  exit 1
fi

# The snapshot already carries the default-branch schema; this picks up any
# migration added on the agent's own branch.
./scripts/migrate.sh
