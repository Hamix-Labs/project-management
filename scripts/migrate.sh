#!/usr/bin/env bash
# Schema migrate — step 1 before dev servers. See CONTRIBUTING.md.
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
echo "=== schema migrate ==="
go run ./cmd/dbcheck -migrate
