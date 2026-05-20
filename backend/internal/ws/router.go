package ws

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/arrivals"
	"github.com/jackc/pgx/v5"
)

// Router dispatches inbound envelopes after handshake.
type Router struct {
	store       *store.Store
	redis       *redisx.Client
	arrivals    *arrivals.Scheduler
	broadcaster *Broadcaster
	move        *MoveHandler
	buy         *BuyHandler
}

// NewRouter creates a message router.
func NewRouter(st *store.Store, rdb *redisx.Client, sched *arrivals.Scheduler, hub *Hub) *Router {
	r := &Router{
		store:       st,
		redis:       rdb,
		arrivals:    sched,
		broadcaster: NewBroadcaster(hub, st, rdb),
	}
	r.move = &MoveHandler{router: r}
	r.buy = &BuyHandler{router: r}
	return r
}

// Handle processes a post-handshake client message.
func (r *Router) Handle(ctx context.Context, c *Client, env proto.Envelope) error {
	switch env.Type {
	case proto.TypeMoveRequest:
		return r.move.Handle(ctx, c, env)
	case proto.TypeUnitBuy:
		return r.buy.Handle(ctx, c, env)
	default:
		return r.sendError(c, proto.CodeUnknownMessage,
			"unknown message type "+env.Type, &env.Seq)
	}
}

// HandleHello validates hello and returns the ack envelope bytes.
func (r *Router) HandleHello(ctx context.Context, env proto.Envelope) ([]byte, int64, error) {
	var payload proto.HelloPayload
	if err := env.DecodePayload(&payload); err != nil {
		return nil, 0, err
	}
	if payload.PlayerID <= 0 {
		return nil, 0, errors.New("invalid playerId")
	}

	ack, err := BuildHelloAck(ctx, r.store, r.redis, payload.PlayerID)
	if err != nil {
		if errors.Is(err, errUnknownPlayer) || errors.Is(err, pgx.ErrNoRows) {
			errEnv, _ := proto.NewEnvelope(proto.TypeError,
				proto.NewErrorPayload(proto.CodeHelloUnknownPlayer, "unknown player", &env.Seq),
				env.Seq)
			b, _ := json.Marshal(errEnv)
			return b, 0, errUnknownPlayer
		}
		return nil, 0, err
	}

	ackEnv, err := proto.NewEnvelope(proto.TypeHelloAck, ack, env.Seq)
	if err != nil {
		return nil, 0, err
	}
	data, err := json.Marshal(ackEnv)
	return data, payload.PlayerID, err
}

func (r *Router) sendError(c *Client, code, message string, refSeq *int64) error {
	env, err := proto.NewEnvelope(proto.TypeError, proto.NewErrorPayload(code, message, refSeq), 0)
	if err != nil {
		return err
	}
	data, err := json.Marshal(env)
	if err != nil {
		return err
	}
	c.enqueue(data)
	return nil
}
