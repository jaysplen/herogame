package ws

import (
	"context"
	"strings"

	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5"
)

// BuildHandler processes castle.build.
type BuildHandler struct {
	router *Router
}

func buildCost(code string) (gold, wood, stone int32) {
	switch code {
	case "defense":
		return 120, 20, 25
	case "barracks":
		return 160, 30, 20
	case "academy":
		return 190, 20, 35
	default:
		return 0, 0, 0
	}
}

func (h *BuildHandler) Handle(ctx context.Context, c *Client, env proto.Envelope) error {
	var req proto.CastleBuildPayload
	if err := env.DecodePayload(&req); err != nil {
		return h.router.sendError(c, proto.CodeHelloInvalidPayload, "invalid castle.build payload", &env.Seq)
	}
	code := strings.TrimSpace(strings.ToLower(req.BuildingCode))
	costGold, costWood, costStone := buildCost(code)
	if req.CastleID <= 0 || code == "" || costGold <= 0 {
		return h.router.sendError(c, proto.CodeBuildInvalid, "invalid build request", &env.Seq)
	}

	castle, err := h.router.store.Q.GetCastleByPlayer(ctx, c.playerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return h.router.sendError(c, proto.CodeBuildInvalid, "castle not found", &env.Seq)
		}
		return err
	}
	if castle.ID != req.CastleID {
		return h.router.sendError(c, proto.CodeBuildInvalid, "castle ownership mismatch", &env.Seq)
	}

	player, err := h.router.store.Q.GetPlayer(ctx, c.playerID)
	if err != nil {
		return err
	}
	gold, _ := player.Gold.Float64Value()
	wood, _ := player.Wood.Float64Value()
	stone, _ := player.Stone.Float64Value()
	if !gold.Valid || gold.Float64 < float64(costGold) ||
		!wood.Valid || wood.Float64 < float64(costWood) ||
		!stone.Valid || stone.Float64 < float64(costStone) {
		return h.router.sendError(c, proto.CodeBuildInsufficientResources, "insufficient resources", &env.Seq)
	}

	err = h.router.store.WithTx(ctx, func(q *gen.Queries) error {
		goldDelta, err := numericDelta(-float64(costGold))
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
		zero, err := numericDelta(0)
		if err != nil {
			return err
		}
		if err := q.IncrementPlayerResources(ctx, gen.IncrementPlayerResourcesParams{
			ID:         c.playerID,
			GoldDelta:  goldDelta,
			MetalDelta: zero,
			GemsDelta:  zero,
			CoalDelta:  zero,
			WoodDelta:  woodDelta,
			StoneDelta: stoneDelta,
		}); err != nil {
			return err
		}
		return q.UpgradeCastleBuilding(ctx, gen.UpgradeCastleBuildingParams{
			BuildingCode: code,
			CastleID:     req.CastleID,
		})
	})
	if err != nil {
		return err
	}
	_ = h.router.broadcaster.BroadcastResourceState(ctx, c.playerID, env.Seq)
	_ = h.router.broadcaster.BroadcastObjectiveState(ctx, c.playerID, env.Seq)
	return nil
}
