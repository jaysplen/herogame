# PoC v0.1 — End-to-End Review

> **Task:** LEAD-001  
> **Date:** 2026-05-20  
> **Scope:** Validate the core loop (connect → buy → move → fight) and capture balance/tech follow-ups.

---

## 1. Executive Summary

The PoC **delivers the intended loop** on the backend: WebSocket handshake, unit purchase, authoritative movement with Redis-scheduled arrivals, deterministic combat at the Bandit Camp, and broadcasts (`move.update`, `combat.resolved`, `castle.tick`, `hero.state`). The React client renders the map, HUD, and combat modal.

**Balance lesson (by design):** 10 Pikemen lose decisively to the 50-stack Bandit Camp; the scripted smoke path typically ends in **defeat**, respawn at the castle, and a 60s lockout. Victory requires a much larger army (see §6).

---

## 2. Scripted Playthrough

### 2.1 Manual (browser)

Prerequisites: `make dev`, backend on `:8080`, `cd frontend && npm run dev`.

| Step | Action | Expected |
|------|--------|----------|
| 1 | Open http://localhost:5173 | `connected`, map + HUD |
| 2 | Buy **+10** Pikemen (needs 500g — wait for gold ticks or buy +1 four times from 200g start) | Army 10, gold debited, `hero.state` |
| 3 | Click **Crossroads** (node 2) | `move.update`, hero animates ~2s |
| 4 | Click **Bandit Camp** (node 5) | Longer trip (~3s with army 10), then **combat modal** |
| 5 | Outcome | **Loss** typical at 10 units; **Win** if army large enough (+500 gold, creep dead) |
| 6 | After loss | Hero at castle (node 1), respawn badge ~60s, move blocked |

### 2.2 Automated (`go test`)

```bash
make dev
cd backend && go test ./internal/ws -run TestPoCPlaythroughSmoke -count=1 -v
cd backend && go test ./internal/ws -run TestArmySlowdownObservable -count=1 -v
```

`TestPoCPlaythroughSmoke` resets world state, grants 500 gold, buys 10 Pikemen, moves 1→2→5, asserts `combat.resolved`, `combat_logs` row, and DB/Redis consistency for win or loss.

---

## 3. Reconciliation (UI vs DB)

| Field | Server source | Client source | Notes |
|-------|---------------|---------------|-------|
| Gold | `players.gold` | `castle.tick` anchor + extrapolation | Snaps on tick/buy/combat; integer floor in HUD |
| Army size | `SUM(hero_units.qty)` | `hero.state.armySize` | PoC single unit type; panel shows Pikeman × qty |
| Combat outcome | `combat_logs.outcome` | `combat.resolved` | Modal log matches `combat_logs.log` JSONB |
| Hero node | `heroes.current_node_id` | `hero.state` / map token | Updates on `move.arrived` |
| Travel time | `movement_orders.arrive_at - depart_at` | `move.update` + `useServerNow()` | Client animation should land within ±0.5s of `arriveAt` |

**Gap:** `hello.ack` does not yet include in-flight `movement_order` for reconnect (see BACKLOG-002).

---

## 4. Army Slowdown (Observable)

Formula: `travel_seconds = ceil((distance / base_speed) * upkeepSlowdown(army))` ([game_rules.md §3](./game_rules.md#3-movement)).

| Edge | Army | `TravelSeconds` (base speed 10) |
|------|------|-----------------------------------|
| Crossroads → Bandit Camp (30) | 10 | **3** |
| Same | 200 | **16** (~5.3×) |

`TestArmySlowdownObservable` enforces ratio > 4×. On the map, a 200-stack march is visibly slower than a 10-stack on the same edge.

---

## 5. Balance Findings

1. **Bandit Camp is not beatable with 10 Pikemen** — §6.6 worked example (30 vs 50) predicts loss; 10 units fare worse.
2. **Starting gold (200)** only buys **4** Pikemen at once; the LEAD script assumes 10 — players must wait for castle income (~60 g/min) or buy in batches.
3. **Straight path 1→2→5** is the fastest route to the creep; North/South forks add distance but may matter post-PoC.
4. **Economy vs combat pacing** — ~8 minutes of passive income to afford 10 units from scratch; combat reward (+500) is meaningful only after a winnable fight.
5. **No creep respawn** — manual DB reset required to re-test combat on node 5.

---

## 6. What Works Well

- Deterministic combat with full round logs (testable, debuggable).
- Import-cycle-free arrivals + tick separation.
- Server-time movement interpolation on the client.
- Transactional arrival + combat in one DB tx.

---

## 7. Follow-Up Backlog (filed in `agent_tasks.md`)

| ID | Title | Priority |
|----|-------|----------|
| BACKLOG-001 | OPS-002: GitHub Actions CI (go test + frontend build) | High — **Done** |
| BACKLOG-002 | Replay in-flight movement in `hello.ack` | Medium |
| BACKLOG-003 | Server `respawnUntil` in `hero.state` (drop client-only timer) | Medium |
| BACKLOG-004 | `hero_units[]` in bootstrap / `hero.state` for multi-unit HUD | Medium |
| BACKLOG-005 | Balance pass: starting gold or Bandit qty for teachable first win | Medium |
| BACKLOG-006 | `scripts/playthrough.sh` + Playwright visual smoke | Low |

---

## 8. Recommendation

**PoC v0.1 is complete** for the stated goal: prove the move → arrive → fight loop with a multiplayer-ready server and a playable client shell. Before calling it “player-ready,” ship **BACKLOG-001** (CI) and **BACKLOG-005** (first-win tuning or onboarding hint in UI).
