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

### [GAMMA-001] Frontend scaffold + WS client
- Owner: Agent Gamma
- Depends on: BETA-001
- Acceptance:
  - `frontend/` exists as a Vite + React 18 + TypeScript app, started with `pnpm dev`.
  - `frontend/src/net/ws.ts` opens a WS to `ws://localhost:8080/ws`, sends `hello`, decodes envelopes into strict TS types mirrored from `backend/internal/proto`.
  - `frontend/src/state/store.ts` is a Zustand store with slices for `connection`, `player`, `hero`, `castle`, `map`, `inFlight`.
  - Clock skew is computed from `serverTime` on every inbound envelope and exposed as `useServerNow()`.
  - A bare `App.tsx` renders connection status + bootstrap snapshot JSON for verification.
- Files to touch:
  - `frontend/package.json`, `frontend/vite.config.ts`, `frontend/tsconfig.json`
  - `frontend/src/main.tsx`, `frontend/src/App.tsx`
  - `frontend/src/net/ws.ts`
  - `frontend/src/state/store.ts`
  - `frontend/src/proto/*.ts` (mirrors `backend/internal/proto`)
- Doc refs: [architecture.md §7](./architecture.md#7-websocket-protocol-contract), [architecture.md §10](./architecture.md#10-authoritative-time--anti-cheat-poc-posture)

### [GAMMA-002] Konva map: nodes, edges, hero token, click-to-move
- Owner: Agent Gamma
- Depends on: GAMMA-001
- Acceptance:
  - `react-konva` `Stage` renders all six PoC nodes (color-coded by `kind`) and seven edges per the seeded coordinates in [architecture.md §9.2](./architecture.md#92-poc-seeded-graph-6-nodes-7-edges).
  - The hero is drawn as a token on its `current_node_id`.
  - Clicking a connected node (must share an edge with current node) emits `move.request`.
  - On `move.update`, the hero token animates along the edge from `departAt` to `arriveAt`, position interpolated using `useServerNow()`; arrives at the target node within ±0.5s of `arriveAt`.
  - In-flight, clicks on other nodes are rejected client-side with a toast ("Hero is moving").
- Files to touch:
  - `frontend/src/map/Map.tsx`
  - `frontend/src/map/Node.tsx`
  - `frontend/src/map/Edge.tsx`
  - `frontend/src/map/HeroToken.tsx`
- Doc refs: [architecture.md §9](./architecture.md#9-map-model), [architecture.md §5](./architecture.md#5-data-flow-move-command)

### [GAMMA-003] HUD: gold, army, hero panel, combat log modal
- Owner: Agent Gamma
- Depends on: GAMMA-002, BETA-003, BETA-004
- Acceptance:
  - Gold readout updates per second by extrapolating from last `castle.tick`'s `goldPerMin`, snapping to authoritative value on each new `castle.tick` / `unit.buy` / `combat.resolved`.
  - Army panel lists `hero_units` with a `[+]` button that emits `unit.buy { qty: 1 }`; `[+10]` shortcut. Disable if `gold < cost * qty`.
  - Hero panel shows `speedEffective` from latest `hero.state` and a small badge while respawning.
  - On `combat.resolved`, a modal displays outcome + full round log.
- Files to touch:
  - `frontend/src/hud/Gold.tsx`
  - `frontend/src/hud/ArmyPanel.tsx`
  - `frontend/src/hud/HeroPanel.tsx`
  - `frontend/src/hud/CombatModal.tsx`
- Doc refs: [game_rules.md §5](./game_rules.md#5-anti-snowball-b-upkeep-gold-cost), [game_rules.md §6](./game_rules.md#6-combat), [architecture.md §7.3](./architecture.md#73-broadcast-frequency)

### [LEAD-001] PoC end-to-end smoke test + balance review
- Owner: Project Lead Agent
- Depends on: GAMMA-003
- Acceptance:
  - Scripted playthrough: connect → buy 10 Pikemen → move to Crossroads → move to Bandit Camp → combat resolves → either victory (gold +500) or defeat (respawn at castle, lockout for 60s).
  - All client UI numbers reconcile with `combat_logs` row and `players.gold` final state.
  - Slowdown is **observable**: a 200-unit army takes visibly longer than a 10-unit army on the same edge.
  - Findings written to a new `docs/poc_review.md` with at least 5 follow-up tickets filed in this Backlog.
- Files to touch: `docs/poc_review.md`, this file
- Doc refs: [game_rules.md](./game_rules.md), [architecture.md](./architecture.md)

### [OPS-002] CI: lint + test for backend + frontend
- Owner: TBD
- Depends on: ALPHA-003, GAMMA-001
- Acceptance:
  - GitHub Actions workflow: `go vet`, `go test`, `pnpm lint`, `pnpm typecheck`, `pnpm test` on each PR.
  - Postgres + Redis services provisioned for backend integration tests.

---

## In Progress

_(empty — agents move tasks here on pickup)_

---

## Review

_(empty — agents move tasks here when acceptance criteria pass)_

---

## Done

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
