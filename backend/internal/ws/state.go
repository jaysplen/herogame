package ws

import (
	"context"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/herogame/backend/internal/world"
)

// BuildHeroState loads hero state for broadcasts (includes Redis respawn lockout).
func BuildHeroState(ctx context.Context, st *store.Store, rdb *redisx.Client, heroID int64) (proto.HeroStatePayload, error) {
	h, err := st.Q.GetHero(ctx, heroID)
	if err != nil {
		return proto.HeroStatePayload{}, err
	}
	units, err := st.Q.ListHeroUnitsByHero(ctx, heroID)
	if err != nil {
		return proto.HeroStatePayload{}, err
	}
	armySize := 0
	var stack []economy.StackLine
	for _, u := range units {
		armySize += int(u.Qty)
		up, _ := u.UpkeepGoldPerHour.Float64Value()
		stack = append(stack, economy.StackLine{
			Qty:               int(u.Qty),
			UpkeepGoldPerHour: up.Float64,
		})
	}

	state := proto.HeroStatePayload{
		HeroID:            h.ID,
		CurrentNodeID:     h.CurrentNodeID,
		ArmySize:          armySize,
		UpkeepGoldPerHour: economy.UpkeepGoldPerHour(stack),
		SpeedEffective:    world.EffectiveSpeed(int(h.BaseSpeed), armySize),
	}

	if rdb != nil {
		if untilMs, ok, err := rdb.RespawnUntilMs(ctx, heroID); err != nil {
			return proto.HeroStatePayload{}, err
		} else if ok {
			state.RespawnUntil = &untilMs
		}
	}

	return state, nil
}

// PlayerGold returns the player's gold as float64.
func PlayerGold(ctx context.Context, q *gen.Queries, playerID int64) (float64, error) {
	p, err := q.GetPlayer(ctx, playerID)
	if err != nil {
		return 0, err
	}
	g, _ := p.Gold.Float64Value()
	return g.Float64, nil
}
