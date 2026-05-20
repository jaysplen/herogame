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

## Bootstrap

This directory is **not implemented yet**. It is created by task **[ALPHA-001]** in [docs/agent_tasks.md](../docs/agent_tasks.md):

- Initialize `go.mod` (`github.com/herogame/backend`)
- Chi router + `GET /healthz`
- Structured logging via `log/slog`

Follow-on: ALPHA-002 (migrations + seed), ALPHA-003 (sqlc store).
