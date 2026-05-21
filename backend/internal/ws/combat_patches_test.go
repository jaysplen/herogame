package ws_test

import (
	"context"
	"testing"
	"time"

	"github.com/herogame/backend/internal/combat"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

/*
Regression tests for the post-arrival combat patches.

Each test below opens its own transaction, runs setup + the combat
function + reads its result, and defers tx.Rollback so the mutations
never get committed to the shared test database. This matters because
this CI runs every package's tests in parallel against the same
Postgres instance — DELETE FROM movement_orders / hero_units in a
helper that doesn't honor that contract was causing flaky failures
in store / tick / bootstrap tests.

The one exception is TestResolveOrderAtPlacesHeroAtMeetingNode, which
exercises arrivals.Scheduler.ResolveOrderAt — that function opens its
own transaction internally and commits. We minimise its blast radius
with targeted DELETE/UPDATE cleanups in t.Cleanup so subsequent tests
see a clean slate for hero 1.
*/

// TestApplyAtNodeReportsLoserHeroIDOnLoss verifies that ApplyAtNode marks
// the hero as the loser (LoserHeroID = heroID, Respawn = true) when a
// creep defeats them. arrivals.resolveOne uses this id to scope the
// respawn lockout to the hero that actually lost.
func TestApplyAtNodeReportsLoserHeroIDOnLoss(t *testing.T) {
	st, rdb, _, _, _ := testE2E(t)
	defer st.Close()
	defer rdb.Close()
	ctx := context.Background()

	tx, err := st.Pool().Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		UPDATE heroes SET current_node_id = 5, spawn_grace_until = to_timestamp(0) WHERE id = 1;
		DELETE FROM hero_units WHERE hero_id = 1;
		UPDATE neutral_creeps
		SET alive = TRUE,
		    node_id = 5,
		    qty = 50, attack = 4, defense = 5, hp = 12,
		    grace_until = now() - interval '1 second'
		WHERE name = 'Bandit Camp';
	`); err != nil {
		t.Fatal(err)
	}

	result, err := combat.ApplyAtNode(ctx, st.Q.WithTx(tx), 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected an ApplyResult on creep-vs-empty-hero loss, got nil")
	}
	if result.Payload.Outcome != "loss" {
		t.Fatalf("outcome = %q, want loss", result.Payload.Outcome)
	}
	if !result.Respawn {
		t.Fatal("Respawn = false on hero loss; want true")
	}
	if result.LoserHeroID != 1 {
		t.Fatalf("LoserHeroID = %d, want 1 (the defeated hero)", result.LoserHeroID)
	}
}

// TestApplyAtNodeNoLoserHeroIDOnWin — creep loss path returns
// LoserHeroID = 0 so arrivals.resolveOne skips the lockout. The creep
// has no hero id to apply a respawn lockout to.
func TestApplyAtNodeNoLoserHeroIDOnWin(t *testing.T) {
	st, rdb, _, _, _ := testE2E(t)
	defer st.Close()
	defer rdb.Close()
	ctx := context.Background()

	tx, err := st.Pool().Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if _, err := tx.Exec(ctx, `
		UPDATE heroes SET current_node_id = 5, spawn_grace_until = to_timestamp(0) WHERE id = 1;
		DELETE FROM hero_units WHERE hero_id = 1;
		INSERT INTO hero_units (hero_id, unit_id, qty) VALUES (1, 1, 500);
		UPDATE neutral_creeps
		SET alive = TRUE,
		    node_id = 5,
		    qty = 5, attack = 4, defense = 5, hp = 12,
		    grace_until = now() - interval '1 second'
		WHERE name = 'Bandit Camp';
	`); err != nil {
		t.Fatal(err)
	}

	result, err := combat.ApplyAtNode(ctx, st.Q.WithTx(tx), 1, 5)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected an ApplyResult on hero win, got nil")
	}
	if result.Payload.Outcome != "win" {
		t.Fatalf("outcome = %q, want win", result.Payload.Outcome)
	}
	if result.Respawn {
		t.Fatal("Respawn = true on hero win; want false")
	}
	if result.LoserHeroID != 0 {
		t.Fatalf("LoserHeroID = %d on hero win, want 0 (no hero loser)", result.LoserHeroID)
	}
}

// TestApplyHeroVsHeroLoserHeroIDIsActualLoser pins down the PvP
// contract. Whichever side loses the fight is reported as
// LoserHeroID, regardless of who initiated the attack.
func TestApplyHeroVsHeroLoserHeroIDIsActualLoser(t *testing.T) {
	st, rdb, _, _, _ := testE2E(t)
	defer st.Close()
	defer rdb.Close()
	ctx := context.Background()

	// First scenario: attacker (hero 1) has the big army and wins.
	t.Run("attacker_wins", func(t *testing.T) {
		tx, err := st.Pool().Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		if _, err := tx.Exec(ctx, `
			UPDATE heroes SET current_node_id = 5, spawn_grace_until = to_timestamp(0) WHERE id IN (1, 2);
			DELETE FROM hero_units WHERE hero_id IN (1, 2);
			INSERT INTO hero_units (hero_id, unit_id, qty) VALUES (1, 1, 400), (2, 1, 5);
		`); err != nil {
			t.Fatal(err)
		}

		result, err := combat.ApplyHeroVsHero(ctx, st.Q.WithTx(tx), 1, 2)
		if err != nil {
			t.Fatal(err)
		}
		if result == nil {
			t.Fatal("expected PvP ApplyResult, got nil")
		}
		if result.Payload.Outcome != "win" {
			t.Fatalf("attacker outcome = %q, want win (attacker had the big army)", result.Payload.Outcome)
		}
		if !result.Respawn {
			t.Fatal("Respawn = false after PvP; want true so the loser is locked out")
		}
		if result.LoserHeroID != 2 {
			t.Fatalf("LoserHeroID = %d, want 2 (defender lost)", result.LoserHeroID)
		}
		if result.Payload.HeroID != 1 {
			t.Fatalf("Payload.HeroID = %d, want 1 (winning attacker)", result.Payload.HeroID)
		}
	})

	// Second scenario: defender has the big army; attacker loses.
	t.Run("defender_wins", func(t *testing.T) {
		tx, err := st.Pool().Begin(ctx)
		if err != nil {
			t.Fatal(err)
		}
		defer func() { _ = tx.Rollback(ctx) }()

		if _, err := tx.Exec(ctx, `
			UPDATE heroes SET current_node_id = 5, spawn_grace_until = to_timestamp(0) WHERE id IN (1, 2);
			DELETE FROM hero_units WHERE hero_id IN (1, 2);
			INSERT INTO hero_units (hero_id, unit_id, qty) VALUES (1, 1, 5), (2, 1, 400);
		`); err != nil {
			t.Fatal(err)
		}

		result, err := combat.ApplyHeroVsHero(ctx, st.Q.WithTx(tx), 1, 2)
		if err != nil {
			t.Fatal(err)
		}
		if result == nil {
			t.Fatal("expected PvP ApplyResult, got nil")
		}
		if result.Payload.Outcome != "loss" {
			t.Fatalf("attacker outcome = %q, want loss (defender had the big army)", result.Payload.Outcome)
		}
		if result.LoserHeroID != 1 {
			t.Fatalf("LoserHeroID = %d, want 1 (attacker lost)", result.LoserHeroID)
		}
	})
}

// TestResolveOrderAtPlacesHeroAtMeetingNode is the targeted Patch 1
// guard: arrivals.Scheduler.ResolveOrderAt(orderID, meetingNodeID)
// must place the hero at meetingNodeID, not at the order's original
// destination. ResolveOrderAt opens its own transaction internally so
// we can't use the rollback-for-isolation trick here — we minimise
// blast radius with targeted cleanup of the rows we touch.
func TestResolveOrderAtPlacesHeroAtMeetingNode(t *testing.T) {
	st, rdb, eng, _, _ := testE2E(t)
	defer st.Close()
	defer rdb.Close()
	ctx := context.Background()

	// Park hero 1 at node 1 with no army, no creeps anywhere, no
	// in-flight movements that could be confused with ours. Use
	// targeted updates that other tests' resets also issue — there's
	// no avoiding the shared rows, but scope the writes tightly.
	originalCreepStates := map[string]bool{}
	rows, err := st.Pool().Query(ctx, `SELECT name, alive FROM neutral_creeps`)
	if err != nil {
		t.Fatal(err)
	}
	for rows.Next() {
		var name string
		var alive bool
		if err := rows.Scan(&name, &alive); err != nil {
			rows.Close()
			t.Fatal(err)
		}
		originalCreepStates[name] = alive
	}
	rows.Close()

	if _, err := st.Pool().Exec(ctx, `
		UPDATE heroes SET current_node_id = 1, spawn_grace_until = to_timestamp(0) WHERE id = 1;
		DELETE FROM hero_units WHERE hero_id = 1;
		UPDATE neutral_creeps SET alive = FALSE;
	`); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		// Restore creep alive state and recenter hero 1 so subsequent
		// tests in any package see a sensible baseline.
		for name, alive := range originalCreepStates {
			_, _ = st.Pool().Exec(ctx,
				`UPDATE neutral_creeps SET alive = $1 WHERE name = $2`, alive, name)
		}
		_, _ = st.Pool().Exec(ctx,
			`UPDATE heroes SET current_node_id = 1 WHERE id = 1`)
	})

	now := time.Now().UTC()
	order1, err := st.Q.InsertMovementOrder(ctx, gen.InsertMovementOrderParams{
		HeroID:     1,
		FromNodeID: 1,
		ToNodeID:   5,
		DepartAt:   pgtype.Timestamptz{Time: now, Valid: true},
		ArriveAt:   pgtype.Timestamptz{Time: now.Add(60 * time.Second), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	order2 := gen.MovementOrder{}
	t.Cleanup(func() {
		_, _ = st.Pool().Exec(ctx,
			`DELETE FROM movement_orders WHERE id IN ($1, $2)`,
			order1.ID, order2.ID)
	})

	// Force-resolve at node 2 (Moss Crossing) — pretending a collision
	// happened mid-path. Hero must end up at node 2, not at node 5.
	if err := eng.Arrivals().ResolveOrderAt(ctx, order1.ID, 2); err != nil {
		t.Fatalf("ResolveOrderAt: %v", err)
	}
	h, err := st.Q.GetHero(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if h.CurrentNodeID != 2 {
		t.Fatalf("hero current_node_id = %d after ResolveOrderAt(_, 2); want 2 (the meeting node, NOT the original destination 5)", h.CurrentNodeID)
	}

	// Sanity: ResolveOrder (no meeting node) still honors the order's
	// original destination.
	now2 := time.Now().UTC()
	order2, err = st.Q.InsertMovementOrder(ctx, gen.InsertMovementOrderParams{
		HeroID:     1,
		FromNodeID: 2,
		ToNodeID:   3,
		DepartAt:   pgtype.Timestamptz{Time: now2, Valid: true},
		ArriveAt:   pgtype.Timestamptz{Time: now2.Add(60 * time.Second), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := eng.Arrivals().ResolveOrder(ctx, order2.ID); err != nil {
		t.Fatalf("ResolveOrder: %v", err)
	}
	h, err = st.Q.GetHero(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	if h.CurrentNodeID != 3 {
		t.Fatalf("hero current_node_id = %d after ResolveOrder; want 3 (order's ToNodeID)", h.CurrentNodeID)
	}
}
