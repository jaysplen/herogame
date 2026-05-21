# Future features

After a sweep of the codebase, three features stand out as the highest-leverage
next steps. Each fits the existing server-authoritative architecture, uses
data shapes already in place, and meaningfully deepens HoMM-style gameplay.

The ordering is by impact-per-effort: **#1 first** delivers the biggest jump in
strategic depth for the least new surface area; **#3 last** is the largest
backend lift but unlocks several follow-ons.

---

## 1. Multi-unit roster with rock-paper-scissors counters

### Problem
`frontend/src/hud/constants.ts` admits it: *"single unit type for now"*. Every
army is just a stack of Pikemen. Combat reduces to "who has more bodies", and
the Garrison panel has nothing interesting to display. `army_test.go` already
calls into a stack-shaped army, and `units` is plural everywhere in the data
model — the schema is ready, the gameplay isn't.

### Why it matters
A three-unit triangle (Pike → Cavalry → Archer → Pike) creates real composition
decisions every recruitment cycle. It also gives the Bandit Camp / Wolf Den
node variants a reason to feel distinct: each could field its own creep
composition that the player must counter.

### Proposed roster (MVP)

| ID | Name      | Cost (g) | Atk | Def | HP | Speed | Counters → | Notes |
|----|-----------|----------|-----|-----|----|-------|------------|-------|
| 1  | Pikeman   | 50       | 4   | 5   | 12 | 1.0   | Cavalry    | Existing baseline |
| 2  | Archer    | 70       | 6   | 2   | 8  | 1.0   | Pikeman    | Ranged: deals dmg in round 1 before melee lands |
| 3  | Cavalry   | 110      | 7   | 4   | 14 | 1.5   | Archer     | Strikes first in melee, fragile to spears |

Counter bonus: +25% damage when attacking the unit listed in `Counters →`.

### Schema changes
- `units` table — already supports rows for new IDs. Seed in a new migration
  `00006_seed_units.sql`.
- `combat_log` — already keyed by unit, no schema change needed.

### Server changes
- `internal/combat/resolve.go` — split a single combat round into two phases
  (ranged → melee), apply the counter multiplier when attacker.unit_id beats
  defender.unit_id.
- `internal/ws/units.go` / `handlers_buy.go` — already iterates `shopUnits`,
  just needs the new rows.

### Client changes
- `ArmyPanel.tsx` — already renders `stacks` and `recruitOnly` per unit. The
  only addition is per-unit icons (we can reuse the existing illustration
  system from `map/Node.tsx`).
- `CombatModal.tsx` — show unit IDs in the log table.

### MVP slice
Migration + counter math + two extra units (Archer, Cavalry). Skip ranged-phase
complexity for v1: just apply the counter multiplier in the existing single-
round loop.

**Estimated effort:** 1–2 days; ~250 lines server, ~80 lines client, 1 migration.

---

## 2. Fog of war with line-of-sight reveal

### Problem
`hello.ack` ships the full `map` snapshot to every client — including the
enemy castle, every creep on the map, and every resource node's owner. The
player can see everything before stepping out of Ironkeep. HoMM is famously
about exploration; we have none.

### Why it matters
Beyond fixing an obvious leak, fog of war:
- Gives the watercolor map a reason to feel large — unexplored regions sit
  under a parchment haze, revealing as the hero scouts.
- Creates raid-or-defend asymmetry: enemy castle activity is invisible until
  a scout is in range.
- Makes the Wolf Den and Bandit Camp creep movements (already implemented in
  `internal/tick/creeps.go`) tactically relevant — you only see what's nearby.

### Mechanic
Each player has a set of *revealed* node IDs. A node becomes revealed when the
hero visits it; it stays revealed forever (classic HoMM "explored fog"), but
*entity state* (creeps, enemy hero, resource ownership) on that node is only
*currently visible* when the hero stands within 1 edge.

### Schema changes
- New table `player_visibility (player_id, node_id, first_seen_at)` — written
  on arrival in `internal/arrivals/arrivals.go`.

### Server changes
- `internal/ws/bootstrap.go` — filter `map.nodes` and `map.edges` to those
  reachable through revealed nodes; include a `visibility` payload listing
  which of those are "currently visible".
- `internal/ws/broadcast.go` — strip creeps/resource owners/enemy heroes
  whose node isn't currently visible to the recipient.
- `internal/ws/handlers_move.go` — on every arrival, INSERT into
  `player_visibility` and push a `visibility.update` envelope.

