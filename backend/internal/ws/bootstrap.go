package ws

import (
	"context"
	"fmt"

	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5"
)

// BuildHelloAck loads the PoC bootstrap snapshot for a player.
func BuildHelloAck(ctx context.Context, st *store.Store, rdb *redisx.Client, playerID int64) (proto.HelloAckPayload, error) {
	player, err := st.Q.GetPlayer(ctx, playerID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return proto.HelloAckPayload{}, errUnknownPlayer
		}
		return proto.HelloAckPayload{}, err
	}

	h, err := st.Q.GetHeroByPlayer(ctx, playerID)
	if err != nil {
		return proto.HelloAckPayload{}, fmt.Errorf("hero for player %d: %w", playerID, err)
	}

	castle, err := st.Q.GetCastleByPlayer(ctx, playerID)
	if err != nil {
		return proto.HelloAckPayload{}, fmt.Errorf("castle for player %d: %w", playerID, err)
	}

	nodes, err := st.Q.ListNodes(ctx)
	if err != nil {
		return proto.HelloAckPayload{}, err
	}
	edges, err := st.Q.ListEdges(ctx)
	if err != nil {
		return proto.HelloAckPayload{}, err
	}

	gold, _ := player.Gold.Float64Value()

	inFlight, err := activeMovePayload(ctx, st, h.ID)
	if err != nil {
		return proto.HelloAckPayload{}, err
	}

	heroState, err := BuildHeroState(ctx, st, rdb, h.ID)
	if err != nil {
		return proto.HelloAckPayload{}, err
	}

	return proto.HelloAckPayload{
		PlayerID: playerID,
		HeroID:   h.ID,
		CastleID: castle.ID,
		Gold:     gold.Float64,
		MapSnapshot: proto.MapSnapshot{
			Nodes: mapNodesToDTO(nodes),
			Edges: mapEdgesToDTO(edges),
		},
		HeroState: heroState,
		InFlight:  inFlight,
	}, nil
}

func activeMovePayload(ctx context.Context, st *store.Store, heroID int64) (*proto.MoveUpdatePayload, error) {
	order, err := st.Q.GetActiveMovementByHero(ctx, heroID)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if !order.DepartAt.Valid || !order.ArriveAt.Valid {
		return nil, nil
	}
	travelSec := int(order.ArriveAt.Time.Sub(order.DepartAt.Time).Seconds())
	if travelSec < 1 {
		travelSec = 1
	}
	return &proto.MoveUpdatePayload{
		HeroID:        order.HeroID,
		FromNodeID:    order.FromNodeID,
		ToNodeID:      order.ToNodeID,
		DepartAt:      order.DepartAt.Time.UnixMilli(),
		ArriveAt:      order.ArriveAt.Time.UnixMilli(),
		TravelSeconds: travelSec,
	}, nil
}

func mapNodesToDTO(nodes []gen.MapNode) []proto.MapNodeDTO {
	out := make([]proto.MapNodeDTO, len(nodes))
	for i, n := range nodes {
		out[i] = proto.MapNodeDTO{
			ID: n.ID, Name: n.Name, X: n.X, Y: n.Y, Kind: n.Kind,
		}
	}
	return out
}

func mapEdgesToDTO(edges []gen.MapEdge) []proto.MapEdgeDTO {
	out := make([]proto.MapEdgeDTO, len(edges))
	for i, e := range edges {
		out[i] = proto.MapEdgeDTO{
			ID: e.ID, FromNodeID: e.FromNodeID, ToNodeID: e.ToNodeID,
			DistanceUnits: e.DistanceUnits,
		}
	}
	return out
}

// errUnknownPlayer is returned when hello references a missing player.
var errUnknownPlayer = fmt.Errorf("unknown player")
