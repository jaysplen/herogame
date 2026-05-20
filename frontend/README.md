# Frontend (React + TypeScript)

Browser client: Konva node map, WebSocket client, HUD (gold, army, combat log).

## Intended layout

Per [docs/architecture.md §3](../docs/architecture.md#3-repository-layout):

```
frontend/
├── package.json
├── vite.config.ts
└── src/
    ├── main.tsx
    ├── App.tsx
    ├── net/ws.ts           # WebSocket client
    ├── state/store.ts      # Zustand
    ├── map/                # react-konva map
    ├── hud/                # gold, army, hero panel
    └── proto/              # TS types mirroring backend/internal/proto
```

## Bootstrap

This directory is **not implemented yet**. It is created by task **[GAMMA-001]** in [docs/agent_tasks.md](../docs/agent_tasks.md):

- Vite + React 18 + TypeScript
- WS client + Zustand store
- `hello` handshake and bootstrap snapshot display

Depends on **BETA-001** (WebSocket gateway) for a live server to connect to.
