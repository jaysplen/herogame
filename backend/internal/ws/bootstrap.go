package ws

import (
	"context"
	"fmt"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/herogame/backend/internal/world"
	"github.com/jackc/pgx/v5"
)

// BuildHelloAck loads the PoC bootstrap snapshot for a player.
func BuildHelloAck(ctx context.Context, st *store.Store, playerID int64) (proto.HelloAckPayload, error) {
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

	units, err := st.Q.ListHeroUnitsByHero(ctx, h.ID)
	if err != nil {
		return proto.HelloAckPayload{}, err
	}

	armySize := 0
	var stack []economy.StackLine
	for _, u := range units {
		armySize += int(u.Qty)
		upkeep, _ := u.UpkeepGoldPerHour.Float64Value()
		stack = append(stack, economy.StackLine{
			Qty:               int(u.Qty),
			UpkeepGoldPerHour: upkeep.Float64,
		})
	}

	gold, _ := player.Gold.Float64Value()
	upkeepGph := economy.UpkeepGoldPerHour(stack)

	return proto.HelloAckPayload{
		PlayerID: playerID,
		HeroID:   h.ID,
		CastleID: castle.ID,
		Gold:     gold.Float64,
		MapSnapshot: proto.MapSnapshot{
			Nodes: mapNodesToDTO(nodes),
			Edges: mapEdgesToDTO(edges),
		},
		HeroState: proto.HeroStatePayload{
			HeroID:            h.ID,
			CurrentNodeID:     h.CurrentNodeID,
			ArmySize:          armySize,
			UpkeepGoldPerHour: upkeepGph,
			SpeedEffective:    world.EffectiveSpeed(int(h.BaseSpeed), armySize),
		},
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
