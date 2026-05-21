#!/usr/bin/env bash
# Run frontend Playwright smoke (BACKLOG-006). Requires Postgres + Redis (make dev).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

export DATABASE_URL="${DATABASE_URL:-postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"

cd "$ROOT/frontend"
if [[ ! -d node_modules/@playwright/test ]]; then
  npm ci
  npx playwright install chromium
fi

# Let playwright.config.ts start backend (:18080) + Vite; do not reuse :8080 dev server.
unset CI
npm run test:e2e
