# Local development setup

> **Task:** OPS-001  
> **Last updated:** 2026-05-20

## Prerequisites

- [Docker](https://docs.docker.com/get-docker/) with Compose v2 (`docker compose`)
- [Make](https://www.gnu.org/software/make/)
- **Go 1.22+** (ALPHA-001+): `go version` must succeed
- (ALPHA-002+) [goose](https://github.com/pressly/goose) for migrations
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
make dev                              # optional: Postgres + Redis
cd backend && go run ./cmd/server
curl http://localhost:8080/healthz      # {"status":"ok"}
```

## Full PoC loop (once backend/frontend exist)

```bash
make dev
cd backend && go run ./cmd/server
cd frontend && pnpm dev
```

See [architecture.md §12](./architecture.md#12-build--run-reference-not-poc-scope).

## Troubleshooting

- **Port in use:** set `POSTGRES_PORT` / `REDIS_PORT` in `.env` and update `DATABASE_URL` / `REDIS_URL`.
- **Migrations skipped:** expected until ALPHA-002 adds `backend/migrations/`.
- **Reset database:** `make down-v && make dev`
