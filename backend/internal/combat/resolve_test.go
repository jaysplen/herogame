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
