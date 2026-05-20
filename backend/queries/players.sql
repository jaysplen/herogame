-- name: GetPlayer :one
SELECT id, name, gold, created_at
FROM players
WHERE id = $1;

-- name: IncrementPlayerGold :exec
UPDATE players
SET gold = gold + sqlc.arg(delta)::numeric
WHERE id = sqlc.arg(id);
