-- name: GetPlayer :one
SELECT id, name, gold, metal, gems, coal, wood, stone, created_at
FROM players
WHERE id = $1;

-- name: IncrementPlayerGold :exec
UPDATE players
SET gold = gold + sqlc.arg(delta)::numeric
WHERE id = sqlc.arg(id);

-- name: IncrementPlayerResources :exec
UPDATE players
SET
    gold = gold + sqlc.arg(gold_delta)::numeric,
    metal = metal + sqlc.arg(metal_delta)::numeric,
    gems = gems + sqlc.arg(gems_delta)::numeric,
    coal = coal + sqlc.arg(coal_delta)::numeric,
    wood = wood + sqlc.arg(wood_delta)::numeric,
    stone = stone + sqlc.arg(stone_delta)::numeric
WHERE id = sqlc.arg(id);
