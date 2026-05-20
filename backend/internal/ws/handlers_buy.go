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

	cost := int64(unit.CostGold) * int64(req.Qty)
	player, err := h.router.store.Q.GetPlayer(ctx, c.playerID)
	if err != nil {
		return err
	}
	gold, _ := player.Gold.Float64Value()
	if !gold.Valid || gold.Float64 < float64(cost) {
		return h.router.sendError(c, proto.CodeBuyInsufficientGold, "insufficient gold", &env.Seq)
	}

	hero, err := h.router.store.Q.GetHeroByPlayer(ctx, c.playerID)
	if err != nil {
		return err
	}

	err = h.router.store.WithTx(ctx, func(q *gen.Queries) error {
		delta, err := numericDelta(-float64(cost))
		if err != nil {
			return err
		}
		if err := q.IncrementPlayerGold(ctx, gen.IncrementPlayerGoldParams{
			ID:    c.playerID,
			Delta: delta,
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
	_ = h.router.broadcaster.BroadcastHeroState(ctx, hero.ID, env.Seq)
	return nil
}
