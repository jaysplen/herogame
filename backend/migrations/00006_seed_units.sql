-- +goose Up
-- Multi-unit roster + rock-paper-scissors counters (docs/future_features.md §1).
-- Triangle: Pikeman (id=1) → Cavalry → Archer → Pikeman.
-- Counter math (+25%) lives in code (internal/combat/resolve.go); this migration
-- only seeds the catalog rows so the recruit UI and combat resolver see them.

-- Rebalance baseline Pikeman to the MVP stat line (Atk 4 / Def 5 / HP 12).
UPDATE units
SET attack = 4, defense = 5, hp = 12, cost_gold = 50, upkeep_gold_per_hour = 2.0
WHERE id = 1;

-- Archer (id=6): ranged baseline, fragile, counters Pikemen.
-- Cavalry (id=7): heavy striker, counters Archers, weak to spears.
-- Both seeded as `neutral` so they show up in every castle's recruit shop
-- regardless of faction; tier=1 so they're available from the starting barracks.
INSERT INTO units (
    id, code, name, cost_gold, attack, defense, hp, upkeep_gold_per_hour,
    cost_metal, cost_gems, cost_coal, cost_wood, cost_stone, faction, tier
) VALUES
    (6, 'archer',  'Archer',  70,  6, 2,  8, 2.8, 0, 0, 0, 6, 0, 'neutral', 1),
    (7, 'cavalry', 'Cavalry', 110, 7, 4, 14, 4.4, 8, 0, 0, 0, 4, 'neutral', 1)
ON CONFLICT (id) DO UPDATE SET
    code = EXCLUDED.code,
    name = EXCLUDED.name,
    cost_gold = EXCLUDED.cost_gold,
    attack = EXCLUDED.attack,
    defense = EXCLUDED.defense,
    hp = EXCLUDED.hp,
    upkeep_gold_per_hour = EXCLUDED.upkeep_gold_per_hour,
    cost_metal = EXCLUDED.cost_metal,
    cost_gems = EXCLUDED.cost_gems,
    cost_coal = EXCLUDED.cost_coal,
    cost_wood = EXCLUDED.cost_wood,
    cost_stone = EXCLUDED.cost_stone,
    faction = EXCLUDED.faction,
    tier = EXCLUDED.tier;

SELECT setval(pg_get_serial_sequence('units', 'id'), (SELECT MAX(id) FROM units));

-- +goose Down
DELETE FROM units WHERE id IN (6, 7);

UPDATE units
SET attack = 3, defense = 2, hp = 10, cost_gold = 50, upkeep_gold_per_hour = 2.0
WHERE id = 1;