### Client changes
- `state/store.ts` — add `visibility: { revealed: Set<number>, current: Set<number> }`.
- `map/MapBackground.tsx` — overlay a parchment-blur layer masked by the
  revealed set; render fully visible only inside the current set.
- `map/Node.tsx` — render unrevealed nodes as a vague silhouette (just the
  base disc, no icon, no label).

### Risk
- Pathfinding hints (which nodes you can move to) must remain consistent —
  the server is the source of truth, but the client must not show "click to
  travel" affordances for nodes the player hasn't seen yet.
- E2E test `smoke.spec.ts` relies on `e2e-move-N` testids being present;
  gating them behind visibility breaks the test. Solution: a `VITE_E2E=1`
  bypass that reveals the full map for tests.

### MVP slice
Just revealed/unrevealed (no "currently visible" sub-distinction). All
revealed nodes show full state; unrevealed ones show nothing. Adds the
exploration loop without the dynamic-occlusion complexity.

**Estimated effort:** 3–4 days; ~400 lines server, ~150 lines client,
1 migration, 1 new WS envelope.

---

## 3. Hero leveling and skill tree

### Problem
The hero has `speedEffective`, `attack`, `defense`, `upkeepGoldPerHour` — and
no way to change any of them. After your first castle build-out, hero
identity stops mattering. The "5-kill objective" is just a checkbox, not a
progression engine.

### Why it matters
HoMM heroes have classes, specialties, and a skill tree. Adding even a
minimal version turns the existing kill-counter into XP, and gives every
victory a meaningful drip of choice.

### Proposed system (MVP)

**XP & levels** — every combat win (vs creep or hero) grants XP proportional
to enemy power. Level up at 100, 250, 500, 1000 XP. Each level grants 1
skill point.

**Three skill branches**, each with three tiers:

| Branch    | T1 (1 pt)             | T2 (3 pts)                | T3 (5 pts)                       |
|-----------|-----------------------|---------------------------|----------------------------------|
| Logistics | +15% hero speed       | -25% travel time on roads | Reveal adjacent nodes for free   |
| War       | +1 attack to all units | +10% combat damage        | Cavalry +30% counter bonus       |
| Economy   | -25% upkeep            | +20% gold/min in your castle | +1 resource/min from owned nodes |

A hero can dabble across branches or commit to one — same shape as HoMM's
secondary-skill tree, just simpler.

### Schema changes
- `heroes` table — add `xp INT NOT NULL DEFAULT 0`,
  `level INT NOT NULL DEFAULT 1`, `unspent_skill_points INT NOT NULL DEFAULT 0`.
- New table `hero_skills (hero_id, branch, tier)` — composite primary key.

### Server changes
- `internal/combat/apply.go` — on combat win, compute XP gain and award
  it; promote level if threshold crossed.
- New `internal/hero/skills.go` — handle skill purchase requests, validate
  tier prerequisites and available points.
- New WS handler `MsgHeroLearnSkill` — branch + tier; idempotent.
- Update `internal/economy/upkeep.go`, `internal/world/distance.go`, and
  `internal/combat/resolve.go` to apply each unlocked perk's modifier.

### Client changes
- New `HeroSkillsPanel.tsx` — show XP bar, unspent points, three-branch
  tree with click-to-buy buttons. Disable buttons when prerequisites unmet.
- `HeroPanel.tsx` — extend `hero-stats` block with Level and XP.

### MVP slice
Ship Logistics branch only (3 perks, easy to test: speed, road bonus,
reveal). Defer War and Economy branches until balance is dialed in.

**Estimated effort:** 4–6 days; ~700 lines server (combat-XP hookup is the
big rock), ~300 lines client, 2 migrations, 1 new envelope.

---

## Summary

| # | Feature                           | Effort | Schema | New envelopes | Strategic depth gained |
|---|-----------------------------------|--------|--------|---------------|------------------------|
| 1 | Multi-unit roster + counters      | 1–2d   | 1      | 0             | Composition matters every recruit cycle |
| 2 | Fog of war (revealed-only MVP)    | 3–4d   | 1      | 1             | Exploration loop; map asymmetry |
| 3 | Hero leveling + Logistics skills  | 4–6d   | 2      | 1             | Per-hero progression; reward variety |

**Recommended sequence:** ship #1 first (small, high-impact, no new client UX
surface). Then #2 (visual + combat depth, leans on the new map). Save #3 for
when the moment-to-moment combat is interesting enough that "more
permanently" matters.
