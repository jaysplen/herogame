package hero

// Stats are flat hero attributes from the heroes row.
type Stats struct {
	Attack  int
	Defense int
}

// StackUnit is one row of army stack (qty of a unit type).
type StackUnit struct {
	Qty    int
	Attack int
	Defense int
	HP     int
}

// Aggregate computes side totals for combat (game_rules.md §6.2).
func Aggregate(h Stats, units []StackUnit) (attack, defense, hp int) {
	attack = h.Attack
	defense = h.Defense
	for _, u := range units {
		attack += u.Attack * u.Qty
		defense += u.Defense * u.Qty
		hp += u.HP * u.Qty
	}
	return attack, defense, hp
}

// ArmySize returns total unit count in the stack.
func ArmySize(units []StackUnit) int {
	n := 0
	for _, u := range units {
		n += u.Qty
	}
	return n
}
