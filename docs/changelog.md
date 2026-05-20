# Changelog

> **Status:** Append-only project log.
> **Companion to:** [architecture.md](./architecture.md), [game_rules.md](./game_rules.md), [agent_tasks.md](./agent_tasks.md).

This file is the audit trail of every meaningful change to the project — code, schema, rules, scope. It is the first place a returning agent (human or AI) looks to catch up.

---

## How to Update

- **When** to append: after closing any task in [agent_tasks.md](./agent_tasks.md), after a doc change that affects scope/rules/schema, or after a balance-tuning constant change.
- **Format:** group entries under a date header `## YYYY-MM-DD` (UTC). Newest dates at the top. Within a date, newest entries at the top.
- **Per entry:** a single bullet starting with the task ID or `[doc]` tag, followed by a one-line summary. If non-trivial, add a sub-bullet citing the section anchor in the relevant doc.
- **Do not delete** entries. To correct a past entry, append a new bullet noting the correction.

Example:

```
## 2026-05-22

- ALPHA-002 — Goose migrations + map seed merged. All 10 PoC tables created; 6 nodes + 7 edges seeded.
  - Schema details: architecture.md §8.
- [doc] Adjusted Pikeman upkeep from 1.0 to 2.0 g/hr to make anti-snowball curve bite at ~1800-unit cap.
  - See game_rules.md §5.2 + §10.
```

---

## 2026-05-20

- BETA-003 — WS move.request and unit.buy handlers; broadcast hub; castle.tick throttle; refactored arrivals to `internal/arrivals` (break import cycle); 2-client move.update test.
- BETA-002 — Tick engine (1 Hz): Redis `arrivals:zset`, rehydrate on boot, arrival resolution + `move.arrived` broadcast, economy/upkeep sweeps; `GetMovementOrder`, `ListHeroes` queries.
- BETA-001 — WebSocket gateway: `GET /ws`, JSON envelopes, 5s hello timeout, `hello.ack` bootstrap from Postgres, hub for broadcast, gateway tests.
- BETA-001 — sqlc: `GetHeroByPlayer`, `ListEdges` queries.
- ALPHA-004 — Pure domain packages: `world` (travel/upkeep slowdown), `hero` (army aggregate), `economy` (upkeep + constants); tests match game_rules.md tables.
- ALPHA-003 — sqlc store: 7 query files, generated `internal/store/gen`, `Store` with `pgxpool` and `WithTx`, integration tests via `TEST_DATABASE_URL`.
- ALPHA-002 — Goose migrations: `00001_init.sql` (10 tables + partial index), `00002_seed_world.sql` (PoC map + players + Bandit Camp); `store.MigrateUp`; `RUN_MIGRATIONS` on server when `DATABASE_URL` set.
- ALPHA-001 — Go backend skeleton: chi router, `GET /healthz`, JSON `slog`, graceful shutdown (`backend/cmd/server`, `backend/internal/httpsrv/`).
- [fix] Makefile `migrate` target runs in a single shell so skip path does not fall through to `goose not found`.
- OPS-001 — Local dev infra: `docker-compose.yml` (Postgres 16 + Redis 7 on `herogame` network), `.env.example`, `Makefile` (`dev`, `down`, `migrate`), [dev_setup.md](./dev_setup.md).
  - `make migrate` skips until ALPHA-002 adds `backend/migrations/`.
- [chore] Initialized git repo + root plumbing (`.gitignore`, `README.md`, `pnpm-workspace.yaml`, `backend/README.md`, `frontend/README.md`).
- LEAD foundation pass — Project Lead Agent established the `/docs` knowledge base.
  - Created [architecture.md](./architecture.md): PoC scope, locked stack (Go + Postgres 16 + Redis 7 + WebSockets + React/Konva), system + sequence diagrams, full WS protocol contract, complete schema for all 10 PoC tables, tick engine design, 6-node seeded map, future-expansion hooks.
  - Created [game_rules.md](./game_rules.md): movement formula with worked examples, anti-snowball upkeep slowdown table (army_size 0–5000), anti-snowball upkeep gold cost with break-even economy analysis, desertion rules, deterministic combat pseudo-code with worked example, full tuning-knob catalog (§10).
  - Created [agent_tasks.md](./agent_tasks.md): Kanban-style board with 12 PoC tasks (ALPHA-001..004, BETA-001..004, GAMMA-001..003, LEAD-001) + 2 OPS placeholders. Tasks ALPHA-001, ALPHA-002, ALPHA-003 flagged **READY TO DELEGATE**.
- [decision] Committed to Go backend over Node.js/TypeScript — rationale: goroutine concurrency for the 1 Hz tick loop and per-connection WS handlers.
  - See architecture.md §2.
- [decision] Committed to React + `react-konva` (Canvas) over SVG — rationale: smooth interpolation of in-flight hero positions at 30 fps without DOM thrash.
  - See architecture.md §2.2.
