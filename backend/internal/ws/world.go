package ws

import (
	"context"

	"github.com/herogame/backend/internal/economy"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

func numericToFloat(n pgtype.Numeric) float64 {
	v, _ := n.Float64Value()
	if !v.Valid {
		return 0
	}
	return v.Float64
}

func buildResourceBag(p gen.GetPlayerRow) proto.ResourceBagDTO {
	return proto.ResourceBagDTO{
		Gold:  numericToFloat(p.Gold),
		Metal: numericToFloat(p.Metal),
		Gems:  numericToFloat(p.Gems),
		Coal:  numericToFloat(p.Coal),
		Wood:  numericToFloat(p.Wood),
		Stone: numericToFloat(p.Stone),
	}
}

func mapCastles(rows []gen.ListAllCastlesRow) []proto.CastleStateDTO {
	out := make([]proto.CastleStateDTO, 0, len(rows))
	for _, c := range rows {
		out = append(out, proto.CastleStateDTO{
			CastleID:     c.ID,
			PlayerID:     c.PlayerID,
			NodeID:       c.NodeID,
			Faction:      c.Faction,
			DefenseBonus: c.DefenseBonus,
			BarracksTier: c.BarracksTier,
			AcademyTier:  c.AcademyTier,
		})
	}
	return out
}

func optionalInt64(v pgtype.Int8) *int64 {
	if !v.Valid {
		return nil
	}
	x := v.Int64
	return &x
}

func optionalMs(v pgtype.Timestamptz) *int64 {
	if !v.Valid {
		return nil
	}
	ms := v.Time.UnixMilli()
	return &ms
}

func mapCreeps(rows []gen.NeutralCreep) []proto.CreepStateDTO {
	out := make([]proto.CreepStateDTO, 0, len(rows))
	for _, c := range rows {
		out = append(out, proto.CreepStateDTO{
			ID:         c.ID,
			Name:       c.Name,
			NodeID:     c.NodeID,
			Qty:        c.Qty,
			Alive:      c.Alive,
			Attack:     c.Attack,
			Defense:    c.Defense,
			HP:         c.Hp,
			GraceUntil: optionalMs(c.GraceUntil),
			FromNodeID: optionalInt64(c.FromNodeID),
			ToNodeID:   optionalInt64(c.ToNodeID),
			DepartAt:   optionalMs(c.DepartAt),
			ArriveAt:   optionalMs(c.ArriveAt),
		})
	}
	return out
}

func mapResourceNodes(rows []gen.ResourceNode) []proto.ResourceNodeDTO {
	out := make([]proto.ResourceNodeDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, proto.ResourceNodeDTO{
			ID:            r.ID,
			NodeID:        r.NodeID,
			ResourceType:  r.ResourceType,
			PerMin:        r.PerMin,
			OwnerPlayerID: optionalInt64(r.OwnerPlayerID),
		})
	}
	return out
}

func objectiveFor(playerID int64) proto.ObjectiveStatePayload {
	enemyID := int64(2)
	if playerID == 2 {
		enemyID = 1
	}
	return proto.ObjectiveStatePayload{
		PlayerID:        playerID,
		EnemyPlayerID:   enemyID,
		EnemyHeroKills:  0,
		TargetHeroKills: economy.ObjectiveHeroKills,
	}
}

func objectiveState(ctx context.Context, st *store.Store, playerID int64) (proto.ObjectiveStatePayload, error) {
	enemyID := int64(2)
	if playerID == 2 {
		enemyID = 1
	}
	kills, err := st.Q.GetPlayerKillCount(ctx, gen.GetPlayerKillCountParams{
		KillerPlayerID: playerID,
		VictimPlayerID: enemyID,
	})
	if err != nil && err != pgx.ErrNoRows {
		return proto.ObjectiveStatePayload{}, err
	}
	if err == pgx.ErrNoRows {
		kills = 0
	}
	return proto.ObjectiveStatePayload{
		PlayerID:        playerID,
		EnemyPlayerID:   enemyID,
		EnemyHeroKills:  kills,
		TargetHeroKills: economy.ObjectiveHeroKills,
	}, nil
}

func resourceState(ctx context.Context, st *store.Store, playerID int64) (proto.ResourceStatePayload, error) {
	player, err := st.Q.GetPlayer(ctx, playerID)
	if err != nil {
		return proto.ResourceStatePayload{}, err
	}
	nodes, err := st.Q.ListResourceNodes(ctx)
	if err != nil {
		return proto.ResourceStatePayload{}, err
	}
	return proto.ResourceStatePayload{
		PlayerID:      playerID,
		Resources:     buildResourceBag(player),
		ResourceNodes: mapResourceNodes(nodes),
	}, nil
}
