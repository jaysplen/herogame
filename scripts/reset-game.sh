#!/usr/bin/env bash
# Reset runtime game state (heroes, armies, movement, combat, Redis) to fresh seed values.
# Keeps schema and map graph. Requires Postgres + Redis (make dev).
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
export DATABASE_URL="${DATABASE_URL:-postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable}"
export REDIS_URL="${REDIS_URL:-redis://localhost:6379/0}"

run_psql() {
  if command -v psql >/dev/null 2>&1; then
    psql "$DATABASE_URL" -v ON_ERROR_STOP=1 "$@"
  elif docker compose -f "$ROOT/docker-compose.yml" ps postgres 2>/dev/null | grep -q Up; then
    docker compose -f "$ROOT/docker-compose.yml" exec -T postgres \
      psql -U herogame -d herogame -v ON_ERROR_STOP=1 "$@"
  else
    echo "psql not found and postgres container is not running. Run: make dev"
    exit 1
  fi
}

echo "Resetting game state in Postgres..."
run_psql <<'SQL'
DELETE FROM movement_orders;
DELETE FROM hero_units;
DELETE FROM combat_logs;
DELETE FROM player_kills;

UPDATE resource_nodes SET owner_player_id = NULL;

UPDATE players
SET gold = 200, metal = 60, gems = 20, coal = 60, wood = 80, stone = 80
WHERE id IN (1, 2);

UPDATE heroes
SET current_node_id = CASE id WHEN 1 THEN 1 WHEN 2 THEN 6 ELSE current_node_id END,
    spawn_grace_until = now() + interval '12 seconds'
WHERE id IN (1, 2);

UPDATE neutral_creeps SET alive = FALSE;

UPDATE neutral_creeps
SET alive = TRUE,
    qty = 40,
    gold_reward = 420,
    node_id = 5,
    from_node_id = 5,
    to_node_id = 9,
    depart_at = now(),
    arrive_at = now() + interval '30 seconds',
    grace_until = now() + interval '12 seconds',
    attack = 3,
    defense = 2,
    hp = 10
WHERE name = 'Bandit Camp';

UPDATE neutral_creeps
SET alive = TRUE,
    qty = 28,
    gold_reward = 280,
    node_id = 15,
    from_node_id = 15,
    to_node_id = 10,
    depart_at = now(),
    arrive_at = now() + interval '26 seconds',
    grace_until = now() + interval '12 seconds',
    attack = 4,
    defense = 2,
    hp = 9
WHERE name = 'Wolf Den';

UPDATE castles
SET defense_bonus = CASE player_id WHEN 1 THEN 3 WHEN 2 THEN 3 ELSE defense_bonus END,
    barracks_tier = CASE player_id WHEN 1 THEN 1 WHEN 2 THEN 2 ELSE barracks_tier END,
    academy_tier = CASE player_id WHEN 1 THEN 1 WHEN 2 THEN 1 ELSE academy_tier END
WHERE player_id IN (1, 2);
SQL

echo "Flushing Redis (arrivals, respawn lockouts)..."
if command -v redis-cli >/dev/null 2>&1; then
  redis-cli -u "$REDIS_URL" FLUSHDB >/dev/null
elif docker compose -f "$ROOT/docker-compose.yml" ps redis 2>/dev/null | grep -q Up; then
  docker compose -f "$ROOT/docker-compose.yml" exec -T redis redis-cli FLUSHDB >/dev/null
else
  echo "Warning: could not flush Redis; restart Redis or run: redis-cli -u $REDIS_URL FLUSHDB"
fi

echo "Done. Restart make server and refresh the browser (hard refresh)."
