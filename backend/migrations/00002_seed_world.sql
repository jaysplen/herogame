-- +goose Up
-- Map nodes (fixed IDs for PoC graph — architecture.md §9.2)
INSERT INTO map_nodes (id, name, x, y, kind) VALUES
    (1, 'Player1 Castle', 100, 300, 'castle'),
    (2, 'Crossroads', 300, 300, 'wild'),
    (3, 'North Fork', 300, 100, 'wild'),
    (4, 'South Fork', 300, 500, 'wild'),
    (5, 'Bandit Camp', 500, 300, 'creep'),
    (6, 'Player2 Castle', 700, 300, 'castle');

SELECT setval(pg_get_serial_sequence('map_nodes', 'id'), (SELECT MAX(id) FROM map_nodes));

-- Bidirectional edges (7 logical connections → 14 rows)
INSERT INTO map_edges (from_node_id, to_node_id, distance_units) VALUES
    (1, 2, 20), (2, 1, 20),
    (2, 3, 15), (3, 2, 15),
    (2, 4, 15), (4, 2, 15),
    (2, 5, 30), (5, 2, 30),
    (3, 5, 35), (5, 3, 35),
    (4, 5, 35), (5, 4, 35),
    (5, 6, 20), (6, 5, 20);

-- Unit catalog (game_rules.md §10)
INSERT INTO units (id, code, name, cost_gold, attack, defense, hp, upkeep_gold_per_hour) VALUES
    (1, 'pikeman', 'Pikeman', 50, 3, 2, 10, 2.0);

SELECT setval(pg_get_serial_sequence('units', 'id'), (SELECT MAX(id) FROM units));

-- Players
INSERT INTO players (id, name, gold) VALUES
    (1, 'Player1', 200),
    (2, 'Player2', 200);

SELECT setval(pg_get_serial_sequence('players', 'id'), (SELECT MAX(id) FROM players));

-- Castles (gold_per_min = 60)
INSERT INTO castles (player_id, node_id, gold_per_min) VALUES
    (1, 1, 60),
    (2, 6, 60);

-- Heroes at home castles (defaults: speed 10, atk 2, def 2)
INSERT INTO heroes (id, player_id, name, current_node_id, base_speed, attack, defense) VALUES
    (1, 1, 'Hero of Player1', 1, 10, 2, 2),
    (2, 2, 'Hero of Player2', 6, 10, 2, 2);

SELECT setval(pg_get_serial_sequence('heroes', 'id'), (SELECT MAX(id) FROM heroes));

-- Neutral creep at Bandit Camp (node 5)
INSERT INTO neutral_creeps (node_id, name, unit_id, qty, alive, gold_reward) VALUES
    (5, 'Bandit Camp', 1, 50, TRUE, 500);

-- +goose Down
DELETE FROM neutral_creeps;
DELETE FROM heroes;
DELETE FROM castles;
DELETE FROM players;
DELETE FROM map_edges;
DELETE FROM map_nodes;
DELETE FROM units;

ALTER SEQUENCE players_id_seq RESTART WITH 1;
ALTER SEQUENCE map_nodes_id_seq RESTART WITH 1;
ALTER SEQUENCE map_edges_id_seq RESTART WITH 1;
ALTER SEQUENCE units_id_seq RESTART WITH 1;
ALTER SEQUENCE castles_id_seq RESTART WITH 1;
ALTER SEQUENCE heroes_id_seq RESTART WITH 1;
ALTER SEQUENCE neutral_creeps_id_seq RESTART WITH 1;
