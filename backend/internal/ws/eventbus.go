package ws

import (
	"context"

	"github.com/herogame/backend/internal/combat"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
)

// EventBus broadcasts WS events and post-combat follow-ups (castle.tick, hero.state).
type EventBus struct {
	hub   *Hub
	store *store.Store
	bc    *Broadcaster
}

// NewEventBus wires the hub and store for tick + gateway use.
func NewEventBus(hub *Hub, st *store.Store, rdb *redisx.Client) *EventBus {
	return &EventBus{
		hub:   hub,
		store: st,
		bc:    NewBroadcaster(hub, st, rdb),
	}
}

// Broadcast sends to all connected clients.
func (e *EventBus) Broadcast(env proto.Envelope) {
	if e.hub != nil {
		e.hub.Broadcast(env)
	}
}

// PostCombat emits combat.resolved plus balance-affecting hero/castle updates.
func (e *EventBus) PostCombat(ctx context.Context, r combat.ApplyResult, seq int64) error {
	env, err := proto.NewEnvelope(proto.TypeCombatResolved, r.Payload, seq)
	if err != nil {
		return err
	}
	e.Broadcast(env)
	_ = e.bc.BroadcastHeroState(ctx, r.Payload.HeroID, seq)
	return e.bc.BroadcastCastleTick(ctx, r.CastleID, r.PlayerID, r.GoldPerMin, true, seq)
}
