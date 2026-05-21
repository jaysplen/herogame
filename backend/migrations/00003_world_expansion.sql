-- +goose Up
ALTER TABLE players
    ADD COLUMN metal NUMERIC(14, 4) NOT NULL DEFAULT 0,
    ADD COLUMN gems NUMERIC(14, 4) NOT NULL DEFAULT 0,
    ADD COLUMN coal NUMERIC(14, 4) NOT NULL DEFAULT 0,
    ADD COLUMN wood NUMERIC(14, 4) NOT NULL DEFAULT 0,
    ADD COLUMN stone NUMERIC(14, 4) NOT NULL DEFAULT 0;

ALTER TABLE units
    ADD COLUMN cost_metal INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN cost_gems INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN cost_coal INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN cost_wood INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN cost_stone INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN faction TEXT NOT NULL DEFAULT 'neutral',
    ADD COLUMN tier INTEGER NOT NULL DEFAULT 1;

ALTER TABLE castles
    ADD COLUMN faction TEXT NOT NULL DEFAULT 'castle',
    ADD COLUMN defense_bonus INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN barracks_tier INTEGER NOT NULL DEFAULT 1,
    ADD COLUMN academy_tier INTEGER NOT NULL DEFAULT 1;

ALTER TABLE neutral_creeps
    ADD COLUMN from_node_id BIGINT REFERENCES map_nodes (id),
    ADD COLUMN to_node_id BIGINT REFERENCES map_nodes (id),
    ADD COLUMN depart_at TIMESTAMPTZ,
    ADD COLUMN arrive_at TIMESTAMPTZ,
    ADD COLUMN speed_units INTEGER NOT NULL DEFAULT 4,
    ADD COLUMN attack INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN defense INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN hp INTEGER NOT NULL DEFAULT 0,
    ADD COLUMN grace_until TIMESTAMPTZ;

ALTER TABLE heroes
    ADD COLUMN spawn_grace_until TIMESTAMPTZ;

