# Backend (Go)

Game server: HTTP health check, WebSocket gateway, 1 Hz tick engine, Postgres persistence, Redis arrival queue.

## Intended layout

Per [docs/architecture.md §3](../docs/architecture.md#3-repository-layout):

```
backend/
├── go.mod
├── cmd/server/main.go
├── internal/
│   ├── world/      # map nodes, edges, travel time math
│   ├── hero/       # hero state, army stack
│   ├── economy/    # castle gold, unit purchase, upkeep
│   ├── combat/     # deterministic auto-resolve
│   ├── tick/       # arrivals ZSET, economy/upkeep sweeps
│   ├── ws/         # WebSocket gateway
│   ├── store/      # sqlc-generated queries
│   └── proto/      # shared message types
├── migrations/     # goose SQL
└── queries/        # sqlc input
```

## Run

```bash
# from repo root
make dev
cp ../.env.example ../.env   # optional

cd backend
export DATABASE_URL=postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable
go run ./cmd/server          # runs migrations when DATABASE_URL is set (disable: RUN_MIGRATIONS=0)
curl http://localhost:8080/healthz
```

Migrations only: `make migrate` from repo root, or see [migrations/README.md](./migrations/README.md).

## Regenerate sqlc

```bash
sqlc generate   # from backend/, after editing queries/*.sql
go test ./...   # includes store integration tests (needs Postgres)
```

## WebSocket (BETA-001)

Requires `DATABASE_URL`. Connect to `ws://localhost:8080/ws`, send `hello` within 5s:

```json
{"type":"hello","payload":{"playerId":1},"seq":1,"serverTime":0}
```

## Next tasks

- **BETA-002** — Redis arrivals ZSET + tick engine
- **BETA-003** — move/buy handlers
