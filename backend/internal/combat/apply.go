package combat

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/herogame/backend/internal/hero"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/economy"
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

// ApplyHeroVsHero resolves one hero army attacking another on the same node.
func ApplyHeroVsHero(ctx context.Context, q *gen.Queries, attackerHeroID, defenderHeroID int64) (*ApplyResult, error) {
	if attackerHeroID == defenderHeroID {
		return nil, nil
	}
	attHero, err := q.GetHero(ctx, attackerHeroID)
	if err != nil {
		return nil, err
	}
	defHero, err := q.GetHero(ctx, defenderHeroID)
	if err != nil {
		return nil, err
	}
	if attHero.CurrentNodeID != defHero.CurrentNodeID {
		return nil, nil
	}

	attRows, err := q.ListHeroUnitsByHero(ctx, attackerHeroID)
	if err != nil {
		return nil, err
	}
	defRows, err := q.ListHeroUnitsByHero(ctx, defenderHeroID)
	if err != nil {
		return nil, err
	}

	var attStack []hero.StackUnit
	for _, row := range attRows {
		attStack = append(attStack, hero.StackUnit{
			UnitID: row.UnitID,
			Qty:    int(row.Qty), Attack: int(row.Attack), Defense: int(row.Defense), HP: int(row.Hp),
		})
	}
	var defStack []hero.StackUnit
	for _, row := range defRows {
		defStack = append(defStack, hero.StackUnit{
			UnitID: row.UnitID,
			Qty:    int(row.Qty), Attack: int(row.Attack), Defense: int(row.Defense), HP: int(row.Hp),
		})
	}

	attStats, err := HeroCombatStats(ctx, q, attHero)
	if err != nil {
		return nil, err
	}
	defStats, err := HeroCombatStats(ctx, q, defHero)
	if err != nil {
		return nil, err
	}
	// Rock-paper-scissors counters use each side's primary (largest) stack as
	// the target identity (docs/future_features.md §1).
	attPrimary := PrimaryUnitID(attStack)
	defPrimary := PrimaryUnitID(defStack)
	attacker := StackFromHeroVs(attStats, attStack, defPrimary)
	defender := StackFromHeroVs(defStats, defStack, attPrimary)
	result := Resolve(attacker, defender)

	logJSON, err := json.Marshal(result.Log)
	if err != nil {
		return nil, err
	}
	casualties := 0
	converted := 0
	winner := attackerHeroID
	loser := defenderHeroID
	loserPlayer := defHero.PlayerID
	loserNode := defHero.CurrentNodeID
	if result.Outcome == "loss" {
		winner = defenderHeroID
		loser = attackerHeroID
		loserPlayer = attHero.PlayerID
		loserNode = attHero.CurrentNodeID
	}

	loserRows, err := q.ListHeroUnitsByHero(ctx, loser)
	if err != nil {
		return nil, err
	}
	for _, row := range loserRows {
		casualties += int(row.Qty)
	}
	if err := q.ClearHeroUnits(ctx, loser); err != nil {
		return nil, err
	}
	home, err := q.GetCastleByPlayer(ctx, loserPlayer)
	if err != nil {
		return nil, err
	}
	if err := q.UpdateHeroNode(ctx, gen.UpdateHeroNodeParams{
		ID: loser, CurrentNodeID: home.NodeID,
	}); err != nil {
		return nil, err
	}
	_ = loserNode

	if len(loserRows) > 0 {
		for _, row := range loserRows {
			gain := int(row.Qty) * economy.EnemyConversionPercent / 100
			if gain <= 0 {
				continue
			}
			converted += gain
			if err := q.AddHeroUnits(ctx, gen.AddHeroUnitsParams{
				HeroID: winner, UnitID: row.UnitID, Qty: int32(gain),
			}); err != nil {
				return nil, err
			}
		}
	}

	_, _ = q.InsertCombatLog(ctx, gen.InsertCombatLogParams{
		HeroID:         winner,
		CreepID:        pgtype.Int8{Valid: false},
		EnemyHeroID:    pgtype.Int8{Int64: loser, Valid: true},
		Outcome:        "win",
		GoldReward:     0,
		ConvertedUnits: int32(converted),
		Log:            logJSON,
	})
	if winner == attackerHeroID {
		_ = q.IncrementPlayerKill(ctx, gen.IncrementPlayerKillParams{
			KillerPlayerID: attHero.PlayerID,
			VictimPlayerID: defHero.PlayerID,
		})
	} else {
		_ = q.IncrementPlayerKill(ctx, gen.IncrementPlayerKillParams{
			KillerPlayerID: defHero.PlayerID,
			VictimPlayerID: attHero.PlayerID,
		})
	}

	winningPlayerID := attHero.PlayerID
	if winner == defenderHeroID {
		winningPlayerID = defHero.PlayerID
	}
	castle, err := q.GetCastleByPlayer(ctx, winningPlayerID)
	if err != nil {
		return nil, err
	}
	enemy := loser
	winnerHero := winner
	resolvedOutcome := "win"
	if winner != attackerHeroID {
		resolvedOutcome = "loss"
	}
	return &ApplyResult{
		Payload: proto.CombatResolvedPayload{
			HeroID:         winnerHero,
			CreepID:        0,
			EnemyHeroID:    &enemy,
			Outcome:        resolvedOutcome,
			GoldReward:     0,
			Casualties:     casualties,
			ConvertedUnits: converted,
			Log:            toProtoLog(result.Log),
		},
		Respawn:    true,
		PlayerID:   winningPlayerID,
		CastleID:   castle.ID,
		GoldPerMin: castle.GoldPerMin,
	}, nil
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
			UnitID:  row.UnitID,
			Qty:     int(row.Qty),
			Attack:  int(row.Attack),
			Defense: int(row.Defense),
			HP:      int(row.Hp),
		})
	}

	if creep.GraceUntil.Valid && time.Now().UTC().Before(creep.GraceUntil.Time) {
		return nil, nil
	}
	heroStats, err := HeroCombatStats(ctx, q, h)
	if err != nil {
		return nil, err
	}
	defAtk := int(creep.Attack)
	defDef := int(creep.Defense)
	defHP := int(creep.Hp)
	if defAtk == 0 {
		defAtk = int(unit.Attack)
	}
	if defDef == 0 {
		defDef = int(unit.Defense)
	}
	if defHP == 0 {
		defHP = int(unit.Hp)
	}
	// Counter triangle: attacker stacks get +25% vs the creep's unit ID;
	// the creep gets +25% if its unit counters the hero's primary stack.
	attPrimary := PrimaryUnitID(stack)
	attacker := StackFromHeroVs(heroStats, stack, creep.UnitID)
	defender := StackFromCreepVs(creep.UnitID, defAtk, defDef, defHP, int(creep.Qty), attPrimary)
	result := Resolve(attacker, defender)

	logJSON, err := json.Marshal(result.Log)
	if err != nil {
		return nil, err
	}

	var goldReward int32
	casualties := 0
	convertedUnits := 0

	switch result.Outcome {
	case "win":
		goldReward = creep.GoldReward
		goldDelta, err := numericDelta(float64(goldReward))
		if err != nil {
			return nil, err
		}
		zero, err := numericDelta(0)
		if err != nil {
			return nil, err
		}
		if err := q.IncrementPlayerResources(ctx, gen.IncrementPlayerResourcesParams{
			ID:         h.PlayerID,
			GoldDelta:  goldDelta,
			MetalDelta: zero,
			GemsDelta:  zero,
			CoalDelta:  zero,
			WoodDelta:  zero,
			StoneDelta: zero,
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
		convertedUnits = max(0, int(creep.Qty)*economy.EnemyConversionPercent/100)
		if convertedUnits > 0 {
			if err := q.AddHeroUnits(ctx, gen.AddHeroUnitsParams{
				HeroID: heroID,
				UnitID: creep.UnitID,
				Qty:    int32(convertedUnits),
			}); err != nil {
				return nil, err
			}
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
		HeroID:         heroID,
		CreepID:        pgtype.Int8{Int64: creep.ID, Valid: true},
		EnemyHeroID:    pgtype.Int8{Valid: false},
		Outcome:        result.Outcome,
		GoldReward:     goldReward,
		ConvertedUnits: int32(convertedUnits),
		Log:            logJSON,
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
			ConvertedUnits: convertedUnits,
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
