-- +goose Up
CREATE TABLE players (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    gold NUMERIC(14, 4) NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE map_nodes (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    x INTEGER NOT NULL,
    y INTEGER NOT NULL,
    kind TEXT NOT NULL CHECK (kind IN ('castle', 'wild', 'creep'))
);

CREATE TABLE map_edges (
    id BIGSERIAL PRIMARY KEY,
    from_node_id BIGINT NOT NULL REFERENCES map_nodes (id),
    to_node_id BIGINT NOT NULL REFERENCES map_nodes (id),
    distance_units INTEGER NOT NULL CHECK (distance_units BETWEEN 1 AND 100)
);

CREATE TABLE units (
    id BIGSERIAL PRIMARY KEY,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    cost_gold INTEGER NOT NULL,
    attack INTEGER NOT NULL,
    defense INTEGER NOT NULL,
    hp INTEGER NOT NULL,
    upkeep_gold_per_hour NUMERIC(8, 4) NOT NULL
);

CREATE TABLE castles (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players (id),
    node_id BIGINT NOT NULL UNIQUE REFERENCES map_nodes (id),
    gold_per_min INTEGER NOT NULL DEFAULT 60,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE heroes (
    id BIGSERIAL PRIMARY KEY,
    player_id BIGINT NOT NULL REFERENCES players (id),
    name TEXT NOT NULL,
    current_node_id BIGINT NOT NULL REFERENCES map_nodes (id),
    base_speed INTEGER NOT NULL DEFAULT 10,
    attack INTEGER NOT NULL DEFAULT 2,
    defense INTEGER NOT NULL DEFAULT 2,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE hero_units (
    hero_id BIGINT NOT NULL REFERENCES heroes (id) ON DELETE CASCADE,
    unit_id BIGINT NOT NULL REFERENCES units (id),
    qty INTEGER NOT NULL CHECK (qty >= 0),
    PRIMARY KEY (hero_id, unit_id)
);

CREATE TABLE movement_orders (
    id BIGSERIAL PRIMARY KEY,
    hero_id BIGINT NOT NULL REFERENCES heroes (id),
    from_node_id BIGINT NOT NULL REFERENCES map_nodes (id),
    to_node_id BIGINT NOT NULL REFERENCES map_nodes (id),
    depart_at TIMESTAMPTZ NOT NULL,
    arrive_at TIMESTAMPTZ NOT NULL,
    status TEXT NOT NULL CHECK (status IN ('in_flight', 'arrived', 'cancelled')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX movement_orders_hero_in_flight_idx ON movement_orders (hero_id)
WHERE status = 'in_flight';

CREATE TABLE neutral_creeps (
    id BIGSERIAL PRIMARY KEY,
    node_id BIGINT NOT NULL UNIQUE REFERENCES map_nodes (id),
    name TEXT NOT NULL,
    unit_id BIGINT NOT NULL REFERENCES units (id),
    qty INTEGER NOT NULL,
    alive BOOLEAN NOT NULL DEFAULT TRUE,
    gold_reward INTEGER NOT NULL
);

CREATE TABLE combat_logs (
    id BIGSERIAL PRIMARY KEY,
    hero_id BIGINT NOT NULL REFERENCES heroes (id),
    creep_id BIGINT REFERENCES neutral_creeps (id),
    outcome TEXT NOT NULL CHECK (outcome IN ('win', 'loss')),
    gold_reward INTEGER NOT NULL DEFAULT 0,
    log JSONB NOT NULL,
    resolved_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS combat_logs;
DROP TABLE IF EXISTS neutral_creeps;
DROP TABLE IF EXISTS movement_orders;
DROP TABLE IF EXISTS hero_units;
DROP TABLE IF EXISTS heroes;
DROP TABLE IF EXISTS castles;
DROP TABLE IF EXISTS units;
DROP TABLE IF EXISTS map_edges;
DROP TABLE IF EXISTS map_nodes;
DROP TABLE IF EXISTS players;
