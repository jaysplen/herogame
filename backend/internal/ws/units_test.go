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
	_, _ = st.Pool().Exec(ctx, `
		UPDATE castles SET barracks_tier = 2 WHERE player_id = 1;
	`)

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
	foundPikeman := false
	for _, s := range ack.ShopUnits {
		if s.Code == "pikeman" {
			foundPikeman = true
			if s.CostGold != 50 {
				t.Fatalf("pikeman shop row = %+v", s)
			}
		}
	}
	if !foundPikeman {
		t.Fatalf("shop missing pikeman: %+v", ack.ShopUnits)
	}
}
