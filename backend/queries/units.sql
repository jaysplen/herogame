-- name: GetUnit :one
SELECT id, code, name, cost_gold, cost_metal, cost_gems, cost_coal, cost_wood, cost_stone, attack, defense, hp, upkeep_gold_per_hour, faction, tier
FROM units
WHERE id = $1;

-- name: GetUnitByCode :one
SELECT id, code, name, cost_gold, cost_metal, cost_gems, cost_coal, cost_wood, cost_stone, attack, defense, hp, upkeep_gold_per_hour, faction, tier
FROM units
WHERE code = $1;

-- name: ListUnits :many
SELECT id, code, name, cost_gold, cost_metal, cost_gems, cost_coal, cost_wood, cost_stone, attack, defense, hp, upkeep_gold_per_hour, faction, tier
FROM units
ORDER BY id;

-- name: ListUnitsByFactionTier :many
SELECT id, code, name, cost_gold, cost_metal, cost_gems, cost_coal, cost_wood, cost_stone, attack, defense, hp, upkeep_gold_per_hour, faction, tier
FROM units
WHERE (faction = sqlc.arg(faction) OR faction = 'neutral')
  AND tier <= sqlc.arg(max_tier)
ORDER BY tier, id;
