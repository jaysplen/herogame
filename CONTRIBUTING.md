# Contributing to herogame

Thanks for joining. This repo is set up for **human + AI agent** teams: docs are the source of truth for scope, rules, and task ownership.

## Quick start (first run)

**Prerequisites:** Docker, Go 1.22+, Node 20+, [goose](https://github.com/pressly/goose) (`go install github.com/pressly/goose/v3/cmd/goose@latest`), and `$(go env GOPATH)/bin` on your `PATH`.

```bash
git clone https://github.com/jaysplen/herogame.git
cd herogame
cp .env.example .env          # optional; defaults match docker-compose
make dev                      # Postgres 16 + Redis 7 + migrations (no sudo)
```

**Do not use `sudo make dev`** — root cannot see your user’s `goose`/`go` install. If Docker needs elevated rights, add your user to the `docker` group instead.

**Terminal 1 — game server (must have WebSocket):**

```bash
make server
# Expect log: "websocket gateway enabled"
```

**Terminal 2 — frontend:**

```bash
make frontend
# Open http://127.0.0.1:5173
```

Do **not** use a bare `/tmp/herogame-server` binary without `DATABASE_URL` — it breaks WebSocket (`/ws` returns 404). Always use `make server`.

Full details: [docs/dev_setup.md](docs/dev_setup.md).

## For AI agent teams

Read these **before** coding:

| Doc | Purpose |
|-----|---------|
| [docs/architecture.md](docs/architecture.md) | Stack, schema, WS protocol, server authority |
| [docs/game_rules.md](docs/game_rules.md) | Movement, combat, economy formulas |
| [docs/agent_tasks.md](docs/agent_tasks.md) | Kanban — pick tasks from Backlog only when deps are Done |
| [docs/changelog.md](docs/changelog.md) | Append one line when closing a task |
| [docs/authority.md](docs/authority.md) | What the server owns vs client display |
| [docs/BRANCH_WORKFLOW.md](docs/BRANCH_WORKFLOW.md) | Epic branch / milestone naming |

**Workflow:**

1. Pick a task in `docs/agent_tasks.md` → move to **In Progress**, set Owner.
2. Work on a feature branch (see [docs/BRANCH_WORKFLOW.md](docs/BRANCH_WORKFLOW.md)).
3. Run tests before PR:
   ```bash
   cd backend && go vet ./... && go test ./... -count=1
   cd frontend && npm ci && npm run build
   cd frontend && npm run test:e2e   # needs make dev + Playwright browsers
   ```
4. Open a PR to `master`; CI runs backend, frontend build, and Playwright smoke.
5. Move task to **Done** and update `docs/changelog.md`.

**Conventions:** Server-authoritative gameplay; minimal diffs; match existing naming; no secrets in git.

## Branches

| Branch | Role |
|--------|------|
| `master` | Stable integration |
| `feature/epic-singleplayer-world` | World expansion (map, creeps, resources, castles, objective) |

Create short-lived branches off `master` or the active feature branch: `mN-short-description`.

## Repo access

Ask the repo owner ([jaysplen](https://github.com/jaysplen)) to add you as a **collaborator** on GitHub (Settings → Collaborators), or fork and open PRs.

## Troubleshooting

| Symptom | Fix |
|---------|-----|
| `websocket error` in UI | `curl -w "%{http_code}" http://127.0.0.1:8080/ws` — must not be **404**. Run `make server`. |
| `goose not found` | `go install github.com/pressly/goose/v3/cmd/goose@latest` and add `~/go/bin` to PATH |
| Port 5173 unreachable | Start frontend: `cd frontend && npm run dev -- --host 0.0.0.0` or `make frontend` |
| WSL + Windows Chrome | Use `http://localhost:5173` or `VITE_WS_URL=ws://127.0.0.1:8080/ws` |

Reset stuck ports: `bash scripts/dev-restart.sh`
