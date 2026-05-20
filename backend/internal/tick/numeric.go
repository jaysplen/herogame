package tick

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"
)

func numericFromFloat64(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(fmt.Sprintf("%f", v)); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}
