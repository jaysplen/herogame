#!/usr/bin/env bash
# Stop stray dev processes and print how to start fresh.
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"

echo "Stopping dev servers on :8080 / :5173 / :18080..."
pkill -f "herogame-server" 2>/dev/null || true
pkill -f "go run ./cmd/server" 2>/dev/null || true
pkill -f "vite" 2>/dev/null || true
sleep 1

for port in 8080 5173 18080; do
  if command -v lsof >/dev/null 2>&1; then
    pids=$(lsof -t -i:"$port" 2>/dev/null || true)
    if [[ -n "${pids:-}" ]]; then
      echo "Killing PID(s) on :$port: $pids"
      kill $pids 2>/dev/null || true
    fi
  fi
done
sleep 1

echo ""
echo "Ports:"
ss -tlnp 2>/dev/null | grep -E '5173|8080|18080' || echo "  5173, 8080, 18080 — free"
echo ""
echo "Start again in two terminals:"
echo ""
echo "  Terminal 1:"
echo "    cd $ROOT"
echo "    make dev      # Postgres + Redis (skip if already up)"
echo "    make server"
echo ""
echo "  Terminal 2:"
echo "    cd $ROOT/frontend"
echo "    npm run dev -- --host 0.0.0.0"
echo ""
echo "Then open http://127.0.0.1:5173"
