-- +goose Up
ALTER TABLE neutral_creeps
    DROP CONSTRAINT IF EXISTS neutral_creeps_node_id_key;

-- +goose Down
ALTER TABLE neutral_creeps
    ADD CONSTRAINT neutral_creeps_node_id_key UNIQUE (node_id);
