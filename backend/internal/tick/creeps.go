package tick

import (
	"context"
	"log/slog"
	"time"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5/pgtype"
)

// Creeps advances roaming neutral armies and resolves coarse collisions.
type Creeps struct {
	store *store.Store
	log   *slog.Logger
}

// NewCreeps creates creep movement sweeper.
func NewCreeps(st *store.Store, log *slog.Logger) *Creeps {
	return &Creeps{store: st, log: log}
}

func validTime(v pgtype.Timestamptz) bool {
	return v.Valid && !v.Time.IsZero()
}

func chooseNeighbor(nodeID int64, edges []gen.MapEdge) int64 {
	for _, e := range edges {
		if e.FromNodeID == nodeID {
			return e.ToNodeID
		}
	}
	for _, e := range edges {
		if e.ToNodeID == nodeID {
			return e.FromNodeID
		}
	}
	return nodeID
}

func (c *Creeps) Sweep(ctx context.Context, now time.Time) error {
	creeps, err := c.store.Q.ListAliveCreeps(ctx)
	if err != nil {
		return err
	}
	edges, err := c.store.Q.ListEdges(ctx)
	if err != nil {
		return err
	}

	for _, cr := range creeps {
		nextNode := cr.NodeID
		to := cr.ToNodeID
		arriveAt := cr.ArriveAt

		if !validTime(arriveAt) || now.After(arriveAt.Time) {
			if to.Valid {
				nextNode = to.Int64
			}
			target := chooseNeighbor(nextNode, edges)
			travelSec := int64(economy.CreepSweepSeconds * 24)
			if cr.SpeedUnits > 0 {
				travelSec = int64(10 + (20 / cr.SpeedUnits))
			}
			if travelSec < 8 {
				travelSec = 8
			}
			depart := pgtype.Timestamptz{Time: now, Valid: true}
			arrive := pgtype.Timestamptz{Time: now.Add(time.Duration(travelSec) * time.Second), Valid: true}
			if err := c.store.Q.UpdateCreepMovement(ctx, gen.UpdateCreepMovementParams{
				ID:         cr.ID,
				NodeID:     nextNode,
				FromNodeID: pgtype.Int8{Int64: nextNode, Valid: true},
				ToNodeID:   pgtype.Int8{Int64: target, Valid: true},
				DepartAt:   depart,
				ArriveAt:   arrive,
			}); err != nil {
				return err
			}
			to = pgtype.Int8{Int64: target, Valid: true}
			arriveAt = arrive
		}
		_ = to
		_ = arriveAt
	}
	return nil
}
