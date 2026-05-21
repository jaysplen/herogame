package tick

import (
	"context"
	"log/slog"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
)

// Upkeep deducts army upkeep per second and triggers desertion when gold < 0.
type Upkeep struct {
	store *store.Store
	log   *slog.Logger
}

// NewUpkeep creates an upkeep sweeper.
func NewUpkeep(st *store.Store, log *slog.Logger) *Upkeep {
	return &Upkeep{store: st, log: log}
}

// Sweep deducts upkeep for every hero's army from the owning player.
func (u *Upkeep) Sweep(ctx context.Context) error {
	heroes, err := u.store.Q.ListHeroes(ctx)
	if err != nil {
		return err
	}
	for _, h := range heroes {
		units, err := u.store.Q.ListHeroUnitsByHero(ctx, h.ID)
		if err != nil {
			return err
		}
		var lines []economy.StackLine
		for _, row := range units {
			up, _ := row.UpkeepGoldPerHour.Float64Value()
			lines = append(lines, economy.StackLine{
				Qty:               int(row.Qty),
				UpkeepGoldPerHour: up.Float64,
			})
		}
		upkeepPerHour := economy.UpkeepGoldPerHour(lines)
		if upkeepPerHour <= 0 {
			continue
		}
		delta := -upkeepPerHour / 3600.0
		n, err := numericFromFloat64(delta)
		if err != nil {
			return err
		}
		zero, err := numericFromFloat64(0)
		if err != nil {
			return err
		}
		if err := u.store.Q.IncrementPlayerResources(ctx, gen.IncrementPlayerResourcesParams{
			ID:         h.PlayerID,
			GoldDelta:  n,
			MetalDelta: zero,
			GemsDelta:  zero,
			CoalDelta:  zero,
			WoodDelta:  zero,
			StoneDelta: zero,
		}); err != nil {
			return err
		}

		player, err := u.store.Q.GetPlayer(ctx, h.PlayerID)
		if err != nil {
			return err
		}
		gold, _ := player.Gold.Float64Value()
		if gold.Valid && gold.Float64 < 0 {
			u.log.Debug("player negative gold, desertion deferred to BETA-004",
				slog.Int64("player_id", h.PlayerID),
				slog.Float64("gold", gold.Float64),
			)
		}
	}
	return nil
}
