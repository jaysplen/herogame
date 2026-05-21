package ws

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/herogame/backend/internal/world"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// MoveHandler processes move.request.
type MoveHandler struct {
	router *Router
}

func (h *MoveHandler) Handle(ctx context.Context, c *Client, env proto.Envelope) error {
	var req proto.MoveRequestPayload
	if err := env.DecodePayload(&req); err != nil {
		return h.router.sendError(c, proto.CodeHelloInvalidPayload, "invalid move.request payload", &env.Seq)
	}
	if req.HeroID <= 0 || req.TargetNodeID <= 0 {
		return h.router.sendError(c, proto.CodeHelloInvalidPayload, "invalid hero or target node", &env.Seq)
	}

	hero, err := h.router.store.Q.GetHero(ctx, req.HeroID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return h.router.sendError(c, proto.CodeHelloInvalidPayload, "hero not found", &env.Seq)
		}
		return err
	}
	if hero.PlayerID != c.playerID {
		return h.router.sendError(c, proto.CodeHelloInvalidPayload, "hero does not belong to player", &env.Seq)
	}

	if h.router.redis != nil {
		respawning, err := h.router.redis.IsRespawning(ctx, req.HeroID, time.Now().UTC())
		if err != nil {
			return err
		}
		if respawning {
			return h.router.sendError(c, proto.CodeMoveHeroRespawning, "hero is respawning", &env.Seq)
		}
	}
	if hero.SpawnGraceUntil.Valid && time.Now().UTC().Before(hero.SpawnGraceUntil.Time) {
		targetNode, err := h.router.store.Q.GetNode(ctx, req.TargetNodeID)
		if err != nil {
			return err
		}
		if strings.ToLower(targetNode.Kind) == "creep" {
			return h.router.sendError(c, proto.CodeMoveGracePeriod, "hero is in grace period", &env.Seq)
		}
	}

	_, err = h.router.store.Q.GetActiveMovementByHero(ctx, req.HeroID)
	if err == nil {
		return h.router.sendError(c, proto.CodeMoveHeroInFlight, "hero already in flight", &env.Seq)
	}
	if err != nil && err != pgx.ErrNoRows {
		return err
	}

	if hero.CurrentNodeID == req.TargetNodeID {
		return h.router.sendError(c, proto.CodeMoveInvalidEdge, "already at target node", &env.Seq)
	}

	edge, err := h.router.store.Q.GetEdge(ctx, gen.GetEdgeParams{
		FromNodeID: hero.CurrentNodeID,
		ToNodeID:   req.TargetNodeID,
	})
	if err != nil {
		if err == pgx.ErrNoRows {
			return h.router.sendError(c, proto.CodeMoveInvalidEdge, "no edge to target node", &env.Seq)
		}
		return err
	}

	units, err := h.router.store.Q.ListHeroUnitsByHero(ctx, req.HeroID)
	if err != nil {
		return err
	}
	armySize := 0
	for _, u := range units {
		armySize += int(u.Qty)
	}

	travelSec := world.TravelSeconds(int(edge.DistanceUnits), int(hero.BaseSpeed), armySize)
	now := time.Now().UTC()
	arriveAt := now.Add(time.Duration(travelSec) * time.Second)

	var order gen.MovementOrder
	err = h.router.store.WithTx(ctx, func(q *gen.Queries) error {
		o, err := q.InsertMovementOrder(ctx, gen.InsertMovementOrderParams{
			HeroID:      req.HeroID,
			FromNodeID:  hero.CurrentNodeID,
			ToNodeID:    req.TargetNodeID,
			DepartAt:    pgtype.Timestamptz{Time: now, Valid: true},
			ArriveAt:    pgtype.Timestamptz{Time: arriveAt, Valid: true},
		})
		if err != nil {
			return err
		}
		order = o
		return nil
	})
	if err != nil {
		return err
	}

	if h.router.arrivals == nil {
		return errors.New("arrivals scheduler not configured")
	}
	if err := h.router.arrivals.ScheduleAt(ctx, order.ID, arriveAt); err != nil {
		return err
	}

	if _, err := h.router.store.Q.GetResourceNodeByNode(ctx, req.TargetNodeID); err == nil {
		_ = h.router.store.WithTx(ctx, func(q *gen.Queries) error {
			return q.CaptureResourceNode(ctx, gen.CaptureResourceNodeParams{
				NodeID: req.TargetNodeID,
				OwnerPlayerID: pgtype.Int8{Int64: c.playerID, Valid: true},
			})
		})
	}

	h.router.broadcaster.BroadcastMoveUpdate(proto.MoveUpdatePayload{
		HeroID:        req.HeroID,
		FromNodeID:    hero.CurrentNodeID,
		ToNodeID:      req.TargetNodeID,
		DepartAt:      now.UnixMilli(),
		ArriveAt:      arriveAt.UnixMilli(),
		TravelSeconds: travelSec,
	}, env.Seq)
	_ = h.router.broadcaster.BroadcastResourceState(ctx, c.playerID, env.Seq)

	return nil
}
