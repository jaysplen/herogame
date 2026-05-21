package economy

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

// ResourceBag is an arithmetic helper for multi-resource economy.
type ResourceBag struct {
	Gold  float64
	Metal float64
	Gems  float64
	Coal  float64
	Wood  float64
	Stone float64
}

func numeric(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(fmt.Sprintf("%f", v)); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

// ToNumerics converts a ResourceBag into pgtype numerics.
func (b ResourceBag) ToNumerics() (gold, metal, gems, coal, wood, stone pgtype.Numeric, err error) {
	if gold, err = numeric(b.Gold); err != nil {
		return
	}
	if metal, err = numeric(b.Metal); err != nil {
		return
	}
	if gems, err = numeric(b.Gems); err != nil {
		return
	}
	if coal, err = numeric(b.Coal); err != nil {
		return
	}
	if wood, err = numeric(b.Wood); err != nil {
		return
	}
	if stone, err = numeric(b.Stone); err != nil {
		return
	}
	return
}
