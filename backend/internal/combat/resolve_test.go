package combat_test

import (
	"testing"

	"github.com/herogame/backend/internal/combat"
	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/hero"
)

func TestResolveWin(t *testing.T) {
	att := combat.StackFromHero(hero.Stats{Attack: 2, Defense: 2}, []hero.StackUnit{
		{Qty: 200, Attack: economy.PikemanAttack, Defense: economy.PikemanDefense, HP: economy.PikemanHP},
	})
	def := combat.StackFromCreep(economy.PikemanAttack, economy.PikemanDefense, economy.PikemanHP, 1)
	r := combat.Resolve(att, def)
	if r.Outcome != "win" {
		t.Fatalf("outcome = %q, want win", r.Outcome)
	}
	if len(r.Log) == 0 {
		t.Fatal("expected combat log entries")
	}
}

func TestResolveLossEmptyArmy(t *testing.T) {
	att := combat.StackFromHero(hero.Stats{Attack: 2, Defense: 2}, nil)
	def := combat.StackFromCreep(economy.PikemanAttack, economy.PikemanDefense, economy.PikemanHP, 50)
	r := combat.Resolve(att, def)
	if r.Outcome != "loss" {
		t.Fatalf("outcome = %q, want loss", r.Outcome)
	}
	if len(r.Log) != 0 {
		t.Fatalf("expected no rounds, got %d log entries", len(r.Log))
	}
}

func TestResolveTieBreakAttackerWins(t *testing.T) {
	// Both sides at 1 HP; attacker strikes first and ends combat before defender hits.
	att := combat.Side{Attack: 10, Defense: 0, HP: 1}
	def := combat.Side{Attack: 10, Defense: 0, HP: 1}
	r := combat.Resolve(att, def)
	if r.Outcome != "win" {
		t.Fatalf("outcome = %q, want win (attacker hits first)", r.Outcome)
	}
	if len(r.Log) != 1 || r.Log[0].Side != "attacker_hits" {
		t.Fatalf("log = %+v, want single attacker_hits entry", r.Log)
	}
}

func TestWorkedExampleRound1(t *testing.T) {
	att := combat.StackFromHero(hero.Stats{Attack: 2, Defense: 2}, []hero.StackUnit{
		{Qty: 30, Attack: 3, Defense: 2, HP: 10},
	})
	def := combat.StackFromCreep(3, 2, 10, 50)
	r := combat.Resolve(att, def)
	if len(r.Log) < 2 {
		t.Fatalf("expected at least 2 log entries, got %d", len(r.Log))
	}
	if r.Log[0].Round != 1 || r.Log[0].Side != "attacker_hits" || r.Log[0].Damage != 1 {
		t.Fatalf("round1 attacker: %+v", r.Log[0])
	}
	if r.Log[0].DefenderHPAfter == nil || *r.Log[0].DefenderHPAfter != 499 {
		t.Fatalf("defender hp after = %v, want 499", r.Log[0].DefenderHPAfter)
	}
	if r.Log[1].Side != "defender_hits" || r.Log[1].Damage != 88 {
		t.Fatalf("round1 defender: %+v", r.Log[1])
	}
	if r.Log[1].AttackerHPAfter == nil || *r.Log[1].AttackerHPAfter != 212 {
		t.Fatalf("attacker hp after = %v, want 212", r.Log[1].AttackerHPAfter)
	}
}

func TestWorkedExampleLoss(t *testing.T) {
	att := combat.StackFromHero(hero.Stats{Attack: 2, Defense: 2}, []hero.StackUnit{
		{Qty: 30, Attack: 3, Defense: 2, HP: 10},
	})
	def := combat.StackFromCreep(3, 2, 10, 50)
	r := combat.Resolve(att, def)
	if r.Outcome != "loss" {
		t.Fatalf("outcome = %q, want loss", r.Outcome)
	}
}

func TestUnitsLost(t *testing.T) {
	if got := combat.UnitsLost(300, 212, 10); got != 8 {
		t.Fatalf("units lost = %d, want 8", got)
	}
}

