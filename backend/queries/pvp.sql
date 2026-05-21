-- name: IncrementPlayerKill :exec
INSERT INTO player_kills (killer_player_id, victim_player_id, kills)
VALUES ($1, $2, 1)
ON CONFLICT (killer_player_id, victim_player_id)
DO UPDATE SET kills = player_kills.kills + 1;

-- name: GetPlayerKillCount :one
SELECT kills
FROM player_kills
WHERE killer_player_id = $1
  AND victim_player_id = $2;
