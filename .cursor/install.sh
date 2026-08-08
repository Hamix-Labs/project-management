#!/usr/bin/env bash
# Cloud agent build step — runs from the repo root, must be idempotent.
# Disk state is snapshotted; running processes and exported vars are not.
set -euo pipefail

# Agent-local Postgres: every cloud agent gets its own VM, so this is a private
# database per agent. Not a credential — localhost only, never reachable.
DSN='postgres://hamix:hamix@localhost:5432/hamix?sslmode=disable'

GO_VERSION="$(sed -n 's/^go \([0-9.]*\)$/\1/p' go.mod)"

if ! /usr/local/go/bin/go version 2>/dev/null | grep -q "go${GO_VERSION}"; then
  curl -fsSL "https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz" -o /tmp/go.tgz
  sudo rm -rf /usr/local/go
  sudo tar -C /usr/local -xzf /tmp/go.tgz
  rm -f /tmp/go.tgz
fi
sudo ln -sf /usr/local/go/bin/go /usr/local/bin/go
sudo ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt

sudo apt-get update
sudo apt-get install -y postgresql
sudo service postgresql start
sudo -u postgres psql -tAc "SELECT 1 FROM pg_roles WHERE rolname='hamix'" | grep -q 1 \
  || sudo -u postgres psql -c "CREATE ROLE hamix LOGIN PASSWORD 'hamix' SUPERUSER;"
sudo -u postgres psql -tAc "SELECT 1 FROM pg_database WHERE datname='hamix'" | grep -q 1 \
  || sudo -u postgres createdb -O hamix hamix

# internal/envload requires the file, not just the variable. Written here so the
# build-time migrate below works and the result is captured in the snapshot.
printf 'DATABASE_URL=%s\n' "$DSN" > .env

# Scoped to ./cmd/... on purpose: ./... would descend into web/node_modules.
go mod download
go build ./cmd/...
./scripts/migrate.sh
./scripts/check-web.sh --install-only
