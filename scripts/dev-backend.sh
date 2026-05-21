#!/usr/bin/env bash
# Start the real game server (WebSocket + DB). Fails fast if :8080 has no /ws.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export DATABASE_URL="${DATABASE_URL:-postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"
export HTTP_ADDR="${HTTP_ADDR:-:8080}"

# WSL often sets XDG_RUNTIME_DIR=/run/user/1000 but that path does not exist.
# Snap Go then fails: mkdir: cannot create directory '/run/user/1000/'.
if [[ -z "${XDG_RUNTIME_DIR:-}" || ! -d "${XDG_RUNTIME_DIR}" || ! -w "${XDG_RUNTIME_DIR}" ]]; then
  export XDG_RUNTIME_DIR="/tmp/xdg-runtime-$(id -u)"
fi
mkdir -p "$XDG_RUNTIME_DIR"
chmod 700 "$XDG_RUNTIME_DIR" 2>/dev/null || true

pick_go() {
  local g
  for g in /usr/local/go/bin/go "$HOME/go/bin/go" "$HOME/.local/go-sdk/bin/go"; do
    [[ -n "$g" && -x "$g" ]] || continue
    echo "$g"
    return 0
  done
  command -v go 2>/dev/null || true
}
GO_BIN="$(pick_go || true)"
if [[ -z "${GO_BIN:-}" ]]; then
  echo "go not found. Install Go 1.22+ (prefer /usr/local/go or snap install go --classic)."
  exit 1
fi

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
BIN="$ROOT/backend/bin/herogame-server"
echo "Starting game server on ${HTTP_ADDR} (DATABASE_URL set, /ws enabled)..."
if [[ ! -x "$BIN" ]] || [[ cmd/server/main.go -nt "$BIN" ]]; then
  echo "Building $BIN ..."
  "$GO_BIN" build -o "$BIN" ./cmd/server
fi
exec "$BIN"
