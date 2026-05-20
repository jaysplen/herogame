# herogame

A multiplayer browser game combining **Heroes of Might & Magic** (node/street maps, heroes, castles, armies) with **OGame**-style real-time movement timers. The project is in **PoC v0.1**: validating the core loop (move → arrive → fight) before adding skill trees, gear, or PvP.

## Stack at a glance

- **Backend:** Go 1.22+, chi, gorilla/websocket, sqlc, pgx
- **Database:** PostgreSQL 16 (durable state) + Redis 7 (arrival timers, sessions)
- **Frontend:** React 18, TypeScript, Vite, react-konva (Canvas), Zustand
- **Realtime:** WebSockets with JSON envelopes

See [docs/architecture.md](docs/architecture.md) for full system design.

## Repository map

```
herogame/
├── docs/          # LLM knowledge base (architecture, rules, tasks, changelog)
├── backend/       # Go game server (bootstrapped by ALPHA-001)
└── frontend/      # React client (bootstrapped by GAMMA-001)
```

## Where to start

**New contributors** — read in order:

1. [docs/architecture.md](docs/architecture.md) — tech stack, schema, WS protocol, PoC scope
2. [docs/game_rules.md](docs/game_rules.md) — movement math, upkeep, combat formulas
3. [docs/agent_tasks.md](docs/agent_tasks.md) — Kanban task board

**Executor agents** — pick a task marked **READY TO DELEGATE** in [docs/agent_tasks.md](docs/agent_tasks.md). Update [docs/changelog.md](docs/changelog.md) when you close a task.

## Status

**PoC v0.1** — documentation locked; backend and frontend code not yet implemented.
