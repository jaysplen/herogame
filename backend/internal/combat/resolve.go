package combat

import "github.com/herogame/backend/internal/hero"

// Unit IDs that participate in the rock-paper-scissors counter triangle
// (docs/future_features.md §1). Keep in sync with migrations/00006_seed_units.sql.
const (
	UnitPikeman int64 = 1
	UnitArcher  int64 = 6
	UnitCavalry int64 = 7
)

// CounterBonusPercent is the extra damage (in percent) an attacker stack
// deals against the unit it counters: Pike→Cav, Cav→Arch, Arch→Pike.
const CounterBonusPercent = 25

// CounterMultiplier returns the damage multiplier applied to an attacker
// stack of attackerUnitID against a target whose primary unit is
// defenderUnitID. 1.25 inside the triangle, 1.0 everywhere else
// (including unknown unit IDs from older saves).
func CounterMultiplier(attackerUnitID, defenderUnitID int64) float64 {
	switch attackerUnitID {
	case UnitPikeman:
		if defenderUnitID == UnitCavalry {
			return 1.0 + float64(CounterBonusPercent)/100.0
		}
	case UnitArcher:
		if defenderUnitID == UnitPikeman {
			return 1.0 + float64(CounterBonusPercent)/100.0
		}
	case UnitCavalry:
		if defenderUnitID == UnitArcher {
			return 1.0 + float64(CounterBonusPercent)/100.0
		}
	}
	return 1.0
}

// PrimaryUnitID picks the unit ID that represents a side for counter math —
// the largest stack by quantity. Returns 0 if the side has no units.
func PrimaryUnitID(units []hero.StackUnit) int64 {
	var best int64
	bestQty := -1
	for _, u := range units {
		if u.Qty > bestQty {
			bestQty = u.Qty
			best = u.UnitID
		}
	}
	return best
}

// Side is a combatant stack aggregate (game_rules.md §6.2).
type Side struct {
	Attack  int
	Defense int
	HP      int
}

// LogEntry is one round event in the combat log.
type LogEntry struct {
	Round           int    `json:"round"`
	Side            string `json:"side"`
	Damage          int    `json:"damage"`
	DefenderHPAfter *int   `json:"defenderHpAfter,omitempty"`
	AttackerHPAfter *int   `json:"attackerHpAfter,omitempty"`
}

// Result is the outcome of resolveCombat (game_rules.md §6.3).
type Result struct {
	Outcome           string
	Log               []LogEntry
	InitialAttackerHP int
	FinalAttackerHP   int
	FinalDefenderHP   int
}

// StackFromHero builds a side aggregate without counter bonuses. Kept for
// callers (and tests) that don't know the opposing unit yet.
func StackFromHero(h hero.Stats, units []hero.StackUnit) Side {
	atk, def, hp := hero.Aggregate(h, units)
	return Side{Attack: atk, Defense: def, HP: hp}
}

// StackFromHeroVs is the counter-aware variant. Each attacker stack's damage
// contribution is multiplied by CounterMultiplier(stack, defenderPrimaryUnitID)
// before summation. Hero base attack and defense are not multiplied — only
// unit damage benefits from the triangle.
func StackFromHeroVs(h hero.Stats, units []hero.StackUnit, defenderPrimaryUnitID int64) Side {
	atk := h.Attack
	def := h.Defense
	hp := 0
	for _, u := range units {
		mult := CounterMultiplier(u.UnitID, defenderPrimaryUnitID)
		atk += int(float64(u.Attack*u.Qty) * mult)
		def += u.Defense * u.Qty
		hp += u.HP * u.Qty
	}
	return Side{Attack: atk, Defense: def, HP: hp}
}

// StackFromCreep builds the defender side from a single-unit creep stack
// without any counter multiplier.
func StackFromCreep(unitAtk, unitDef, unitHP, qty int) Side {
	return Side{
		Attack:  unitAtk * qty,
		Defense: unitDef * qty,
		HP:      unitHP * qty,
	}
}

// StackFromCreepVs is the counter-aware creep builder. If the creep's unit
// counters attackerPrimaryUnitID, its attack gets the +25% multiplier.
func StackFromCreepVs(creepUnitID int64, unitAtk, unitDef, unitHP, qty int, attackerPrimaryUnitID int64) Side {
	mult := CounterMultiplier(creepUnitID, attackerPrimaryUnitID)
	return Side{
		Attack:  int(float64(unitAtk*qty) * mult),
		Defense: unitDef * qty,
		HP:      unitHP * qty,
	}
}

// Resolve runs the deterministic round loop (game_rules.md §6.3).
func Resolve(attacker, defender Side) Result {
	initialHP := attacker.HP
	att := attacker
	def := defender
	log := make([]LogEntry, 0, 8)
	round := 0

	for att.HP > 0 && def.HP > 0 {
		round++
		dmgDef := max(1, att.Attack-def.Defense)
		def.HP -= dmgDef
		defAfter := max(0, def.HP)
		log = append(log, LogEntry{
			Round:           round,
			Side:            "attacker_hits",
			Damage:          dmgDef,
			DefenderHPAfter: &defAfter,
		})
		if def.HP <= 0 {
			break
		}

		dmgAtt := max(1, def.Attack-att.Defense)
		att.HP -= dmgAtt
		attAfter := max(0, att.HP)
		log = append(log, LogEntry{
			Round:           round,
			Side:            "defender_hits",
			Damage:          dmgAtt,
			AttackerHPAfter: &attAfter,
		})
	}

	outcome := "loss"
	if def.HP <= 0 {
		outcome = "win"
	}
	return Result{
		Outcome:           outcome,
		Log:               log,
		InitialAttackerHP: initialHP,
		FinalAttackerHP:   att.HP,
		FinalDefenderHP:   def.HP,
	}
}

// UnitsLost returns whole units removed from hp loss (game_rules.md §6.4).
func UnitsLost(initialHP, finalHP, unitHP int) int {
	if unitHP <= 0 {
		return 0
	}
	lost := initialHP - finalHP
	if lost <= 0 {
		return 0
	}
	return lost / unitHP
}
