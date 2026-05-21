package ws

import (
	"context"

	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5"
)

// BuyHandler processes unit.buy.
type BuyHandler struct {
	router *Router
}

func (h *BuyHandler) Handle(ctx context.Context, c *Client, env proto.Envelope) error {
	var req proto.UnitBuyPayload
	if err := env.DecodePayload(&req); err != nil {
		return h.router.sendError(c, proto.CodeHelloInvalidPayload, "invalid unit.buy payload", &env.Seq)
	}
	if req.CastleID <= 0 || req.UnitTypeID <= 0 || req.Qty <= 0 {
		return h.router.sendError(c, proto.CodeHelloInvalidPayload, "invalid buy parameters", &env.Seq)
	}

	castle, err := h.router.store.Q.GetCastleByPlayer(ctx, c.playerID)
	if err != nil {
		return err
	}
	if castle.ID != req.CastleID {
		return h.router.sendError(c, proto.CodeHelloInvalidPayload, "castle does not belong to player", &env.Seq)
	}

	unit, err := h.router.store.Q.GetUnit(ctx, req.UnitTypeID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return h.router.sendError(c, proto.CodeHelloInvalidPayload, "unknown unit type", &env.Seq)
		}
		return err
	}

	costGold := int64(unit.CostGold) * int64(req.Qty)
	costMetal := int64(unit.CostMetal) * int64(req.Qty)
	costGems := int64(unit.CostGems) * int64(req.Qty)
	costCoal := int64(unit.CostCoal) * int64(req.Qty)
	costWood := int64(unit.CostWood) * int64(req.Qty)
	costStone := int64(unit.CostStone) * int64(req.Qty)
	player, err := h.router.store.Q.GetPlayer(ctx, c.playerID)
	if err != nil {
		return err
	}
	gold, _ := player.Gold.Float64Value()
	metal, _ := player.Metal.Float64Value()
	gems, _ := player.Gems.Float64Value()
	coal, _ := player.Coal.Float64Value()
	wood, _ := player.Wood.Float64Value()
	stone, _ := player.Stone.Float64Value()
	if !gold.Valid || gold.Float64 < float64(costGold) ||
		!metal.Valid || metal.Float64 < float64(costMetal) ||
		!gems.Valid || gems.Float64 < float64(costGems) ||
		!coal.Valid || coal.Float64 < float64(costCoal) ||
		!wood.Valid || wood.Float64 < float64(costWood) ||
		!stone.Valid || stone.Float64 < float64(costStone) {
		return h.router.sendError(c, proto.CodeBuyInsufficientGold, "insufficient gold", &env.Seq)
	}

	if unit.Faction != "neutral" && unit.Faction != castle.Faction {
		return h.router.sendError(c, proto.CodeBuyWrongFaction, "unit faction mismatch", &env.Seq)
	}
	if unit.Tier > castle.BarracksTier {
		return h.router.sendError(c, proto.CodeBuyTierLocked, "unit tier locked", &env.Seq)
	}

	hero, err := h.router.store.Q.GetHeroByPlayer(ctx, c.playerID)
	if err != nil {
		return err
	}

	err = h.router.store.WithTx(ctx, func(q *gen.Queries) error {
		goldDelta, err := numericDelta(-float64(costGold))
		if err != nil {
			return err
		}
		metalDelta, err := numericDelta(-float64(costMetal))
		if err != nil {
			return err
		}
		gemsDelta, err := numericDelta(-float64(costGems))
		if err != nil {
			return err
		}
		coalDelta, err := numericDelta(-float64(costCoal))
		if err != nil {
			return err
		}
		woodDelta, err := numericDelta(-float64(costWood))
		if err != nil {
			return err
		}
		stoneDelta, err := numericDelta(-float64(costStone))
		if err != nil {
			return err
		}
		if err := q.IncrementPlayerResources(ctx, gen.IncrementPlayerResourcesParams{
			ID:         c.playerID,
			GoldDelta:  goldDelta,
			MetalDelta: metalDelta,
			GemsDelta:  gemsDelta,
			CoalDelta:  coalDelta,
			WoodDelta:  woodDelta,
			StoneDelta: stoneDelta,
		}); err != nil {
			return err
		}
		return q.AddHeroUnits(ctx, gen.AddHeroUnitsParams{
			HeroID: hero.ID,
			UnitID: req.UnitTypeID,
			Qty:    int32(req.Qty),
		})
	})
	if err != nil {
		return err
	}

	_ = h.router.broadcaster.BroadcastCastleTick(ctx, castle.ID, c.playerID, castle.GoldPerMin, true, env.Seq)
	_ = h.router.broadcaster.BroadcastResourceState(ctx, c.playerID, env.Seq)
	_ = h.router.broadcaster.BroadcastHeroState(ctx, hero.ID, env.Seq)
	return nil
}
