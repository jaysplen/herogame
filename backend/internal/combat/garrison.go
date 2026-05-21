package combat

import (
	"context"

	"github.com/herogame/backend/internal/hero"
	"github.com/herogame/backend/internal/store/gen"
)

// HeroCombatStats returns hero stats with castle garrison defense when stationed at home.
func HeroCombatStats(ctx context.Context, q *gen.Queries, h gen.GetHeroRow) (hero.Stats, error) {
	stats := hero.Stats{Attack: int(h.Attack), Defense: int(h.Defense)}
	castle, err := q.GetCastleByPlayer(ctx, h.PlayerID)
	if err != nil {
		return stats, err
	}
	if h.CurrentNodeID == castle.NodeID {
		stats.Defense += int(castle.DefenseBonus)
	}
	return stats, nil
}
