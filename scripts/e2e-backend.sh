#!/usr/bin/env bash
# Start game server for Playwright e2e (expects Postgres + Redis on localhost).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export DATABASE_URL="${DATABASE_URL:-postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"
export HTTP_ADDR="${HTTP_ADDR:-:18080}"

cd "$ROOT/backend"
if command -v goose >/dev/null 2>&1 && [[ -d migrations ]]; then
  goose -dir migrations postgres "$DATABASE_URL" up || true
fi
exec go run ./cmd/server
