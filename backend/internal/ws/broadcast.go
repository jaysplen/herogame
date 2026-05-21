package ws

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5"
)

const castleTickMinInterval = 5 * time.Second

// Broadcaster sends server events with castle.tick throttling (architecture.md §7.3).
type Broadcaster struct {
	hub   *Hub
	store *store.Store
	redis *redisx.Client
	mu    sync.Mutex
	last  map[int64]time.Time // castleID -> last castle.tick
}

// NewBroadcaster creates a broadcaster.
func NewBroadcaster(hub *Hub, st *store.Store, rdb *redisx.Client) *Broadcaster {
	return &Broadcaster{hub: hub, store: st, redis: rdb, last: make(map[int64]time.Time)}
}

// Broadcast sends to all connected clients.
func (b *Broadcaster) Broadcast(env proto.Envelope) {
	if b.hub != nil {
		b.hub.Broadcast(env)
	}
}

// BroadcastToPlayer sends one envelope to one player's sessions.
func (b *Broadcaster) BroadcastToPlayer(playerID int64, env proto.Envelope) {
	if b.hub != nil {
		b.hub.BroadcastToPlayer(playerID, env)
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
	state, err := BuildHeroState(ctx, b.store, b.redis, heroID)
	if err != nil {
		return err
	}
	env, err := proto.NewEnvelope(proto.TypeHeroState, state, seq)
	if err != nil {
		return err
	}
	b.BroadcastToPlayer(state.PlayerID, env)
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

	resState, err := resourceState(ctx, b.store, playerID)
	if err != nil {
		return err
	}
	tickEnv, err := proto.NewEnvelope(proto.TypeCastleTick, proto.CastleTickPayload{
		CastleID:   castleID,
		Gold:       resState.Resources.Gold,
		GoldPerMin: goldPerMin,
	}, seq)
	if err != nil {
		return err
	}
	resourceEnv, err := proto.NewEnvelope(proto.TypeResourceState, resState, seq)
	if err == nil {
		b.BroadcastToPlayer(playerID, resourceEnv)
	}
	b.Broadcast(tickEnv)
	return nil
}

// BroadcastCreepState emits current neutral armies.
func (b *Broadcaster) BroadcastCreepState(ctx context.Context, seq int64) error {
	creeps, err := b.store.Q.ListAliveCreeps(ctx)
	if err != nil {
		return err
	}
	env, err := proto.NewEnvelope(proto.TypeCreepState, proto.CreepStatePayload{
		Creeps: mapCreeps(creeps),
	}, seq)
	if err != nil {
		return err
	}
	b.Broadcast(env)
	return nil
}

// BroadcastResourceState emits resource wallet and node ownership.
func (b *Broadcaster) BroadcastResourceState(ctx context.Context, playerID int64, seq int64) error {
	payload, err := resourceState(ctx, b.store, playerID)
	if err != nil {
		return err
	}
	env, err := proto.NewEnvelope(proto.TypeResourceState, payload, seq)
	if err != nil {
		return err
	}
	b.BroadcastToPlayer(playerID, env)
	return nil
}

// BroadcastObjectiveState sends elimination progress.
func (b *Broadcaster) BroadcastObjectiveState(ctx context.Context, playerID int64, seq int64) error {
	enemyID := int64(2)
	if playerID == 2 {
		enemyID = 1
	}
	kills, err := b.store.Q.GetPlayerKillCount(ctx, gen.GetPlayerKillCountParams{
		KillerPlayerID: playerID,
		VictimPlayerID: enemyID,
	})
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		kills = 0
	}
	if errors.Is(err, pgx.ErrNoRows) {
		kills = 0
	}
	env, err := proto.NewEnvelope(proto.TypeObjectiveState, proto.ObjectiveStatePayload{
		PlayerID:        playerID,
		EnemyPlayerID:   enemyID,
		EnemyHeroKills:  kills,
		TargetHeroKills: economy.ObjectiveHeroKills,
	}, seq)
	if err != nil {
		return err
	}
	b.BroadcastToPlayer(playerID, env)
	return nil
}
