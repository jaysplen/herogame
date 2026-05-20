package hero_test

import (
	"testing"

	"github.com/herogame/backend/internal/hero"
)

func TestAggregate(t *testing.T) {
	h := hero.Stats{Attack: 2, Defense: 2}
	units := []hero.StackUnit{
		{Qty: 30, Attack: 3, Defense: 2, HP: 10},
	}
	atk, def, hp := hero.Aggregate(h, units)
	if atk != 92 || def != 62 || hp != 300 {
		t.Fatalf("got atk=%d def=%d hp=%d", atk, def, hp)
	}
}
