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

// PostArrival emits hero.state after movement resolution (authoritative snapshot).
func (e *EventBus) PostArrival(ctx context.Context, heroID int64, seq int64) error {
	if heroID > 0 {
		if err := e.bc.BroadcastHeroState(ctx, heroID, seq); err != nil {
			return err
		}
	}
	if err := e.bc.BroadcastCreepState(ctx, seq); err != nil {
		return err
	}
	nodes, err := e.store.Q.ListResourceNodes(ctx)
	if err == nil {
		env, e2 := proto.NewEnvelope(proto.TypeResourceState, proto.ResourceStatePayload{
			PlayerID:      0,
			Resources:     proto.ResourceBagDTO{},
			ResourceNodes: mapResourceNodes(nodes),
		}, seq)
		if e2 == nil {
			e.Broadcast(env)
		}
	}
	return nil
}

// BroadcastCreepState emits creep positions for map interpolation.
func (e *EventBus) BroadcastCreepState(ctx context.Context, seq int64) error {
	return e.bc.BroadcastCreepState(ctx, seq)
}

// PostCombat emits combat.resolved plus balance-affecting hero/castle updates.
func (e *EventBus) PostCombat(ctx context.Context, r combat.ApplyResult, seq int64) error {
	env, err := proto.NewEnvelope(proto.TypeCombatResolved, r.Payload, seq)
	if err != nil {
		return err
	}
	e.Broadcast(env)
	_ = e.bc.BroadcastHeroState(ctx, r.Payload.HeroID, seq)
	_ = e.bc.BroadcastCastleTick(ctx, r.CastleID, r.PlayerID, r.GoldPerMin, true, seq)
	_ = e.bc.BroadcastResourceState(ctx, r.PlayerID, seq)
	_ = e.bc.BroadcastObjectiveState(ctx, r.PlayerID, seq)
	return e.bc.BroadcastCreepState(ctx, seq)
}
