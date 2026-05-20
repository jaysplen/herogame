# Server authority (PoC)

> All gameplay timing and outcomes are computed on the server. The client is a renderer and command sender.

## Server-owned

| Concern | Where |
|--------|--------|
| Travel duration | `world.TravelSeconds` in `handlers_move.go` |
| `departAt` / `arriveAt` | Set on `INSERT movement_orders`; replayed in `hello.ack.inFlight` |
| Arrival resolution | `arrivals.Scheduler` + tick engine (Redis ZSET) |
| Combat | `combat.Resolve` + `combat.ApplyAtNode` in arrival transaction |
| Gold income / upkeep | `tick/economy.go`, `tick/upkeep.go` |
| Respawn lockout | Redis + `hero.state.respawnUntil` |
| Hero position after events | `heroes.current_node_id` → `move.arrived` + `hero.state` |

## Client-owned (display only)

| Concern | Behavior |
|--------|----------|
| Hero token along edge | Interpolate using server `departAt`/`arriveAt` and skew-corrected `serverTime` from inbound envelopes |
| Gold HUD | Estimate between `castle.tick` snapshots; snaps on each authoritative tick |
| Adjacent-node highlight | UX only; `move.request` is validated again on the server |
| Outbound `serverTime` | Always `0`; ignored by the server |

## Inbound rule

Every **server → client** envelope sets `serverTime` via `proto.NewEnvelope` (wall clock on the game server). The client updates `clockSkewMs` from that field only.
