-- name: GetUnit :one
SELECT id, code, name, cost_gold, attack, defense, hp, upkeep_gold_per_hour
FROM units
WHERE id = $1;

-- name: GetUnitByCode :one
SELECT id, code, name, cost_gold, attack, defense, hp, upkeep_gold_per_hour
FROM units
WHERE code = $1;

-- name: ListUnits :many
SELECT id, code, name, cost_gold, attack, defense, hp, upkeep_gold_per_hour
FROM units
ORDER BY id;
