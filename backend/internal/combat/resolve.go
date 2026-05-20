package combat

import "github.com/herogame/backend/internal/hero"

// Side is a combatant stack aggregate (game_rules.md §6.2).
type Side struct {
	Attack  int
	Defense int
	HP      int
}

// LogEntry is one round event in the combat log.
type LogEntry struct {
	Round             int    `json:"round"`
	Side              string `json:"side"`
	Damage            int    `json:"damage"`
	DefenderHPAfter   *int   `json:"defenderHpAfter,omitempty"`
	AttackerHPAfter   *int   `json:"attackerHpAfter,omitempty"`
}

// Result is the outcome of resolveCombat (game_rules.md §6.3).
type Result struct {
	Outcome           string
	Log               []LogEntry
	InitialAttackerHP int
	FinalAttackerHP   int
	FinalDefenderHP   int
}

// StackFromHero builds the attacker side from hero stats and army rows.
func StackFromHero(h hero.Stats, units []hero.StackUnit) Side {
	atk, def, hp := hero.Aggregate(h, units)
	return Side{Attack: atk, Defense: def, HP: hp}
}

// StackFromCreep builds the defender side (no hero bonus on creeps).
func StackFromCreep(unitAtk, unitDef, unitHP, qty int) Side {
	return Side{
		Attack:  unitAtk * qty,
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
