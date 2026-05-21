-- name: InsertCombatLog :one
INSERT INTO combat_logs (hero_id, creep_id, enemy_hero_id, outcome, gold_reward, converted_units, log)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING id, hero_id, creep_id, enemy_hero_id, outcome, gold_reward, converted_units, log, resolved_at;

-- name: GetCreepByNode :one
SELECT id, node_id, name, unit_id, qty, alive, gold_reward, from_node_id, to_node_id, depart_at, arrive_at, speed_units, attack, defense, hp, grace_until
FROM neutral_creeps
WHERE node_id = $1
  AND alive = TRUE;

-- name: SetCreepDead :exec
UPDATE neutral_creeps
SET alive = FALSE
WHERE id = $1;

-- name: ListAliveCreeps :many
SELECT id, node_id, name, unit_id, qty, alive, gold_reward, from_node_id, to_node_id, depart_at, arrive_at, speed_units, attack, defense, hp, grace_until
FROM neutral_creeps
WHERE alive = TRUE
ORDER BY id;

-- name: UpdateCreepMovement :exec
UPDATE neutral_creeps
SET
    node_id = sqlc.arg(node_id),
    from_node_id = sqlc.arg(from_node_id),
    to_node_id = sqlc.arg(to_node_id),
    depart_at = sqlc.arg(depart_at),
    arrive_at = sqlc.arg(arrive_at)
WHERE id = sqlc.arg(id);
