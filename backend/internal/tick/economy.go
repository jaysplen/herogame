package tick

import (
	"context"
	"log/slog"

	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
)

// Economy applies per-second castle income.
type Economy struct {
	store *store.Store
	log   *slog.Logger
}

// NewEconomy creates an economy sweeper.
func NewEconomy(st *store.Store, log *slog.Logger) *Economy {
	return &Economy{store: st, log: log}
}

// Sweep credits each castle owner gold_per_min / 60 per tick second.
func (e *Economy) Sweep(ctx context.Context) error {
	castles, err := e.store.Q.ListAllCastles(ctx)
	if err != nil {
		return err
	}
	for _, c := range castles {
		delta := float64(c.GoldPerMin) / 60.0
		goldN, err := numericFromFloat64(delta)
		if err != nil {
			return err
		}
		zero, err := numericFromFloat64(0)
		if err != nil {
			return err
		}
		if err := e.store.Q.IncrementPlayerResources(ctx, gen.IncrementPlayerResourcesParams{
			ID:         c.PlayerID,
			GoldDelta:  goldN,
			MetalDelta: zero,
			GemsDelta:  zero,
			CoalDelta:  zero,
			WoodDelta:  zero,
			StoneDelta: zero,
		}); err != nil {
			return err
		}
	}
	return nil
}
