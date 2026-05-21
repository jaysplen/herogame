# Frontend (React + TypeScript)

Browser client: WebSocket + Zustand (GAMMA-001), Konva map (GAMMA-002), HUD (GAMMA-003), strategy-loop expansion (EPIC-ROADMAP-001).

**HUD:** Multi-resource wallet (gold, metal, gems, coal, wood, stone), unit recruiting by tier/faction costs, castle build actions, hero respawn/grace messaging, combat modal with converted-unit rewards.

Click a **yellow-highlighted** adjacent node to move. While traveling, other clicks show “Hero is moving”.

Map now includes:
- larger node graph with castles, creep zones, and resource nodes
- roaming neutral creeps with live map movement
- node tooltip intel (enemy stack stats + estimated win chance)
- objective banner (eliminate enemy hero 5x)

## Run

```bash
# From repo root (backend must be running on :8080)
cd frontend && npm install && npm run dev
```

Open http://localhost:5173 — connects to `ws://localhost:8080/ws` as Player 1.
Optional player swap: `http://localhost:5173/?playerId=2`.

Override WS URL: `VITE_WS_URL=ws://localhost:8080/ws npm run dev`

## E2E smoke (BACKLOG-006)

With `make dev` running (Postgres + Redis):

```bash
npm ci
npx playwright install chromium
npm run test:e2e
```

Playwright boots the game server on `:18080` and Vite with `VITE_E2E=1` (DOM move/build buttons, no canvas clicks). Flow: connect → wait grace → buy army → move → combat modal.

## Layout

```
src/
├── main.tsx
├── App.tsx
├── net/ws.ts           # WebSocket client
├── state/store.ts      # Zustand + useServerNow()
├── map/                # react-konva map (GAMMA-002)
├── hud/                # gold, army, hero, combat modal (GAMMA-003)
└── proto/              # TS types mirroring backend/internal/proto
```
