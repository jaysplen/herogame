package combat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/herogame/backend/internal/hero"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store/gen"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// ApplyResult is state to broadcast after combat commits.
type ApplyResult struct {
	Payload    proto.CombatResolvedPayload
	Respawn    bool
	PlayerID   int64
	CastleID   int64
	GoldPerMin int32
}

// ApplyAtNode resolves combat when an alive creep occupies nodeID (same DB tx as arrival).
func ApplyAtNode(ctx context.Context, q *gen.Queries, heroID, nodeID int64) (*ApplyResult, error) {
	creep, err := q.GetCreepByNode(ctx, nodeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	h, err := q.GetHero(ctx, heroID)
	if err != nil {
		return nil, err
	}
	rows, err := q.ListHeroUnitsByHero(ctx, heroID)
	if err != nil {
		return nil, err
	}
	unit, err := q.GetUnit(ctx, creep.UnitID)
	if err != nil {
		return nil, err
	}

	var stack []hero.StackUnit
	for _, row := range rows {
		stack = append(stack, hero.StackUnit{
			Qty:     int(row.Qty),
			Attack:  int(row.Attack),
			Defense: int(row.Defense),
			HP:      int(row.Hp),
		})
	}

	attacker := StackFromHero(hero.Stats{Attack: int(h.Attack), Defense: int(h.Defense)}, stack)
	defender := StackFromCreep(int(unit.Attack), int(unit.Defense), int(unit.Hp), int(creep.Qty))
	result := Resolve(attacker, defender)

	logJSON, err := json.Marshal(result.Log)
	if err != nil {
		return nil, err
	}

	var goldReward int32
	casualties := 0

	switch result.Outcome {
	case "win":
		goldReward = creep.GoldReward
		delta, err := numericDelta(float64(goldReward))
		if err != nil {
			return nil, err
		}
		if err := q.IncrementPlayerGold(ctx, gen.IncrementPlayerGoldParams{
			ID:    h.PlayerID,
			Delta: delta,
		}); err != nil {
			return nil, err
		}
		if err := q.SetCreepDead(ctx, creep.ID); err != nil {
			return nil, err
		}
		casualties, err = applyWinCasualties(ctx, q, heroID, rows, result)
		if err != nil {
			return nil, err
		}
	case "loss":
		castle, err := q.GetCastleByPlayer(ctx, h.PlayerID)
		if err != nil {
			return nil, err
		}
		casualties = hero.ArmySize(stack)
		if err := q.ClearHeroUnits(ctx, heroID); err != nil {
			return nil, err
		}
		if err := q.UpdateHeroNode(ctx, gen.UpdateHeroNodeParams{
			ID:            heroID,
			CurrentNodeID: castle.NodeID,
		}); err != nil {
			return nil, err
		}
	}

	_, err = q.InsertCombatLog(ctx, gen.InsertCombatLogParams{
		HeroID:     heroID,
		CreepID:    pgtype.Int8{Int64: creep.ID, Valid: true},
		Outcome:    result.Outcome,
		GoldReward: goldReward,
		Log:        logJSON,
	})
	if err != nil {
		return nil, err
	}

	castle, err := q.GetCastleByPlayer(ctx, h.PlayerID)
	if err != nil {
		return nil, err
	}

	return &ApplyResult{
		Payload: proto.CombatResolvedPayload{
			HeroID:     heroID,
			CreepID:    creep.ID,
			Outcome:    result.Outcome,
			GoldReward: goldReward,
			Casualties: casualties,
			Log:        toProtoLog(result.Log),
		},
		Respawn:    result.Outcome == "loss",
		PlayerID:   h.PlayerID,
		CastleID:   castle.ID,
		GoldPerMin: castle.GoldPerMin,
	}, nil
}

func applyWinCasualties(ctx context.Context, q *gen.Queries, heroID int64, rows []gen.ListHeroUnitsByHeroRow, result Result) (int, error) {
	if len(rows) == 0 {
		return 0, nil
	}
	unitHP := int(rows[0].Hp)
	toRemove := UnitsLost(result.InitialAttackerHP, max(0, result.FinalAttackerHP), unitHP)
	total := 0
	for _, row := range rows {
		if toRemove <= 0 {
			break
		}
		old := int(row.Qty)
		remove := toRemove
		if remove > old {
			remove = old
		}
		newQty := old - remove
		toRemove -= remove
		total += remove
		if err := q.SetHeroUnitQty(ctx, gen.SetHeroUnitQtyParams{
			HeroID: heroID,
			UnitID: row.UnitID,
			Qty:    int32(newQty),
		}); err != nil {
			return total, err
		}
	}
	return total, nil
}

func toProtoLog(entries []LogEntry) []proto.CombatLogEntry {
	out := make([]proto.CombatLogEntry, len(entries))
	for i, e := range entries {
		out[i] = proto.CombatLogEntry{
			Round:           e.Round,
			Side:            e.Side,
			Damage:          e.Damage,
			DefenderHPAfter: e.DefenderHPAfter,
			AttackerHPAfter: e.AttackerHPAfter,
		}
	}
	return out
}

func numericDelta(v float64) (pgtype.Numeric, error) {
	var n pgtype.Numeric
	if err := n.Scan(fmt.Sprintf("%f", v)); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}
