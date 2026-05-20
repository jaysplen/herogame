-- name: GetCastleByPlayer :one
SELECT id, player_id, node_id, gold_per_min, created_at
FROM castles
WHERE player_id = $1;

-- name: ListAllCastles :many
SELECT id, player_id, node_id, gold_per_min, created_at
FROM castles
ORDER BY id;
