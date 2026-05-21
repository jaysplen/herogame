#!/usr/bin/env bash
# Start the real game server (WebSocket + DB). Fails fast if :8080 has no /ws.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export DATABASE_URL="${DATABASE_URL:-postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"
export HTTP_ADDR="${HTTP_ADDR:-:8080}"

ws_code() {
  curl -s -o /dev/null -w "%{http_code}" "http://127.0.0.1:8080/ws" 2>/dev/null || echo "000"
}

if curl -sf "http://127.0.0.1:8080/healthz" >/dev/null 2>&1; then
  code="$(ws_code)"
  if [[ "$code" == "404" ]]; then
    echo "Port 8080 is up but /ws is missing (stale binary without DATABASE_URL?)."
    echo "Stop it, then re-run this script. Example:"
    echo "  kill \$(lsof -t -i:8080)"
    exit 1
  fi
  if [[ "$code" == "400" || "$code" == "426" ]]; then
    echo "Game server already listening on :8080 with WebSocket — not starting another."
    exit 0
  fi
fi

cd "$ROOT/backend"
echo "Starting go server on ${HTTP_ADDR} (DATABASE_URL set, /ws enabled)..."
exec go run ./cmd/server
