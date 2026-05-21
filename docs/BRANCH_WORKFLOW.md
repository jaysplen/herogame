# Branch workflow: epic single-player world

Integration branch: `feature/epic-singleplayer-world`

## Milestone branches

| Branch | Scope |
|--------|--------|
| `m1-respawn-map` | Dead-click alerts, expanded map seed |
| `m2-roaming-creeps` | Moving neutrals, path collision |
| `m3-resources` | Multi-resource economy, capturable nodes |
| `m4-castle-progression` | Factions, tiers, build menu |
| `m5-tactical-rules` | Grace period, garrison defense |
| `m6-combat-objective` | Hover intel, conversion rewards, 5-kill win |
| `m7-multiplayer-seams` | Per-player broadcast filtering |

Merge each `m*` into `feature/epic-singleplayer-world` when tests pass, then open a PR to `master`.

## Local setup

```bash
make dev
make server          # terminal 1
make frontend        # terminal 2
```

## Tests

```bash
cd backend && go test ./... -count=1
cd frontend && npm run test:e2e
```

## Multiplayer later

Keep commands player-scoped (`hero.PlayerID == c.playerID`). Use `Hub.BroadcastToPlayer` for private state; global broadcast only for shared world (creeps, resource node ownership).
