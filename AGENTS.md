# Agent instructions (herogame)

This file is for **Cursor / Codex / other coding agents** working in this repository.

> **Claude Code / Cowork users:** see also `CLAUDE.md` for sandbox-specific
> rules (Windows mount write reliability, git filemode noise, stale
> `.git/index.lock`, push auth).

## Read first

1. [docs/architecture.md](docs/architecture.md) — authority model and WS envelopes
2. [docs/game_rules.md](docs/game_rules.md) — gameplay math
3. [docs/agent_tasks.md](docs/agent_tasks.md) — only implement tasks listed; update kanban + changelog when done
4. [CLAUDE.md](CLAUDE.md) — sandbox + tooling reliability rules

## Commands

```bash
make dev && make server    # backend
make frontend              # http://127.0.0.1:5173
cd backend && go test ./... -count=1
cd frontend && npm run test:e2e
```

Use `make server` (not ad-hoc binaries without `DATABASE_URL`).

## Rules

- Server owns: travel time, arrivals, combat, economy ticks, hero position, resources.
- Client: interpolate movement, display-only gold estimate between `castle.tick`.
- Minimal diffs; match existing patterns; no commits unless the user asks.
- Branch workflow: [docs/BRANCH_WORKFLOW.md](docs/BRANCH_WORKFLOW.md)
- When working through a sandbox on a Windows-mounted folder, **use bash
  heredoc for any file write over ~50 lines** (see CLAUDE.md §1) — the
  `Write`/`Edit` tools have been observed to silently truncate large files
  on that mount.

Human onboarding: [CONTRIBUTING.md](CONTRIBUTING.md)
