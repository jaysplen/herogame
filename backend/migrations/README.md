# Database migrations (goose)

SQL migrations for the PoC schema and world seed. See [docs/architecture.md](../../docs/architecture.md) §8–§9.

## Prerequisites

- Postgres 16 running (`make dev` from repo root)
- [goose](https://github.com/pressly/goose) CLI:

  ```bash
  go install github.com/pressly/goose/v3/cmd/goose@latest
  ```

- `DATABASE_URL` in environment or repo `.env` (see `.env.example`)

## From repo root

```bash
make migrate              # goose up
make migrate-status       # current version
```

## From backend/

```bash
export DATABASE_URL=postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable

goose -dir migrations postgres "$DATABASE_URL" up
goose -dir migrations postgres "$DATABASE_URL" status
goose -dir migrations postgres "$DATABASE_URL" down   # one step
goose -dir migrations postgres "$DATABASE_URL" reset  # all down
```

## Files

| File | Purpose |
|---|---|
| `00001_init.sql` | All 10 PoC tables + partial index on `movement_orders` |
| `00002_seed_world.sql` | 6 nodes, 14 edges, 2 players, castles, heroes, pikeman, Bandit Camp |

## Programmatic (server startup)

Set `DATABASE_URL` and optionally `RUN_MIGRATIONS=1` (default on when `DATABASE_URL` is set):

```bash
cd backend && RUN_MIGRATIONS=1 go run ./cmd/server
```

Uses [internal/store/migrate.go](../internal/store/migrate.go).
