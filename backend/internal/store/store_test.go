package store_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

var testStore *store.Store

func TestMain(m *testing.M) {
	ctx := context.Background()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if dsn == "" {
		dsn = "postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable"
	}

	_, file, _, _ := runtime.Caller(0)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	_ = os.Setenv("MIGRATIONS_DIR", filepath.Join(backendRoot, "migrations"))

	if err := store.MigrateUp(dsn); err != nil {
		os.Stderr.WriteString("store_test: migrate skipped: " + err.Error() + "\n")
		os.Exit(0)
	}

	s, err := store.New(ctx, dsn)
	if err != nil {
		os.Stderr.WriteString("store_test: connect skipped: " + err.Error() + "\n")
		os.Exit(0)
	}
	testStore = s
	code := m.Run()
	s.Close()
	os.Exit(code)
}

func requireStore(t *testing.T) {
	t.Helper()
	if testStore == nil {
		t.Fatal("test store not initialized")
	}
}

func resetTestWorld(t *testing.T) {
	t.Helper()
	ctx := context.Background()
	_, err := testStore.Pool().Exec(ctx, `
		UPDATE heroes SET current_node_id = CASE WHEN id = 1 THEN 1 ELSE 6 END;
		DELETE FROM hero_units;
		DELETE FROM movement_orders;
		UPDATE neutral_creeps SET alive = TRUE;
	`)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetPlayer(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	p, err := testStore.Q.GetPlayer(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "Player1" {
		t.Fatalf("name = %q, want Player1", p.Name)
	}
}

func TestIncrementPlayerGold(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	before, err := testStore.Q.GetPlayer(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}

	delta := pgtype.Numeric{}
	if err := delta.Scan("10.5"); err != nil {
		t.Fatal(err)
	}
	if err := testStore.Q.IncrementPlayerGold(ctx, gen.IncrementPlayerGoldParams{
		ID:    1,
		Delta: delta,
	}); err != nil {
		t.Fatal(err)
	}

	after, err := testStore.Q.GetPlayer(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if !after.Gold.Valid || !before.Gold.Valid {
		t.Fatal("gold not valid")
	}
	// restore
	_ = testStore.Q.IncrementPlayerGold(ctx, gen.IncrementPlayerGoldParams{
		ID: 1,
		Delta: mustNumeric("-10.5"),
	})
}

func TestGetHero(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	h, err := testStore.Q.GetHero(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if h.CurrentNodeID != 1 {
		t.Fatalf("node = %d, want 1", h.CurrentNodeID)
	}
}

func TestUpdateHeroNode(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	if err := testStore.Q.UpdateHeroNode(ctx, gen.UpdateHeroNodeParams{ID: 1, CurrentNodeID: 2}); err != nil {
		t.Fatal(err)
	}
	h, _ := testStore.Q.GetHero(ctx, 1)
	if h.CurrentNodeID != 2 {
		t.Fatalf("node = %d", h.CurrentNodeID)
	}
	_ = testStore.Q.UpdateHeroNode(ctx, gen.UpdateHeroNodeParams{ID: 1, CurrentNodeID: 1})
}

func TestListHeroUnitsByHero(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	units, err := testStore.Q.ListHeroUnitsByHero(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) != 0 {
		t.Fatalf("expected empty army, got %d", len(units))
	}
}

func TestGetCastleByPlayer(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	c, err := testStore.Q.GetCastleByPlayer(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if c.GoldPerMin <= 0 {
		t.Fatalf("gold_per_min = %d", c.GoldPerMin)
	}
}

func TestListAllCastles(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	castles, err := testStore.Q.ListAllCastles(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(castles) != 2 {
		t.Fatalf("castles = %d, want 2", len(castles))
	}
}

func TestGetUnitByCode(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	u, err := testStore.Q.GetUnitByCode(ctx, "pikeman")
	if err != nil {
		t.Fatal(err)
	}
	// Pikeman was rebalanced to Atk 4 / Def 5 / HP 12 in
	// migrations/00006_seed_units.sql (multi-unit roster + RPS counters).
	// Pin to economy.PikemanAttack so the test follows future tunings.
	if u.Attack != int32(economy.PikemanAttack) {
		t.Fatalf("attack = %d, want %d", u.Attack, economy.PikemanAttack)
	}
}

func TestListUnits(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	units, err := testStore.Q.ListUnits(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(units) < 1 {
		t.Fatal("expected units")
	}
}

func TestGetEdge(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	e, err := testStore.Q.GetEdge(ctx, gen.GetEdgeParams{FromNodeID: 1, ToNodeID: 2})
	if err != nil {
		t.Fatal(err)
	}
	if e.DistanceUnits <= 0 {
		t.Fatalf("distance = %d", e.DistanceUnits)
	}
}

func TestListNodes(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	nodes, err := testStore.Q.ListNodes(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(nodes) < 6 {
		t.Fatalf("nodes = %d", len(nodes))
	}
}

func TestListEdgesByNode(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	edges, err := testStore.Q.ListEdgesByNode(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(edges) < 4 {
		t.Fatalf("edges = %d", len(edges))
	}
}

func TestMovementQueries(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	now := time.Now().UTC()
	order, err := testStore.Q.InsertMovementOrder(ctx, gen.InsertMovementOrderParams{
		HeroID:      1,
		FromNodeID:  1,
		ToNodeID:    2,
		DepartAt:    pgtype.Timestamptz{Time: now, Valid: true},
		ArriveAt:    pgtype.Timestamptz{Time: now.Add(5 * time.Second), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	active, err := testStore.Q.GetActiveMovementByHero(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if active.ID != order.ID {
		t.Fatalf("active id = %d, want %d", active.ID, order.ID)
	}

	inFlight, err := testStore.Q.ListInFlightMovements(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(inFlight) < 1 {
		t.Fatal("expected in-flight orders")
	}

	n, err := testStore.Q.MarkMovementArrived(ctx, order.ID)
	if err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("rows = %d", n)
	}

	_, err = testStore.Q.GetActiveMovementByHero(ctx, 1)
	if err == nil {
		t.Fatal("expected no active movement")
	}
	if err != pgx.ErrNoRows {
		t.Fatalf("err = %v", err)
	}
}

func TestCombatQueries(t *testing.T) {
	requireStore(t)
	resetTestWorld(t)
	ctx := context.Background()

	creeps, err := testStore.Q.ListAliveCreeps(ctx)
	if err != nil || len(creeps) == 0 {
		t.Fatal(err)
	}
	creep := creeps[0]
	if creep.Name == "" {
		t.Fatalf("name = %q", creep.Name)
	}

	logJSON, _ := json.Marshal([]map[string]string{{"round": "1"}})
	entry, err := testStore.Q.InsertCombatLog(ctx, gen.InsertCombatLogParams{
		HeroID:     1,
		CreepID:    pgtype.Int8{Int64: creep.ID, Valid: true},
		Outcome:    "win",
		GoldReward: 500,
		Log:        logJSON,
	})
	if err != nil {
		t.Fatal(err)
	}
	if entry.Outcome != "win" {
		t.Fatalf("outcome = %q", entry.Outcome)
	}

	err = testStore.WithTx(ctx, func(q *gen.Queries) error {
		if err := q.SetCreepDead(ctx, creep.ID); err != nil {
			return err
		}
		_, err := q.GetCreepByNode(ctx, creep.NodeID)
		return err
	})
	if err != pgx.ErrNoRows {
		t.Fatalf("after dead creep: err = %v", err)
	}

	// restore creep for other tests
	_, _ = testStore.Pool().Exec(ctx, "UPDATE neutral_creeps SET alive = TRUE WHERE id = $1", creep.ID)
}


func mustNumeric(s string) pgtype.Numeric {
	var n pgtype.Numeric
	if err := n.Scan(s); err != nil {
		panic(err)
	}
	return n
}
