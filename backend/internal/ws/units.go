package ws

import (
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store/gen"
)

func mapHeroUnitStacks(rows []gen.ListHeroUnitsByHeroRow) []proto.HeroUnitStackDTO {
	out := make([]proto.HeroUnitStackDTO, 0, len(rows))
	for _, r := range rows {
		out = append(out, proto.HeroUnitStackDTO{
			UnitID:   r.UnitID,
			Code:     r.Code,
			Name:     r.Name,
			Qty:      int(r.Qty),
			CostGold: r.CostGold,
			CostMetal: 0,
			CostGems:  0,
			CostCoal:  0,
			CostWood:  0,
			CostStone: 0,
			Tier:      1,
			Faction:   "neutral",
		})
	}
	return out
}

func mapShopUnits(catalog []gen.ListUnitsByFactionTierRow) []proto.HeroUnitStackDTO {
	out := make([]proto.HeroUnitStackDTO, 0, len(catalog))
	for _, u := range catalog {
		out = append(out, proto.HeroUnitStackDTO{
			UnitID:   u.ID,
			Code:     u.Code,
			Name:     u.Name,
			Qty:      0,
			CostGold: u.CostGold,
			CostMetal: u.CostMetal,
			CostGems:  u.CostGems,
			CostCoal:  u.CostCoal,
			CostWood:  u.CostWood,
			CostStone: u.CostStone,
			Tier:      u.Tier,
			Faction:   u.Faction,
		})
	}
	return out
}
