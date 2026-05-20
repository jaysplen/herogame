package ws_test

import (
	"context"
	"testing"

	"github.com/herogame/backend/internal/ws"
)

func TestHelloAckIncludesUnitStacks(t *testing.T) {
	st := testStore(t)
	defer st.Close()
	ctx := context.Background()

	ack, err := ws.BuildHelloAck(ctx, st, nil, 1)
	if err != nil {
		t.Fatal(err)
	}
	if ack.HeroState.Units == nil {
		t.Fatal("expected units slice")
	}
	if len(ack.ShopUnits) == 0 {
		t.Fatal("expected shop catalog")
	}
	if ack.ShopUnits[0].Code != "pikeman" || ack.ShopUnits[0].CostGold != 50 {
		t.Fatalf("shop = %+v", ack.ShopUnits[0])
	}
}
