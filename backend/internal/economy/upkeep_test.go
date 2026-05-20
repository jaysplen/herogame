package economy_test

import (
	"math"
	"testing"

	"github.com/herogame/backend/internal/economy"
)

func TestUpkeepGoldPerHour(t *testing.T) {
	lines := []economy.StackLine{
		{Qty: 50, UpkeepGoldPerHour: 2.0},
	}
	got := economy.UpkeepGoldPerHour(lines)
	if got != 100 {
		t.Fatalf("got %v", got)
	}
}

func TestDeltaGoldPerSecond(t *testing.T) {
	// 60 g/min income, 100 g/hr upkeep → net positive
	net := economy.DeltaGoldPerSecond(100, 60)
	want := 1.0 - 100.0/3600.0
	if math.Abs(net-want) > 0.0001 {
		t.Fatalf("got %v want %v", net, want)
	}
}
