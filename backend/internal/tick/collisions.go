package tick

import (
	"context"
	"log/slog"
	"time"

	"github.com/herogame/backend/internal/arrivals"
	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/world"
	"github.com/jackc/pgx/v5/pgtype"
)

// Collisions resolves in-flight hero movement when paths intersect roaming creeps.
type Collisions struct {
	store    *store.Store
	arrivals *arrivals.Scheduler
	log      *slog.Logger
}

// NewCollisions wires collision sweeper to the arrivals scheduler.
func NewCollisions(st *store.Store, sched *arrivals.Scheduler, log *slog.Logger) *Collisions {
	return &Collisions{store: st, arrivals: sched, log: log}
}

func pathPoint(
	now time.Time,
	depart, arrive pgtype.Timestamptz,
	ax, ay, bx, by float64,
) (float64, float64, bool) {
	if !depart.Valid || !arrive.Valid {
		return 0, 0, false
	}
	if now.Before(depart.Time) {
		return ax, ay, true
	}
	if !now.Before(arrive.Time) {
		return bx, by, true
	}
	span := arrive.Time.Sub(depart.Time).Seconds()
	if span <= 0 {
		return bx, by, true
	}
	t := now.Sub(depart.Time).Seconds() / span
	return ax + (bx-ax)*t, ay + (by-ay)*t, true
}

// Sweep checks hero segments vs creep segments and forces early arrival on hit.
func (c *Collisions) Sweep(ctx context.Context, now time.Time) error {
	orders, err := c.store.Q.ListInFlightMovements(ctx)
	if err != nil || len(orders) == 0 {
		return err
	}
	nodes, err := c.store.Q.ListNodes(ctx)
	if err != nil {
		return err
	}
	xy := make(map[int64]struct{ x, y float64 }, len(nodes))
	for _, n := range nodes {
		xy[n.ID] = struct{ x, y float64 }{float64(n.X), float64(n.Y)}
	}

	creeps, err := c.store.Q.ListAliveCreeps(ctx)
	if err != nil || len(creeps) == 0 {
		return err
	}

	for _, order := range orders {
		from, okFrom := xy[order.FromNodeID]
		to, okTo := xy[order.ToNodeID]
		if !okFrom || !okTo {
			continue
		}
		hx1, hy1, ok := pathPoint(now, order.DepartAt, order.ArriveAt, from.x, from.y, to.x, to.y)
		if !ok {
			continue
		}
		hx0, hy0, _ := pathPoint(now.Add(-time.Second), order.DepartAt, order.ArriveAt, from.x, from.y, to.x, to.y)

		for _, cr := range creeps {
			if cr.GraceUntil.Valid && now.Before(cr.GraceUntil.Time) {
				continue
			}
			cFrom, okCF := xy[cr.NodeID]
			if !okCF {
				continue
			}
			cx1, cy1 := cFrom.x, cFrom.y
			cx0, cy0 := cx1, cy1
			if cr.FromNodeID.Valid && cr.ToNodeID.Valid && cr.DepartAt.Valid && cr.ArriveAt.Valid {
				tFrom, okTF := xy[cr.FromNodeID.Int64]
				tTo, okTT := xy[cr.ToNodeID.Int64]
				if okTF && okTT {
					cx1, cy1, _ = pathPoint(now, cr.DepartAt, cr.ArriveAt, tFrom.x, tFrom.y, tTo.x, tTo.y)
					cx0, cy0, _ = pathPoint(now.Add(-time.Second), cr.DepartAt, cr.ArriveAt, tFrom.x, tFrom.y, tTo.x, tTo.y)
				}
			}
			d := world.SegmentDistance(hx0, hy0, hx1, hy1, cx0, cy0, cx1, cy1)
			if d > economy.CreepCollisionDistance {
				continue
			}
			if err := c.arrivals.ResolveOrder(ctx, order.ID); err != nil {
				c.log.Error("collision resolve", slog.Int64("order_id", order.ID), slog.String("error", err.Error()))
			}
			break
		}
	}
	return nil
}
