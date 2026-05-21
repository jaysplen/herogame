# Local development setup

> **Task:** OPS-001  
> **Last updated:** 2026-05-20

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with Compose v2 (`docker compose`)
- [Make](https://www.gnu.org/software/make/)
- **Go 1.22+** (ALPHA-001+): `go version` must succeed
- (ALPHA-002+) [goose](https://github.com/pressly/goose) for migrations — install once:
  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```
  Ensure `$(go env GOPATH)/bin` is on your `PATH` (e.g. add to `~/.bashrc`: `export PATH="$HOME/go/bin:$PATH"`).
- (GAMMA-001+) pnpm for the frontend

### Verify Go

```bash
go version          # expect go1.22 or newer
which go            # e.g. /snap/bin/go after: snap install go --classic
cd backend && go vet ./... && go build ./cmd/server
```

If `which go` shows `~/.local/go-sdk/bin/go` (an older agent-local install) **before** `/snap/bin/go`, your shell may be using Go 1.22.10 from that tree instead of a system/snap install. Both work for this repo; to prefer snap, put `/snap/bin` earlier in `PATH` or remove `~/.local/go-sdk` from `PATH`.

## Quick start

```bash
cp .env.example .env          # optional; defaults match docker-compose
make dev                      # postgres:16 + redis:7, then migrations if present
```

Run **`make dev` without `sudo`**. Under sudo, `goose` and `go` from your user profile are not on PATH. If Docker permission errors occur, add your user to the `docker` group rather than using sudo.

**Snap Go + `make server`:** If you see `XDG_RUNTIME_DIR` / `permission denied`, the dev script sets a fallback under `/tmp` and builds `backend/bin/herogame-server`. Prefer a non-snap Go when possible: `sudo snap install go --classic` or install from https://go.dev/dl/.

Verify services:

```bash
make ps
docker compose exec postgres pg_isready -U herogame
docker compose exec redis redis-cli ping
```

## Connection strings

From `.env.example` (host machine → published ports):

| Variable | Default |
|---|---|
| `DATABASE_URL` | `postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable` |
| `REDIS_URL` | `redis://localhost:6379/0` |

Inside another Docker container on the `herogame` network, use hostnames `postgres` and `redis` instead of `localhost`.

## Make targets

| Target | Description |
|---|---|
| `make dev` | `docker compose up -d --wait`, then `make migrate` |
| `make down` | Stop containers; keep volumes |
| `make down-v` | Stop containers and delete `postgres_data` / `redis_data` |
| `make migrate` | Run `goose up` when `backend/migrations/` exists (ALPHA-002) |
| `make logs` | Follow compose logs |

## Backend (ALPHA-001)

```bash
make dev          # Postgres + Redis
make server       # go run with DATABASE_URL; fails if :8080 has no /ws
curl http://localhost:8080/healthz      # {"status":"ok"}
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:8080/ws   # expect 400, not 404
```

If the UI shows **websocket error** but `/healthz` works, port 8080 is usually a **stale** `/tmp/herogame-server` (no `DATABASE_URL`). Free the port and use `make server`:

```bash
kill $(lsof -t -i:8080)   # WSL/Linux
make server
```

## Frontend (GAMMA-001)

```bash
cd frontend && npm install && npm run dev
```

Opens http://localhost:5173 (WebSocket to `ws://localhost:8080/ws`, Player 1).
Use `?playerId=2` to connect as Player 2.

## Full PoC loop

```bash
make dev
cd backend && go run ./cmd/server
cd frontend && npm run dev
```

See [architecture.md §12](./architecture.md#12-build--run-reference-not-poc-scope).

## CI (BACKLOG-001 / BACKLOG-006)

Pull requests and pushes to `master`/`main` run [.github/workflows/ci.yml](../.github/workflows/ci.yml):

- **backend** — `go vet`, `go test ./...` with Postgres 16 + Redis 7 service containers
- **frontend** — `npm ci`, `npm run build` (TypeScript + Vite)
- **e2e** — Playwright smoke (starts game server on `:18080` + Vite with `VITE_E2E=1`)

Reproduce locally:

```bash
make dev
cd backend && go vet ./... && go test ./... -count=1
cd frontend && npm ci && npm run build
```

## Playwright smoke (BACKLOG-006)

Requires Postgres + Redis (`make dev`). Playwright starts the backend on **port 18080** (so it does not collide with a dev server on `:8080` that might lack `/ws`) and Vite on `:5173`.

```bash
cd frontend && npm ci
npx playwright install chromium
npm run test:e2e
# or from repo root:
bash scripts/playwright-smoke.sh
```

The test uses DOM shortcuts (`data-testid` recruit/move buttons) instead of Konva canvas clicks. Interactive UI: `npm run test:e2e:ui`.

## Strategy-loop expansion branch

Integration branch: `feature/epic-singleplayer-world`.

Highlights:
- expanded map graph and seeded world (`00003_world_expansion.sql`)
- neutral creep roaming state + broadcasts
- conquerable resource nodes with per-second income
- multi-resource unit/build costs and castle progression tiers
- grace-period gating and respawn UX hooks
- objective tracking (eliminate enemy hero 5 times)
- player-scoped broadcast helper for multiplayer migration seam

## Troubleshooting

- **Port in use:** set `POSTGRES_PORT` / `REDIS_PORT` in `.env` and update `DATABASE_URL` / `REDIS_URL`.
- **Migrations skipped:** expected until ALPHA-002 adds `backend/migrations/`.
- **Reset game state (keep DB, fresh heroes/creeps/resources):** `make reset` then restart `make server`
- **Full database wipe:** `make down-v && make dev`