// Counter triangle: Pike->Cav, Cav->Arch, Arch->Pike at +25%
// (docs/future_features.md §1).
func TestCounterMultiplier(t *testing.T) {
	cases := []struct {
		name       string
		attackerID int64
		defenderID int64
		want       float64
	}{
		{"pike beats cav", combat.UnitPikeman, combat.UnitCavalry, 1.25},
		{"arch beats pike", combat.UnitArcher, combat.UnitPikeman, 1.25},
		{"cav beats arch", combat.UnitCavalry, combat.UnitArcher, 1.25},
		{"pike vs pike", combat.UnitPikeman, combat.UnitPikeman, 1.0},
		{"cav vs pike (countered)", combat.UnitCavalry, combat.UnitPikeman, 1.0},
		{"arch vs cav (countered)", combat.UnitArcher, combat.UnitCavalry, 1.0},
		{"unknown unit", 999, combat.UnitPikeman, 1.0},
		{"missing defender", combat.UnitPikeman, 0, 1.0},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := combat.CounterMultiplier(tc.attackerID, tc.defenderID)
			if got != tc.want {
				t.Fatalf("CounterMultiplier(%d, %d) = %v, want %v",
					tc.attackerID, tc.defenderID, got, tc.want)
			}
		})
	}
}

// PrimaryUnitID picks the largest stack by qty so per-side counter math has
// a deterministic "target identity".
func TestPrimaryUnitID(t *testing.T) {
	units := []hero.StackUnit{
		{UnitID: combat.UnitPikeman, Qty: 10},
		{UnitID: combat.UnitArcher, Qty: 25},
		{UnitID: combat.UnitCavalry, Qty: 3},
	}
	if got := combat.PrimaryUnitID(units); got != combat.UnitArcher {
		t.Fatalf("PrimaryUnitID = %d, want %d", got, combat.UnitArcher)
	}
	if got := combat.PrimaryUnitID(nil); got != 0 {
		t.Fatalf("PrimaryUnitID(nil) = %d, want 0", got)
	}
}

// 100 Archers vs 100 Pikemen primary: without counter, attack = 100*6 = 600.
// With +25%, attack = int(100*6*1.25) = 750. Defense and HP unchanged.
func TestStackFromHeroVsAppliesCounter(t *testing.T) {
	units := []hero.StackUnit{{UnitID: combat.UnitArcher, Qty: 100, Attack: 6, Defense: 2, HP: 8}}
	withCounter := combat.StackFromHeroVs(hero.Stats{Attack: 0, Defense: 0}, units, combat.UnitPikeman)
	noCounter := combat.StackFromHeroVs(hero.Stats{Attack: 0, Defense: 0}, units, combat.UnitArcher)
	if withCounter.Attack != 750 {
		t.Fatalf("counter attack = %d, want 750", withCounter.Attack)
	}
	if noCounter.Attack != 600 {
		t.Fatalf("non-counter attack = %d, want 600", noCounter.Attack)
	}
	if withCounter.Defense != 200 || withCounter.HP != 800 {
		t.Fatalf("def/hp untouched by counter: def=%d hp=%d", withCounter.Defense, withCounter.HP)
	}
}

// Hero base stats do NOT receive the counter multiplier — only unit damage does.
func TestStackFromHeroVsHeroBaseUnboosted(t *testing.T) {
	units := []hero.StackUnit{{UnitID: combat.UnitArcher, Qty: 10, Attack: 6, Defense: 2, HP: 8}}
	side := combat.StackFromHeroVs(hero.Stats{Attack: 7, Defense: 3}, units, combat.UnitPikeman)
	// 7 (hero) + int(10*6*1.25) = 7 + 75 = 82
	if side.Attack != 82 {
		t.Fatalf("attack = %d, want 82", side.Attack)
	}
	// hero 3 + 10*2 = 23
	if side.Defense != 23 {
		t.Fatalf("defense = %d, want 23", side.Defense)
	}
}

// StackFromCreepVs applies the counter multiplier when the creep counters the
// attacker's primary unit.
func TestStackFromCreepVs(t *testing.T) {
	// 50 Pikemen creep stack, attacker primary = Cavalry → Pike counters Cav.
	withCounter := combat.StackFromCreepVs(combat.UnitPikeman, 4, 5, 12, 50, combat.UnitCavalry)
	if withCounter.Attack != 250 { // int(4*50*1.25)
		t.Fatalf("creep counter attack = %d, want 250", withCounter.Attack)
	}
	// Same stack vs Archer attacker (Archer counters Pike, so Pike does NOT counter Archer)
	noCounter := combat.StackFromCreepVs(combat.UnitPikeman, 4, 5, 12, 50, combat.UnitArcher)
	if noCounter.Attack != 200 {
		t.Fatalf("creep non-counter attack = %d, want 200", noCounter.Attack)
	}
}
