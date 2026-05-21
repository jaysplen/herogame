-- name: ListResourceNodes :many
SELECT id, node_id, resource_type, per_min, owner_player_id, created_at
FROM resource_nodes
ORDER BY id;

-- name: GetResourceNodeByNode :one
SELECT id, node_id, resource_type, per_min, owner_player_id, created_at
FROM resource_nodes
WHERE node_id = $1;

-- name: CaptureResourceNode :exec
UPDATE resource_nodes
SET owner_player_id = sqlc.arg(owner_player_id)
WHERE node_id = sqlc.arg(node_id);

-- name: ListResourceNodesByOwner :many
SELECT id, node_id, resource_type, per_min, owner_player_id, created_at
FROM resource_nodes
WHERE owner_player_id = $1
ORDER BY id;
