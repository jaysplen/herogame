package economy

// StackLine is one hero_units row with catalog upkeep rate.
type StackLine struct {
	Qty                 int
	UpkeepGoldPerHour   float64
}

// UpkeepGoldPerHour sums qty * upkeep per hour (game_rules.md §5.1).
func UpkeepGoldPerHour(lines []StackLine) float64 {
	var total float64
	for _, l := range lines {
		total += float64(l.Qty) * l.UpkeepGoldPerHour
	}
	return total
}

// DeltaGoldPerSecond returns net gold change per tick second:
// castle income (goldPerMin/60) minus upkeep proration (upkeepGph/3600).
func DeltaGoldPerSecond(upkeepGph float64, goldPerMin int) float64 {
	income := float64(goldPerMin) / 60.0
	upkeep := upkeepGph / 3600.0
	return income - upkeep
}
