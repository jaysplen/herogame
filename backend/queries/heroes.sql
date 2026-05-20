-- name: GetHero :one
SELECT id, player_id, name, current_node_id, base_speed, attack, defense, created_at
FROM heroes
WHERE id = $1;

-- name: GetHeroByPlayer :one
SELECT id, player_id, name, current_node_id, base_speed, attack, defense, created_at
FROM heroes
WHERE player_id = $1
LIMIT 1;

-- name: ListHeroes :many
SELECT id, player_id, name, current_node_id, base_speed, attack, defense, created_at
FROM heroes
ORDER BY id;

-- name: UpdateHeroNode :exec
UPDATE heroes
SET current_node_id = $2
WHERE id = $1;

-- name: ListHeroUnitsByHero :many
SELECT
    hu.hero_id,
    hu.unit_id,
    hu.qty,
    u.code,
    u.name,
    u.cost_gold,
    u.attack,
    u.defense,
    u.hp,
    u.upkeep_gold_per_hour
FROM hero_units hu
INNER JOIN units u ON u.id = hu.unit_id
WHERE hu.hero_id = $1
  AND hu.qty > 0;
