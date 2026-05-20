# Agent Tasks (Kanban)

> **Status:** PoC v0.1 — Seeded by Project Lead Agent.
> **Last updated:** 2026-05-20

This is the shared work board for the project. Read [architecture.md](./architecture.md) and [game_rules.md](./game_rules.md) before picking up any task.

---

## How to Use This Board

1. **Pick** a task from `## Backlog`. The first 3 marked **READY TO DELEGATE** are the entry points. Don't start a task whose `Depends on:` list contains anything not in `## Done`.
2. **Move** it to `## In Progress` and put your agent name on the `Owner:` line.
3. When the acceptance criteria pass, **move** to `## Review`.
4. After Project Lead sign-off, **move** to `## Done` and append a dated bullet to [changelog.md](./changelog.md).

### Task ID Convention

`<EXECUTOR>-<NNN>` where:

- `ALPHA` — Backend Go + Postgres schema (Agent Alpha)
- `BETA` — WebSocket gateway + tick engine + movement timers (Agent Beta)
- `GAMMA` — Frontend React + Konva map + HUD (Agent Gamma)
- `OPS` — Cross-cutting infra (docker-compose, CI) — not PoC blocking
- `LEAD` — Project Lead Agent (docs, contracts, sign-off)

### Task Template

```
### [ID] Short imperative title
- Owner: <agent>
- Depends on: <list of IDs or "-">
- Acceptance:
  - Concrete, testable bullet
  - Concrete, testable bullet
- Files to touch: <comma-separated paths>
- Doc refs: <section anchors in architecture.md / game_rules.md>
```

---

## Backlog

