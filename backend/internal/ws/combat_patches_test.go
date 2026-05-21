package ws_test

import (
	"context"
	"testing"
	"time"

	"github.com/herogame/backend/internal/combat"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

// resetSimpleWorld wipes mutable state for the patch tests. Kept separate
// from resetPoCWorld (in smoke_test.go) so changes here don't perturb the
// existing smoke test scenario.
func resetSimpleWorld(t *testing.T, st *store.Store, rdb *redisx.Client) {
	t.Helper()
	ctx := context.Background()
	_, err := st.Pool().Exec(ctx, `
		UPDATE heroes SET current_node_id = 1, spawn_grace_until = to_timestamp(0) WHERE id = 1;
		UPDATE heroes SET current_node_id = 6, spawn_grace_until = to_timestamp(0) WHERE id = 2;
		DELETE FROM hero_units;
		DELETE FROM movement_orders;
		DELETE FROM combat_logs;
		UPDATE neutral_creeps SET alive = FALSE;
	`)
	if err != nil {
		t.Fatal(err)
	}
	_ = rdb.Underlying().Del(ctx,
		"arrivals:zset",
		"hero:1:respawn_until",
		"hero:2:respawn_until",
	).Err()
}

// TestApplyAtNodeReportsLoserHeroIDOnLoss verifies that ApplyAtNode marks
// the hero as the loser (LoserHeroID = heroID, Respawn = true) when a
// creep defeats them. arrivals.resolveOne uses this id to scope the
// respawn lockout to the hero that actually lost — pre-patch the lockout
// went to whichever hero's order was being resolved, which is fine for
// creep-vs-hero but wrong for PvP. The field also has to be 0 when a
// hero wins (creep has no hero id to lock out).
//
// This file is the regression guard for both rules.
func TestApplyAtNodeReportsLoserHeroIDOnLoss(t *testing.T) {
	st, rdb, _, _, _ := testE2E(t)
	defer st.Close()
	defer rdb.Close()

	ctx := context.Background()
	resetSimpleWorld(t, st, rdb)

	// Place hero 1 at the Bandit Camp node with an empty army so combat
	// resolves into "loss" against a non-trivial creep stack.
	if _, err := st.Pool().Exec(ctx, `
		UPDATE heroes SET current_node_id = 5 WHERE id = 1;
		UPDATE neutral_creeps
		SET alive = TRUE,
		    node_id = 5,
		    qty = 50, attack = 4, defense = 5, hp = 12,
		    grace_until = now() - interval '1 second'
		WHERE name = 'Bandit Camp';
	`); err != nil {
		t.Fatal(err)
	}

	var result *combat.ApplyResult
	err := st.WithTx(ctx, func(q *gen.Queries) error {
		r, err := combat.ApplyAtNode(ctx, q, 1, 5)
		result = r
		return err
	})
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
// LoserHeroID = 0 so arrivals.resolveOne skips the lockout. The creep has
// no hero id to apply a respawn lockout to.
func TestApplyAtNodeNoLoserHeroIDOnWin(t *testing.T) {
	st, rdb, _, _, _ := testE2E(t)
	defer st.Close()
	defer rdb.Close()

	ctx := context.Background()
	resetSimpleWorld(t, st, rdb)

	if _, err := st.Pool().Exec(ctx, `
		UPDATE heroes SET current_node_id = 5 WHERE id = 1;
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

	var result *combat.ApplyResult
	err := st.WithTx(ctx, func(q *gen.Queries) error {
		r, err := combat.ApplyAtNode(ctx, q, 1, 5)
		result = r
		return err
	})
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

// TestApplyHeroVsHeroLoserHeroIDIsActualLoser pins down the PvP contract.
// Whichever side loses the fight is reported as LoserHeroID, regardless
// of who initiated the attack. Pre-patch the attacker (resolved.HeroID)
// got the respawn lockout no matter what — this guards against
// reintroducing that.
func TestApplyHeroVsHeroLoserHeroIDIsActualLoser(t *testing.T) {
	st, rdb, _, _, _ := testE2E(t)
	defer st.Close()
	defer rdb.Close()

	ctx := context.Background()
	resetSimpleWorld(t, st, rdb)

	// Co-locate both heroes at node 5 with different army sizes so the
	// outcome is deterministic. Hero 1 (attacker) has the big army.
	if _, err := st.Pool().Exec(ctx, `
		UPDATE heroes SET current_node_id = 5 WHERE id IN (1, 2);
		INSERT INTO hero_units (hero_id, unit_id, qty) VALUES (1, 1, 400), (2, 1, 5);
	`); err != nil {
		t.Fatal(err)
	}

	var result *combat.ApplyResult
	err := st.WithTx(ctx, func(q *gen.Queries) error {
		r, err := combat.ApplyHeroVsHero(ctx, q, 1, 2)
		result = r
		return err
	})
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

	// Flip armies and rerun: now the attacker should lose.
	resetSimpleWorld(t, st, rdb)
	if _, err := st.Pool().Exec(ctx, `
		UPDATE heroes SET current_node_id = 5 WHERE id IN (1, 2);
		INSERT INTO hero_units (hero_id, unit_id, qty) VALUES (1, 1, 5), (2, 1, 400);
	`); err != nil {
		t.Fatal(err)
	}

	err = st.WithTx(ctx, func(q *gen.Queries) error {
		r, err := combat.ApplyHeroVsHero(ctx, q, 1, 2)
		result = r
		return err
	})
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
}

// TestResolveOrderAtPlacesHeroAtMeetingNode is the targeted Patch 1
// guard: the new Scheduler.ResolveOrderAt(orderID, meetingNodeID) must
// place the hero at meetingNodeID, not at the order's original
// destination. tick.Collisions relies on this so combat resolves at the
// actual creep-encounter node instead of teleporting the hero past it
// (which is why no second battle ever fired after the first creep died
// — collisions consumed travel time without producing a battle).
func TestResolveOrderAtPlacesHeroAtMeetingNode(t *testing.T) {
	st, rdb, eng, _, _ := testE2E(t)
	defer st.Close()
	defer rdb.Close()

	ctx := context.Background()
	resetSimpleWorld(t, st, rdb)

	now := time.Now().UTC()
	order, err := st.Q.InsertMovementOrder(ctx, gen.InsertMovementOrderParams{
		HeroID:     1,
		FromNodeID: 1,
		ToNodeID:   5,
		DepartAt:   pgtype.Timestamptz{Time: now, Valid: true},
		ArriveAt:   pgtype.Timestamptz{Time: now.Add(60 * time.Second), Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}

	// Force-resolve at node 2 (Moss Crossing) — pretending a collision
	// happened mid-path. Hero must end up at node 2, not at node 5.
	if err := eng.Arrivals().ResolveOrderAt(ctx, order.ID, 2); err != nil {
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
	// original destination. Schedule another order and resolve it the
	// old way.
	now2 := time.Now().UTC()
	order2, err := st.Q.InsertMovementOrder(ctx, gen.InsertMovementOrderParams{
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
