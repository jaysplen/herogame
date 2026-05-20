# Frontend (React + TypeScript)

Browser client: WebSocket + Zustand (GAMMA-001), Konva map (GAMMA-002), HUD (GAMMA-003).

## Run

```bash
# From repo root (backend must be running on :8080)
cd frontend && npm install && npm run dev
```

Open http://localhost:5173 — connects to `ws://localhost:8080/ws` as Player 1.

Override WS URL: `VITE_WS_URL=ws://localhost:8080/ws npm run dev`

## Layout

```
src/
├── main.tsx
├── App.tsx
├── net/ws.ts           # WebSocket client
├── state/store.ts      # Zustand + useServerNow()
└── proto/              # TS types mirroring backend/internal/proto
```
