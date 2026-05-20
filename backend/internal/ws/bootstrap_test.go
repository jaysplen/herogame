package ws_test

import (
	"context"
	"testing"
	"time"

	"github.com/herogame/backend/internal/store/gen"
	"github.com/herogame/backend/internal/ws"
	"github.com/jackc/pgx/v5/pgtype"
)

func TestHelloAckIncludesInFlightMovement(t *testing.T) {
	st := testStore(t)
	defer st.Close()
	ctx := context.Background()

	now := time.Now().UTC()
	arrive := now.Add(30 * time.Second)
	order, err := st.Q.InsertMovementOrder(ctx, gen.InsertMovementOrderParams{
		HeroID:     1,
		FromNodeID: 1,
		ToNodeID:   2,
		DepartAt:   pgtype.Timestamptz{Time: now, Valid: true},
		ArriveAt:   pgtype.Timestamptz{Time: arrive, Valid: true},
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = st.Pool().Exec(ctx, `DELETE FROM movement_orders WHERE id = $1`, order.ID)
	})

	ack, err := ws.BuildHelloAck(ctx, st, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if ack.InFlight == nil {
		t.Fatal("expected inFlight in hello.ack")
	}
	if ack.InFlight.HeroID != 1 || ack.InFlight.FromNodeID != 1 || ack.InFlight.ToNodeID != 2 {
		t.Fatalf("inFlight = %+v", ack.InFlight)
	}
	if ack.InFlight.DepartAt != now.UnixMilli() {
		t.Fatalf("departAt = %d, want %d", ack.InFlight.DepartAt, now.UnixMilli())
	}
}

func TestHelloAckNoInFlightWhenIdle(t *testing.T) {
	st := testStore(t)
	defer st.Close()
	ctx := context.Background()

	_, _ = st.Pool().Exec(ctx, `DELETE FROM movement_orders WHERE hero_id = 1 AND status = 'in_flight'`)

	ack, err := ws.BuildHelloAck(ctx, st, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if ack.InFlight != nil {
		t.Fatalf("expected no inFlight, got %+v", ack.InFlight)
	}
}
