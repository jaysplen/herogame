// Command e2e-reset pins the game world into a deterministic state for
// the Playwright smoke test. Runs the same DB/Redis cleanup that
// scripts/reset-game.sh ships, but without requiring psql/redis-cli in
// the CI image — uses the same pgx/redis libraries the server already
// depends on.
//
// Crucially, the Bandit Camp neutral creep is pinned at node 5 with a
// 10-minute `arrive_at` so the creep sweeper does NOT advance it off
// node 5 during the e2e run. Without this pin, by the time Playwright
// has waited 13 s for spawn grace, bought four pikemen, marched
// 1 -> 2, and then 2 -> 5 (typically ~80 s after migration), the
// creep's seeded 30 s leg has long expired and GetCreepByNode(5)
// returns ErrNoRows — so ApplyAtNode sees no creep at the hero's
// destination and combat never resolves.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	redis "github.com/redis/go-redis/v9"
)

const resetSQL = `
DELETE FROM movement_orders;
DELETE FROM hero_units;
DELETE FROM combat_logs;
DELETE FROM player_kills;

UPDATE resource_nodes SET owner_player_id = NULL;

UPDATE players
SET gold = 1000, metal = 200, gems = 100, coal = 200, wood = 200, stone = 200
WHERE id IN (1, 2);

UPDATE heroes
SET current_node_id = CASE id WHEN 1 THEN 1 WHEN 2 THEN 6 ELSE current_node_id END,
    spawn_grace_until = to_timestamp(0)
WHERE id IN (1, 2);

UPDATE neutral_creeps SET alive = FALSE;

-- Pin Bandit Camp at node 5 with a 10-minute non-moving leg so it
-- stays where the e2e test expects to find it.
UPDATE neutral_creeps
SET alive       = TRUE,
    qty         = 40,
    gold_reward = 420,
    node_id     = 5,
    from_node_id = 5,
    to_node_id   = 5,
    depart_at   = now(),
    arrive_at   = now() + interval '10 minutes',
    grace_until = now() - interval '1 second',
    attack      = 3,
    defense     = 2,
    hp          = 10
WHERE name = 'Bandit Camp';

-- Wolf Den stays dead during the e2e smoke test to avoid second
-- creeps wandering into the hero's path.
UPDATE neutral_creeps SET alive = FALSE WHERE name = 'Wolf Den';
`

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		logger.Error("DATABASE_URL is required")
		os.Exit(1)
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		logger.Error("pgxpool new", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer pool.Close()

	if _, err := pool.Exec(ctx, resetSQL); err != nil {
		logger.Error("reset SQL", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("postgres state reset (heroes, creeps, movement, resources)")

	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		logger.Error("redis url parse", slog.String("error", err.Error()))
		os.Exit(1)
	}
	rdb := redis.NewClient(opt)
	defer rdb.Close()
	if err := rdb.FlushDB(ctx).Err(); err != nil {
		logger.Error("redis FLUSHDB", slog.String("error", err.Error()))
		os.Exit(1)
	}
	logger.Info("redis FLUSHDB done (arrivals, respawn lockouts cleared)")

	fmt.Println("e2e-reset OK")
}
