package tick_test

import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/herogame/backend/internal/tick"
	"github.com/herogame/backend/internal/ws"
	"github.com/jackc/pgx/v5/pgtype"
	"log/slog"
)

func testEnv(t *testing.T) (*store.Store, *redisx.Client) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable"
	}
	redisURL := os.Getenv("TEST_REDIS_URL")
	if redisURL == "" {
		redisURL = os.Getenv("REDIS_URL")
	}
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}

	_, file, _, _ := runtime.Caller(0)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	_ = os.Setenv("MIGRATIONS_DIR", filepath.Join(backendRoot, "migrations"))

	if err := store.MigrateUp(dsn); err != nil {
		t.Skip("postgres:", err)
	}
	ctx := context.Background()
	st, err := store.New(ctx, dsn)
	if err != nil {
		t.Skip("postgres connect:", err)
	}
	rdb, err := redisx.New(ctx, redisURL)
	if err != nil {
		st.Close()
		t.Skip("redis:", err)
	}
	return st, rdb
}

func TestArrivalResolutionOnce(t *testing.T) {
	st, rdb := testEnv(t)
	defer st.Close()
	defer rdb.Close()

	ctx := context.Background()
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	hub := ws.NewHub(logger)
	eng := tick.NewEngine(st, rdb, hub, logger)

	if err := eng.Arrivals().Rehydrate(ctx); err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	past := now.Add(-2 * time.Second)

	order, err := st.Q.InsertMovementOrder(ctx, gen.InsertMovementOrderParams{
		HeroID:      1,
		FromNodeID:  1,
		ToNodeID:    2,
		DepartAt:    pgtype.Timestamptz{Time: past.Add(-5 * time.Second), Valid: true},
		ArriveAt:    pgtype.Timestamptz{Time: past, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	if err := eng.Arrivals().ScheduleAt(ctx, order.ID, past); err != nil {
		t.Fatal(err)
	}

	eng.TickOnce(ctx)

	updated, err := st.Q.GetMovementOrder(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Status != "arrived" {
		t.Fatalf("status = %q", updated.Status)
	}

	hero, err := st.Q.GetHero(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if hero.CurrentNodeID != 2 {
		t.Fatalf("hero node = %d, want 2", hero.CurrentNodeID)
	}

	// Idempotent second tick must not change state.
	eng.TickOnce(ctx)
	hero2, _ := st.Q.GetHero(ctx, 1)
	if hero2.CurrentNodeID != 2 {
		t.Fatal("double resolution mutated hero")
	}

	ids, err := rdb.DueArrivals(ctx, now, 10)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range ids {
		if id == order.ID {
			t.Fatal("order still in redis zset")
		}
	}

	// cleanup for other tests
	_, _ = st.Pool().Exec(ctx, `UPDATE heroes SET current_node_id = 1 WHERE id = 1`)
	_, _ = st.Pool().Exec(ctx, `DELETE FROM movement_orders WHERE id = $1`, order.ID)
}
