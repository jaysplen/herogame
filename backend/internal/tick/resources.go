package tick

import (
	"context"
	"log/slog"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
)

// Resources applies resource node income every tick.
type Resources struct {
	store *store.Store
}

// NewResources creates resource income sweeper.
func NewResources(st *store.Store, _ *slog.Logger) *Resources {
	return &Resources{store: st}
}

func resourceDelta(kind string, perMin int32) economy.ResourceBag {
	perSec := float64(perMin) / 60.0
	switch kind {
	case "gold":
		return economy.ResourceBag{Gold: perSec}
	case "metal":
		return economy.ResourceBag{Metal: perSec}
	case "gems":
		return economy.ResourceBag{Gems: perSec}
	case "coal":
		return economy.ResourceBag{Coal: perSec}
	case "wood":
		return economy.ResourceBag{Wood: perSec}
	case "stone":
		return economy.ResourceBag{Stone: perSec}
	default:
		return economy.ResourceBag{}
	}
}

func (r *Resources) Sweep(ctx context.Context) error {
	nodes, err := r.store.Q.ListResourceNodes(ctx)
	if err != nil {
		return err
	}
	for _, n := range nodes {
		if !n.OwnerPlayerID.Valid {
			continue
		}
		bag := resourceDelta(n.ResourceType, n.PerMin)
		gold, metal, gems, coal, wood, stone, err := bag.ToNumerics()
		if err != nil {
			return err
		}
		err = r.store.Q.IncrementPlayerResources(ctx, gen.IncrementPlayerResourcesParams{
			ID:         n.OwnerPlayerID.Int64,
			GoldDelta:  gold,
			MetalDelta: metal,
			GemsDelta:  gems,
			CoalDelta:  coal,
			WoodDelta:  wood,
			StoneDelta: stone,
		})
		if err != nil {
			return err
		}
	}
	return nil
}

