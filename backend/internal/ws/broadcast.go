package ws

import (
	"context"
	"sync"
	"time"

	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store"
)

const castleTickMinInterval = 5 * time.Second

// Broadcaster sends server events with castle.tick throttling (architecture.md §7.3).
type Broadcaster struct {
	hub   *Hub
	store *store.Store
	mu    sync.Mutex
	last  map[int64]time.Time // castleID -> last castle.tick
}

// NewBroadcaster creates a broadcaster.
func NewBroadcaster(hub *Hub, st *store.Store) *Broadcaster {
	return &Broadcaster{hub: hub, store: st, last: make(map[int64]time.Time)}
}

// Broadcast sends to all connected clients.
func (b *Broadcaster) Broadcast(env proto.Envelope) {
	if b.hub != nil {
		b.hub.Broadcast(env)
	}
}

// BroadcastMoveUpdate emits move.update to all clients.
func (b *Broadcaster) BroadcastMoveUpdate(payload proto.MoveUpdatePayload, seq int64) {
	env, err := proto.NewEnvelope(proto.TypeMoveUpdate, payload, seq)
	if err == nil {
		b.Broadcast(env)
	}
}

// BroadcastHeroState emits hero.state for a hero.
func (b *Broadcaster) BroadcastHeroState(ctx context.Context, heroID int64, seq int64) error {
	state, err := BuildHeroState(ctx, b.store, heroID)
	if err != nil {
		return err
	}
	env, err := proto.NewEnvelope(proto.TypeHeroState, state, seq)
	if err != nil {
		return err
	}
	b.Broadcast(env)
	return nil
}

// BroadcastCastleTick emits castle.tick if throttle allows, or if force is true.
func (b *Broadcaster) BroadcastCastleTick(ctx context.Context, castleID, playerID int64, goldPerMin int32, force bool, seq int64) error {
	b.mu.Lock()
	last := b.last[castleID]
	if !force && time.Since(last) < castleTickMinInterval {
		b.mu.Unlock()
		return nil
	}
	b.last[castleID] = time.Now()
	b.mu.Unlock()

	gold, err := PlayerGold(ctx, b.store.Q, playerID)
	if err != nil {
		return err
	}
	env, err := proto.NewEnvelope(proto.TypeCastleTick, proto.CastleTickPayload{
		CastleID:   castleID,
		Gold:       gold,
		GoldPerMin: goldPerMin,
	}, seq)
	if err != nil {
		return err
	}
	b.Broadcast(env)
	return nil
}