ALTER TABLE combat_logs
    ADD COLUMN enemy_hero_id BIGINT REFERENCES heroes (id),
    ADD COLUMN converted_units INTEGER NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS resource_nodes (
    id BIGSERIAL PRIMARY KEY,
    node_id BIGINT NOT NULL UNIQUE REFERENCES map_nodes (id),
    resource_type TEXT NOT NULL CHECK (resource_type IN ('gold', 'metal', 'gems', 'coal', 'wood', 'stone')),
    per_min INTEGER NOT NULL,
    owner_player_id BIGINT REFERENCES players (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS player_kills (
    killer_player_id BIGINT NOT NULL REFERENCES players (id),
    victim_player_id BIGINT NOT NULL REFERENCES players (id),
    kills INTEGER NOT NULL DEFAULT 0,
    PRIMARY KEY (killer_player_id, victim_player_id)
);

CREATE TABLE IF NOT EXISTS castle_buildings (
    id BIGSERIAL PRIMARY KEY,
    castle_id BIGINT NOT NULL REFERENCES castles (id) ON DELETE CASCADE,
    building_code TEXT NOT NULL,
    level INTEGER NOT NULL DEFAULT 1,
    UNIQUE (castle_id, building_code)
);

-- Reset static world and reseed larger graph.
DELETE FROM movement_orders;
DELETE FROM hero_units;
DELETE FROM combat_logs;
DELETE FROM neutral_creeps;
DELETE FROM castles;
DELETE FROM heroes;
DELETE FROM resource_nodes;
DELETE FROM map_edges;
DELETE FROM map_nodes;

INSERT INTO map_nodes (id, name, x, y, kind) VALUES
    (1, 'Ironkeep Castle', 120, 360, 'castle'),
    (2, 'Moss Crossing', 250, 360, 'wild'),
    (3, 'North Forest', 250, 220, 'wild'),
    (4, 'South Quarry', 250, 500, 'wild'),
    (5, 'Bandit Camp', 390, 360, 'creep'),
    (6, 'Sunspire Castle', 760, 360, 'castle'),
    (7, 'Gem Caves', 390, 220, 'wild'),
    (8, 'Coal Pit', 390, 500, 'wild'),
    (9, 'Ruined Watch', 520, 260, 'wild'),
    (10, 'Old Lumber', 520, 460, 'wild'),
    (11, 'Stone Ridge', 640, 260, 'wild'),
    (12, 'Golden Field', 640, 460, 'wild'),
    (13, 'East Pass', 760, 240, 'wild'),
    (14, 'South Gate', 760, 480, 'wild'),
    (15, 'Wolf Den', 520, 360, 'creep'),
    (16, 'Mercury Marsh', 640, 360, 'wild');

SELECT setval(pg_get_serial_sequence('map_nodes', 'id'), (SELECT MAX(id) FROM map_nodes));

INSERT INTO map_edges (from_node_id, to_node_id, distance_units) VALUES
    (1,2,14),(2,1,14),
    (2,3,12),(3,2,12),
    (2,4,12),(4,2,12),
    (2,5,16),(5,2,16),
    (3,7,10),(7,3,10),
    (4,8,10),(8,4,10),
    (5,7,10),(7,5,10),
    (5,8,10),(8,5,10),
    (5,9,12),(9,5,12),
    (5,10,12),(10,5,12),
    (9,11,10),(11,9,10),
    (10,12,10),(12,10,10),
    (9,15,8),(15,9,8),
    (10,15,8),(15,10,8),
    (15,16,12),(16,15,12),
    (11,16,10),(16,11,10),
    (12,16,10),(16,12,10),
    (11,13,10),(13,11,10),
    (12,14,10),(14,12,10),
    (13,6,10),(6,13,10),
    (14,6,10),(6,14,10),
    (16,6,12),(6,16,12),
    (9,16,12),(16,9,12),
    (10,16,12),(16,10,12);

INSERT INTO heroes (id, player_id, name, current_node_id, base_speed, attack, defense, spawn_grace_until) VALUES
    (1, 1, 'Hero of Player1', 1, 10, 2, 2, now() + interval '12 seconds'),
    (2, 2, 'Hero of Player2', 6, 10, 2, 2, now() + interval '12 seconds');

SELECT setval(pg_get_serial_sequence('heroes', 'id'), (SELECT GREATEST(MAX(id), 2) FROM heroes));

UPDATE players
SET gold = 200, metal = 60, gems = 20, coal = 60, wood = 80, stone = 80
WHERE id IN (1, 2);

INSERT INTO castles (player_id, node_id, gold_per_min, faction, defense_bonus, barracks_tier, academy_tier) VALUES
    (1, 1, 65, 'knight', 3, 1, 1),
    (2, 6, 65, 'warlock', 3, 2, 1);

INSERT INTO neutral_creeps (
    node_id, name, unit_id, qty, alive, gold_reward,
    from_node_id, to_node_id, depart_at, arrive_at, speed_units, attack, defense, hp, grace_until
) VALUES
    (5, 'Bandit Camp', 1, 40, TRUE, 420, 5, 9, now(), now() + interval '30 seconds', 4, 3, 2, 10, now() + interval '12 seconds'),
    (15, 'Wolf Den', 1, 28, TRUE, 280, 15, 10, now(), now() + interval '26 seconds', 4, 4, 2, 9, now() + interval '12 seconds');

INSERT INTO resource_nodes (node_id, resource_type, per_min, owner_player_id) VALUES
    (3, 'wood', 24, NULL),
    (4, 'stone', 24, NULL),
    (7, 'gems', 8, NULL),
    (8, 'coal', 20, NULL),
    (11, 'metal', 18, NULL),
    (12, 'gold', 30, NULL),
    (16, 'gold', 20, NULL);

INSERT INTO castle_buildings (castle_id, building_code, level)
SELECT c.id, 'defense', CASE WHEN c.player_id = 1 THEN 1 ELSE 2 END
FROM castles c
UNION ALL
SELECT c.id, 'barracks', c.barracks_tier
FROM castles c
UNION ALL
SELECT c.id, 'academy', c.academy_tier
FROM castles c;

-- Add faction/tiered unit options for progression.
INSERT INTO units (
    id, code, name, cost_gold, attack, defense, hp, upkeep_gold_per_hour,
    cost_metal, cost_gems, cost_coal, cost_wood, cost_stone, faction, tier
) VALUES
    (2, 'crossbow', 'Crossbowman', 90, 5, 3, 14, 3.4, 10, 0, 0, 8, 4, 'knight', 2),
    (3, 'griffin', 'Griffin', 180, 9, 7, 28, 6.8, 20, 3, 6, 0, 8, 'knight', 3),
    (4, 'gargoyle', 'Gargoyle', 85, 5, 4, 16, 3.3, 8, 0, 4, 0, 8, 'warlock', 2),
    (5, 'mage', 'Mage', 170, 10, 5, 20, 6.5, 12, 4, 6, 0, 0, 'warlock', 3)
ON CONFLICT (id) DO NOTHING;

SELECT setval(pg_get_serial_sequence('units', 'id'), (SELECT MAX(id) FROM units));

UPDATE units
SET
    cost_metal = 0,
    cost_gems = 0,
    cost_coal = 0,
    cost_wood = 0,
    cost_stone = 0,
    faction = 'neutral',
    tier = 1
WHERE id = 1;

-- +goose Down
DELETE FROM castle_buildings;
DELETE FROM player_kills;
DELETE FROM resource_nodes;

ALTER TABLE combat_logs
    DROP COLUMN IF EXISTS converted_units,
    DROP COLUMN IF EXISTS enemy_hero_id;

ALTER TABLE heroes
    DROP COLUMN IF EXISTS spawn_grace_until;

ALTER TABLE neutral_creeps
    DROP COLUMN IF EXISTS grace_until,
    DROP COLUMN IF EXISTS hp,
    DROP COLUMN IF EXISTS defense,
    DROP COLUMN IF EXISTS attack,
    DROP COLUMN IF EXISTS speed_units,
    DROP COLUMN IF EXISTS arrive_at,
    DROP COLUMN IF EXISTS depart_at,
    DROP COLUMN IF EXISTS to_node_id,
    DROP COLUMN IF EXISTS from_node_id;

ALTER TABLE castles
    DROP COLUMN IF EXISTS academy_tier,
    DROP COLUMN IF EXISTS barracks_tier,
    DROP COLUMN IF EXISTS defense_bonus,
    DROP COLUMN IF EXISTS faction;

ALTER TABLE units
    DROP COLUMN IF EXISTS tier,
    DROP COLUMN IF EXISTS faction,
    DROP COLUMN IF EXISTS cost_stone,
    DROP COLUMN IF EXISTS cost_wood,
    DROP COLUMN IF EXISTS cost_coal,
    DROP COLUMN IF EXISTS cost_gems,
    DROP COLUMN IF EXISTS cost_metal;

ALTER TABLE players
    DROP COLUMN IF EXISTS stone,
    DROP COLUMN IF EXISTS wood,
    DROP COLUMN IF EXISTS coal,
    DROP COLUMN IF EXISTS gems,
    DROP COLUMN IF EXISTS metal;

DROP TABLE IF EXISTS castle_buildings;
DROP TABLE IF EXISTS player_kills;
DROP TABLE IF EXISTS resource_nodes;
