-- name: InsertCombatLog :one
INSERT INTO combat_logs (hero_id, creep_id, outcome, gold_reward, log)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, hero_id, creep_id, outcome, gold_reward, log, resolved_at;

-- name: GetCreepByNode :one
SELECT id, node_id, name, unit_id, qty, alive, gold_reward
FROM neutral_creeps
WHERE node_id = $1
  AND alive = TRUE;

-- name: SetCreepDead :exec
UPDATE neutral_creeps
SET alive = FALSE
WHERE id = $1;
