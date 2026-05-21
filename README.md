# herogame

A multiplayer browser game combining **Heroes of Might & Magic** (node maps, heroes, castles, armies) with **OGame**-style real-time movement. **PoC v0.1** is complete; active work is on the **single-player world expansion** (larger map, roaming creeps, multi-resource economy, castle progression).

## Stack

- **Backend:** Go 1.22+, chi, WebSocket, sqlc, pgx
- **Database:** PostgreSQL 16 + Redis 7
- **Frontend:** React 18, TypeScript, Vite, react-konva, Zustand

See [docs/architecture.md](docs/architecture.md).

## Quick start

```bash
git clone https://github.com/jaysplen/herogame.git
cd herogame
make dev          # Postgres + Redis + migrations
make server       # terminal 1 — game server :8080
make frontend     # terminal 2 — Vite :5173 → http://127.0.0.1:5173
```

**Contributing / agent teams:** [CONTRIBUTING.md](CONTRIBUTING.md)

## Docs (read first)

| Doc | |
|-----|--|
| [docs/architecture.md](docs/architecture.md) | System design, WS protocol |
| [docs/game_rules.md](docs/game_rules.md) | Rules and formulas |
| [docs/agent_tasks.md](docs/agent_tasks.md) | Task board (Kanban) |
| [docs/dev_setup.md](docs/dev_setup.md) | Local setup, CI, Playwright |
| [docs/BRANCH_WORKFLOW.md](docs/BRANCH_WORKFLOW.md) | Feature branches |

## Repository layout

```
herogame/
├── backend/       # Go server, migrations, sqlc queries
├── frontend/      # React client + Playwright e2e
├── docs/          # Architecture, rules, tasks, changelog
├── scripts/       # dev-backend, dev-restart, e2e helpers
└── .github/       # CI (go test, frontend build, Playwright)
```

## Tests

```bash
cd backend && go test ./... -count=1
cd frontend && npm ci && npm run build && npm run test:e2e
```

## Status

- **master** — PoC loop + CI + Playwright smoke
- **feature/epic-singleplayer-world** — 16-node map, roaming creeps, resources, castle build/factions, 5-kill objective

Review: [docs/poc_review.md](docs/poc_review.md)