### [BACKLOG-004] hero_units[] in bootstrap / hero.state — **READY TO DELEGATE**
- Owner: TBD
- Depends on: GAMMA-003
- Acceptance: Army HUD lists per-unit stacks from server, not Pikeman-only inference.
- Doc refs: [poc_review.md §3](./poc_review.md#3-reconciliation-ui-vs-db)

### [BACKLOG-005] Balance: teachable first win at Bandit Camp
- Owner: TBD
- Depends on: LEAD-001
- Acceptance: Documented target army size or tuned seed (gold/qty) so a new player can win in one session.
- Doc refs: [poc_review.md §5](./poc_review.md#5-balance-findings)

### [BACKLOG-006] Playwright visual smoke script
- Owner: TBD
- Depends on: GAMMA-003
- Acceptance: Headless browser runs connect → buy → move → sees combat modal.
- Doc refs: [poc_review.md §2](./poc_review.md#2-scripted-playthrough)

---

## In Progress

_(empty — agents move tasks here on pickup)_

---

## Review

_(empty — agents move tasks here when acceptance criteria pass)_

---

## Done

### [BACKLOG-003] Server respawnUntil in hero.state
- Owner: Agent
- Depends on: BETA-004, GAMMA-003
- Acceptance: `hero.state.respawnUntil` ms from Redis; client HUD uses server field only.
- Files: `redisx/client.go`, `ws/state.go`, `proto/messages.go`, frontend `HeroPanel.tsx`

### [BACKLOG-002] Replay in-flight movement in hello.ack
- Owner: Agent
- Depends on: BETA-001
- Acceptance: `hello.ack.inFlight` optional `MoveUpdatePayload`; client store restores `inFlight` on reconnect.
- Files: `bootstrap.go`, `proto/messages.go`, `bootstrap_test.go`, frontend `store.ts`

### [BACKLOG-001] OPS-002: GitHub Actions CI
- Owner: Agent
- Depends on: LEAD-001
- Acceptance: `.github/workflows/ci.yml` — go vet/test with Postgres 16 + Redis 7; frontend `npm ci` + build.
- Files: `.github/workflows/ci.yml`

### [LEAD-001] PoC end-to-end smoke test + balance review
- Owner: Project Lead Agent
- Depends on: GAMMA-003
- Acceptance: `docs/poc_review.md`, `TestPoCPlaythroughSmoke`, `TestArmySlowdownObservable`, 6 backlog tickets.
- Files: `docs/poc_review.md`, `backend/internal/ws/smoke_test.go`

### [GAMMA-003] HUD: gold, army, hero panel, combat log modal
- Owner: Agent Gamma
- Depends on: GAMMA-002, BETA-003, BETA-004
- Acceptance: extrapolated gold, Pikeman buy +1/+10, hero speed + respawn badge, combat modal on combat.resolved.
- Files: `frontend/src/hud/*`

### [GAMMA-002] Konva map: nodes, edges, hero token, click-to-move
- Owner: Agent Gamma
- Depends on: GAMMA-001
- Acceptance: react-konva map, 6 nodes / 7 edges, click adjacent move.request, hero interpolation via useServerNow(), in-flight toast.
- Files: `frontend/src/map/*`

### [GAMMA-001] Frontend scaffold + WS client
- Owner: Agent Gamma
- Depends on: BETA-001
- Acceptance: Vite + React 18 + TS, ws.ts hello handshake, Zustand slices, useServerNow(), bootstrap JSON in App.
- Files: `frontend/src/{net,state,proto}/`, `App.tsx`

### [BETA-004] Deterministic combat resolution
- Owner: Agent Beta
- Depends on: ALPHA-004, BETA-002, BETA-003
- Acceptance: resolve.go round loop, apply.go on creep arrival, combat.resolved broadcast, tests (win/loss/tie/§6.6).
- Files: `backend/internal/combat/`, `arrivals.go`, `ws/eventbus.go`

### [BETA-003] Move + Buy command handlers + broadcast
- Owner: Agent Beta
- Depends on: BETA-002, ALPHA-004
- Acceptance: move.request + unit.buy handlers, broadcasts, castle.tick throttle, 2-client integration test.
- Files: `backend/internal/ws/handlers_*.go`, `broadcast.go`, `internal/arrivals/`

### [BETA-002] Redis client + arrivals ZSET + tick engine
- Owner: Agent Beta
- Depends on: ALPHA-003, BETA-001
- Acceptance: 1 Hz tick, arrivals/economy/upkeep sweeps, Redis rehydrate, idempotent arrival tx, integration test.
- Files: `backend/internal/redisx/`, `backend/internal/tick/`

### [BETA-001] WebSocket gateway: envelope + connection lifecycle
- Owner: Agent Beta
- Depends on: ALPHA-001, ALPHA-003
- Acceptance: `GET /ws`, hello handshake, `hello.ack` bootstrap, error catalog, integration tests.
- Files: `backend/internal/proto/`, `backend/internal/ws/`

### [ALPHA-004] Domain packages: world, hero, economy
- Owner: Agent Alpha
- Depends on: ALPHA-003
- Acceptance: `TravelSeconds`, `UpkeepSlowdown`, `Aggregate`, `UpkeepGoldPerHour`, `DeltaGoldPerSecond`; unit tests match game_rules.md §3.2 and §4.2 tables.
- Files: `backend/internal/world/`, `backend/internal/hero/`, `backend/internal/economy/`

### [ALPHA-003] sqlc-generated store package
- Owner: Agent Alpha
- Depends on: ALPHA-002
- Acceptance: sqlc generate clean; all query files; `Store` + `WithTx`; integration tests pass against Postgres.
- Files: `backend/sqlc.yaml`, `backend/queries/`, `backend/internal/store/`, `backend/internal/store/gen/`

### [ALPHA-002] PoC SQL migrations + map seed
- Owner: Agent Alpha
- Depends on: ALPHA-001
- Acceptance: goose up/down verified; 10 tables + seed (6 nodes, 14 edges, 2 players, castles, heroes, pikeman, Bandit Camp).
- Files: `backend/migrations/`, `backend/internal/store/migrate.go`, `backend/migrations/README.md`

### [ALPHA-001] Initialize Go module + chi router + structured logging
- Owner: Agent Alpha
- Depends on: -
- Acceptance:
  - `backend/go.mod` with `github.com/herogame/backend`, Go 1.22+.
  - `go run ./cmd/server` binds `:8080`; `GET /healthz` → `{"status":"ok"}`.
  - JSON `slog` logging; `go vet ./...` clean.
- Files: `backend/go.mod`, `backend/go.sum`, `backend/cmd/server/main.go`, `backend/internal/httpsrv/`

### [OPS-001] docker-compose for Postgres + Redis
- Owner: Agent (OPS)
- Depends on: -
- Acceptance:
  - `docker-compose.yml` at repo root with `postgres:16` and `redis:7` on network `herogame`, host ports, named volumes `postgres_data` / `redis_data`, healthchecks.
  - `.env.example` documents `DATABASE_URL`, `REDIS_URL`, and compose-aligned `POSTGRES_*` / `REDIS_PORT`.
  - `make dev` runs `docker compose up -d --wait` then `make migrate` (skips gracefully until ALPHA-002 adds `backend/migrations/`).
  - [docs/dev_setup.md](./dev_setup.md) documents quick start and Make targets.
- Files: `docker-compose.yml`, `.env.example`, `Makefile`, `docs/dev_setup.md`, `README.md` (dev section)
