package ws_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/tick"
	"github.com/herogame/backend/internal/world"
	"github.com/herogame/backend/internal/ws"
	"github.com/jackc/pgx/v5"
	"log/slog"
)

func testE2E(t *testing.T) (*store.Store, *redisx.Client, *tick.Engine, *ws.Gateway, string) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
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
		t.Skip("postgres:", err)
	}
	rdb, err := redisx.New(ctx, redisURL)
	if err != nil {
		st.Close()
		t.Skip("redis:", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	hub := ws.NewHub(logger)
	bus := ws.NewEventBus(hub, st, rdb)
	eng := tick.NewEngine(st, rdb, bus, logger)
	_ = eng.Arrivals().Rehydrate(ctx)
	gw := ws.NewGateway(st, rdb, eng.Arrivals(), hub, logger)
	srv := httptest.NewServer(ws.Handler(gw))
	t.Cleanup(srv.Close)
	return st, rdb, eng, gw, "ws" + strings.TrimPrefix(srv.URL, "http")
}

func resetPoCWorld(t *testing.T, st *store.Store, rdb *redisx.Client) {
	t.Helper()
	ctx := context.Background()
	_, err := st.Pool().Exec(ctx, `
		UPDATE neutral_creeps SET alive = TRUE, qty = 50 WHERE node_id = 5;
		UPDATE heroes SET current_node_id = 1 WHERE id = 1;
		UPDATE players SET gold = 500 WHERE id = 1;
		DELETE FROM hero_units WHERE hero_id = 1;
		DELETE FROM movement_orders WHERE hero_id = 1;
		DELETE FROM combat_logs WHERE hero_id = 1;
	`)
	if err != nil {
		t.Fatal(err)
	}
	_ = rdb.Underlying().Del(ctx, "arrivals:zset")
	_ = rdb.Underlying().Del(ctx, "hero:1:respawn_until")
}

func readUntil(t *testing.T, conn *websocket.Conn, wantType string, timeout time.Duration) proto.Envelope {
	t.Helper()
	_ = conn.SetReadDeadline(time.Now().Add(timeout))
	for {
		var env proto.Envelope
		if err := conn.ReadJSON(&env); err != nil {
			t.Fatalf("read %s: %v", wantType, err)
		}
		if env.Type == wantType {
			return env
		}
		if env.Type == proto.TypeError {
			var errP proto.ErrorPayload
			_ = json.Unmarshal(env.Payload, &errP)
			t.Fatalf("unexpected error: %+v", errP)
		}
	}
}

func waitForArrival(t *testing.T, eng *tick.Engine, maxWait time.Duration) {
	t.Helper()
	deadline := time.Now().Add(maxWait)
	for time.Now().Before(deadline) {
		eng.TickOnce(context.Background())
		time.Sleep(200 * time.Millisecond)
	}
	time.Sleep(100 * time.Millisecond)
}

// TestPoCPlaythroughSmoke runs LEAD-001 scripted loop against a live gateway.
func TestPoCPlaythroughSmoke(t *testing.T) {
	st, rdb, eng, _, wsURL := testE2E(t)
	defer st.Close()
	defer rdb.Close()
	resetPoCWorld(t, st, rdb)

	conn := dialHello(t, wsURL, 1)
	defer conn.Close()

	// Buy 10 Pikemen (500 gold; test grants 500g).
	buy, _ := proto.NewEnvelope(proto.TypeUnitBuy, proto.UnitBuyPayload{
		CastleID: 1, UnitTypeID: 1, Qty: 10,
	}, 2)
	if err := conn.WriteJSON(buy); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		env := readUntil(t, conn, proto.TypeHeroState, 3*time.Second)
		var hs proto.HeroStatePayload
		_ = json.Unmarshal(env.Payload, &hs)
		if hs.ArmySize == 10 {
			break
		}
		if i == 2 {
			t.Fatalf("army size after buy = %d, want 10", hs.ArmySize)
		}
	}

	// Move castle → Crossroads (node 2).
	move1, _ := proto.NewEnvelope(proto.TypeMoveRequest, proto.MoveRequestPayload{
		HeroID: 1, TargetNodeID: 2,
	}, 3)
	if err := conn.WriteJSON(move1); err != nil {
		t.Fatal(err)
	}
	readUntil(t, conn, proto.TypeMoveUpdate, 3*time.Second)
	waitForArrival(t, eng, 8*time.Second)
	readUntil(t, conn, proto.TypeMoveArrived, 3*time.Second)

	hero, err := st.Q.GetHero(context.Background(), 1)
	if err != nil || hero.CurrentNodeID != 2 {
		t.Fatalf("hero at node %d after leg 1, want 2", hero.CurrentNodeID)
	}

	// Crossroads → Bandit Camp (node 5) → combat on arrival.
	move2, _ := proto.NewEnvelope(proto.TypeMoveRequest, proto.MoveRequestPayload{
		HeroID: 1, TargetNodeID: 5,
	}, 4)
	if err := conn.WriteJSON(move2); err != nil {
		t.Fatal(err)
	}
	readUntil(t, conn, proto.TypeMoveUpdate, 3*time.Second)
	waitForArrival(t, eng, 10*time.Second)
	arrivedEnv := readUntil(t, conn, proto.TypeMoveArrived, 3*time.Second)
	combatEnv := readUntil(t, conn, proto.TypeCombatResolved, 5*time.Second)
	heroEnv := readUntil(t, conn, proto.TypeHeroState, 3*time.Second)

	var arrived proto.MoveArrivedPayload
	if err := json.Unmarshal(arrivedEnv.Payload, &arrived); err != nil {
		t.Fatal(err)
	}
	var combat proto.CombatResolvedPayload
	if err := json.Unmarshal(combatEnv.Payload, &combat); err != nil {
		t.Fatal(err)
	}
	var heroState proto.HeroStatePayload
	if err := json.Unmarshal(heroEnv.Payload, &heroState); err != nil {
		t.Fatal(err)
	}
	if combat.Outcome != "win" && combat.Outcome != "loss" {
		t.Fatalf("combat outcome = %q", combat.Outcome)
	}
	t.Logf("combat outcome=%s casualties=%d goldReward=%d logEntries=%d",
		combat.Outcome, combat.Casualties, combat.GoldReward, len(combat.Log))

	// DB reconciliation (LEAD-001).
	ctx := context.Background()
	player, err := st.Q.GetPlayer(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	gold, _ := player.Gold.Float64Value()

	var logCount int
	var dbOutcome string
	var dbGoldReward int32
	err = st.Pool().QueryRow(ctx, `
		SELECT outcome, gold_reward
		FROM combat_logs
		WHERE hero_id = 1
		ORDER BY id DESC
		LIMIT 1
	`).Scan(&dbOutcome, &dbGoldReward)
	if err == nil {
		logCount = 1
	}
	if logCount == 0 {
		t.Fatal("expected combat_logs row")
	}
	if dbOutcome != combat.Outcome {
		t.Fatalf("combat_logs.outcome = %q, envelope = %q", dbOutcome, combat.Outcome)
	}
	if dbGoldReward != combat.GoldReward {
		t.Fatalf("combat_logs.gold_reward = %d, envelope = %d", dbGoldReward, combat.GoldReward)
	}
	if combat.Outcome == "win" {
		if gold.Float64 < 0 {
			t.Fatalf("player gold after win = %v", gold.Float64)
		}
		if _, err := st.Q.GetCreepByNode(ctx, 5); err != pgx.ErrNoRows {
			t.Fatal("creep should be dead after win")
		}
	} else {
		if arrived.NodeID != 1 {
			t.Fatalf("move.arrived node = %d, want castle node 1 after defeat teleport", arrived.NodeID)
		}
		if hero, _ := st.Q.GetHero(ctx, 1); hero.CurrentNodeID != 1 {
			t.Fatalf("defeated hero should be at castle node 1, got %d", hero.CurrentNodeID)
		}
		respawning, err := rdb.IsRespawning(ctx, 1, time.Now().UTC())
		if err != nil || !respawning {
			t.Fatal("expected respawn lockout in redis after loss")
		}
		if heroState.RespawnUntil == nil || *heroState.RespawnUntil <= time.Now().UnixMilli() {
			t.Fatalf("hero.state missing respawnUntil after loss: %+v", heroState)
		}
	}

	resetPoCWorld(t, st, rdb)
}

// TestArmySlowdownObservable verifies larger armies take longer on the same edge (LEAD-001).
func TestArmySlowdownObservable(t *testing.T) {
	const dist = 30 // Crossroads → Bandit Camp
	const base = economy.HeroBaseSpeedDefault
	small := world.TravelSeconds(dist, base, 10)
	large := world.TravelSeconds(dist, base, 200)
	if large <= small {
		t.Fatalf("TravelSeconds(30,10,200)=%d must exceed TravelSeconds(30,10,10)=%d", large, small)
	}
	ratio := float64(large) / float64(small)
	if ratio < 4 {
		t.Fatalf("slowdown ratio %.2f too small; want clearly observable (>4×)", ratio)
	}
	t.Logf("travel seconds: army10=%d army200=%d ratio=%.2f", small, large, ratio)
}
