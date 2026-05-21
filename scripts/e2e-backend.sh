#!/usr/bin/env bash
# Start game server for Playwright e2e (expects Postgres + Redis on localhost).
#
# Before launching the server, run cmd/e2e-reset which:
#   * pins the Bandit Camp creep at node 5 with a 10-minute leg so
#     it stays put for the entire smoke-test run (otherwise the
#     migration-seeded 30 s leg lapses well before Playwright clicks
#     "move to bandit camp"),
#   * kills Wolf Den, repositions both heroes, clears in-flight
#     movements / combat logs, FLUSHDBs Redis.
# This makes the smoke test deterministic without requiring psql or
# redis-cli on the CI image.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export DATABASE_URL="${DATABASE_URL:-postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"
export HTTP_ADDR="${HTTP_ADDR:-:18080}"

cd "$ROOT/backend"
if command -v goose >/dev/null 2>&1 && [[ -d migrations ]]; then
  goose -dir migrations postgres "$DATABASE_URL" up || true
fi

echo "Resetting game state for Playwright e2e..."
go run ./cmd/e2e-reset

exec go run ./cmd/server
