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
		})
	}
	return out
}

func mapShopUnits(catalog []gen.Unit) []proto.HeroUnitStackDTO {
	out := make([]proto.HeroUnitStackDTO, 0, len(catalog))
	for _, u := range catalog {
		out = append(out, proto.HeroUnitStackDTO{
			UnitID:   u.ID,
			Code:     u.Code,
			Name:     u.Name,
			Qty:      0,
			CostGold: u.CostGold,
		})
	}
	return out
}
