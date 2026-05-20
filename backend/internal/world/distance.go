package world

import (
	"math"
)

// UpkeepSlowdown returns the movement multiplier for a given army size.
// Formula: 1 + log2(max(1, armySize / 10)) — game_rules.md §4.1.
func UpkeepSlowdown(armySize int) float64 {
	if armySize < 0 {
		armySize = 0
	}
	ratio := math.Max(1, float64(armySize)/10)
	return 1 + math.Log2(ratio)
}

// TravelSeconds computes authoritative travel duration in whole seconds.
// Formula: ceil((distanceUnits / baseSpeed) * upkeepSlowdown(armySize)) — game_rules.md §3.
func TravelSeconds(distanceUnits, baseSpeed, armySize int) int {
	if baseSpeed <= 0 {
		baseSpeed = 1
	}
	if distanceUnits < 0 {
		distanceUnits = 0
	}
	raw := (float64(distanceUnits) / float64(baseSpeed)) * UpkeepSlowdown(armySize)
	return int(math.Ceil(raw))
}

// EffectiveSpeed is baseSpeed / upkeepSlowdown, for UI hints (game_rules.md §4.4).
func EffectiveSpeed(baseSpeed, armySize int) float64 {
	slow := UpkeepSlowdown(armySize)
	if slow <= 0 {
		return float64(baseSpeed)
	}
	return math.Round(float64(baseSpeed)/slow*100) / 100
}
