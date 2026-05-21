-- name: GetCastleByPlayer :one
SELECT id, player_id, node_id, gold_per_min, faction, defense_bonus, barracks_tier, academy_tier, created_at
FROM castles
WHERE player_id = $1;

-- name: ListAllCastles :many
SELECT id, player_id, node_id, gold_per_min, faction, defense_bonus, barracks_tier, academy_tier, created_at
FROM castles
ORDER BY id;

-- name: GetCastleByNode :one
SELECT id, player_id, node_id, gold_per_min, faction, defense_bonus, barracks_tier, academy_tier, created_at
FROM castles
WHERE node_id = $1;

-- name: UpgradeCastleBuilding :exec
UPDATE castles
SET
    defense_bonus = CASE WHEN sqlc.arg(building_code)::text = 'defense' THEN defense_bonus + 1 ELSE defense_bonus END,
    barracks_tier = CASE WHEN sqlc.arg(building_code)::text = 'barracks' THEN barracks_tier + 1 ELSE barracks_tier END,
    academy_tier = CASE WHEN sqlc.arg(building_code)::text = 'academy' THEN academy_tier + 1 ELSE academy_tier END
WHERE id = sqlc.arg(castle_id);
