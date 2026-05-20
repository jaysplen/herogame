package ws_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/ws"
)

func TestBuildHeroStateRespawnUntil(t *testing.T) {
	st, rdb := testStoreRedis(t)
	defer st.Close()
	defer rdb.Close()
	ctx := context.Background()

	until := time.Now().UTC().Add(economy.RespawnLockoutSeconds * time.Second)
	if err := rdb.SetRespawnUntil(ctx, 1, until); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_ = rdb.Underlying().Del(ctx, "hero:1:respawn_until")
	})

	state, err := ws.BuildHeroState(ctx, st, rdb, 1)
	if err != nil {
		t.Fatal(err)
	}
	if state.RespawnUntil == nil {
		t.Fatal("expected respawnUntil in hero.state")
	}
	if *state.RespawnUntil < time.Now().UnixMilli() {
		t.Fatalf("respawnUntil %d is in the past", *state.RespawnUntil)
	}
}

func testStoreRedis(t *testing.T) (*store.Store, *redisx.Client) {
	t.Helper()
	st := testStore(t)
	ctx := context.Background()
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	rdb, err := redisx.New(ctx, redisURL)
	if err != nil {
		t.Skip("redis:", err)
	}
	return st, rdb
}
