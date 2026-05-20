package arrivals

import (
	"context"
	"log/slog"
	"time"

	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
)

const batchLimit = 100

// Broadcaster pushes server events to connected clients.
type Broadcaster interface {
	Broadcast(env proto.Envelope)
}

// Scheduler manages Redis ZSET arrival timing and resolution.
type Scheduler struct {
	store *store.Store
	redis *redisx.Client
	bus   Broadcaster
	log   *slog.Logger
}

// New creates an arrivals scheduler.
func New(st *store.Store, rdb *redisx.Client, bus Broadcaster, log *slog.Logger) *Scheduler {
	return &Scheduler{store: st, redis: rdb, bus: bus, log: log}
}

// Rehydrate rebuilds Redis from Postgres in-flight orders.
func (a *Scheduler) Rehydrate(ctx context.Context) error {
	orders, err := a.store.Q.ListInFlightMovements(ctx)
	if err != nil {
		return err
	}
	entries := make([]redisx.ArrivalEntry, 0, len(orders))
	for _, o := range orders {
		if !o.ArriveAt.Valid {
			continue
		}
		entries = append(entries, redisx.ArrivalEntry{
			OrderID:  o.ID,
			ArriveAt: o.ArriveAt.Time,
		})
	}
	if err := a.redis.Rehydrate(ctx, entries); err != nil {
		return err
	}
	a.log.Info("arrivals rehydrated", slog.Int("count", len(entries)))
	return nil
}

// ScheduleAt adds an arrival to the Redis ZSET.
func (a *Scheduler) ScheduleAt(ctx context.Context, orderID int64, arriveAt time.Time) error {
	return a.redis.AddArrival(ctx, orderID, arriveAt)
}

// Sweep resolves due arrivals (architecture.md §6).
func (a *Scheduler) Sweep(ctx context.Context, now time.Time) error {
	ids, err := a.redis.DueArrivals(ctx, now, batchLimit)
	if err != nil {
		return err
	}
	for _, orderID := range ids {
		if err := a.resolveOne(ctx, orderID); err != nil {
			a.log.Error("resolve arrival", slog.Int64("order_id", orderID), slog.String("error", err.Error()))
		}
	}
	return nil
}

func (a *Scheduler) resolveOne(ctx context.Context, orderID int64) error {
	var resolved gen.MovementOrder

	err := a.store.WithTx(ctx, func(q *gen.Queries) error {
		n, err := q.MarkMovementArrived(ctx, orderID)
		if err != nil {
			return err
		}
		if n == 0 {
			return errNoop
		}
		order, err := q.GetMovementOrder(ctx, orderID)
		if err != nil {
			return err
		}
		if err := q.UpdateHeroNode(ctx, gen.UpdateHeroNodeParams{
			ID:            order.HeroID,
			CurrentNodeID: order.ToNodeID,
		}); err != nil {
			return err
		}
		resolved = order
		return nil
	})
	if err == errNoop {
		_ = a.redis.RemoveArrival(ctx, orderID)
		return nil
	}
	if err != nil {
		return err
	}

	if err := a.redis.RemoveArrival(ctx, orderID); err != nil {
		return err
	}

	if a.bus != nil {
		env, err := proto.NewEnvelope(proto.TypeMoveArrived, proto.MoveArrivedPayload{
			HeroID: resolved.HeroID,
			NodeID: resolved.ToNodeID,
		}, 0)
		if err == nil {
			a.bus.Broadcast(env)
		}
	}

	a.log.Info("movement arrived",
		slog.Int64("order_id", orderID),
		slog.Int64("hero_id", resolved.HeroID),
		slog.Int64("node_id", resolved.ToNodeID),
	)
	return nil
}

var errNoop = errNoopSentinel{}

type errNoopSentinel struct{}

func (errNoopSentinel) Error() string { return "arrival already resolved" }
