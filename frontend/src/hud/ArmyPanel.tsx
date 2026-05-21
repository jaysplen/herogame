import { send } from "../net/ws";
import { MsgCastleBuild, MsgUnitBuy } from "../proto/types";
import type { HeroUnitStackDTO } from "../proto/messages";
import { useGameStore } from "../state/store";
import { useDisplayGoldInt } from "./useDisplayGold";
import { counterHint, unitGlyph } from "./constants";

function UnitRow({
  stack,
  displayGold,
  onBuy,
  unitNameById,
}: {
  stack: HeroUnitStackDTO;
  displayGold: number | null;
  onBuy: (unitTypeId: number, qty: number) => void;
  unitNameById: (id: number) => string;
}) {
  const canBuy1 =
    displayGold != null && displayGold >= stack.costGold;
  const canBuy10 =
    displayGold != null && displayGold >= stack.costGold * 10;
  const hint = counterHint(stack.unitId, unitNameById);

  return (
    <li className="army-row">
      <span className="army-name">
        <span className="army-glyph" aria-hidden="true">
          {unitGlyph(stack.unitId)}
        </span>{" "}
        {stack.name} × {stack.qty}
        {hint ? <span className="army-counter"> · {hint}</span> : null}
      </span>
      <span className="army-actions">
        <button
          type="button"
          data-testid={`recruit-${stack.unitId}-plus1`}
          disabled={!canBuy1}
          onClick={() => onBuy(stack.unitId, 1)}
          title={`${stack.costGold} gold`}
        >
          +1
        </button>
        <button
          type="button"
          data-testid={`recruit-${stack.unitId}-plus10`}
          disabled={!canBuy10}
          onClick={() => onBuy(stack.unitId, 10)}
          title={`${stack.costGold * 10} gold`}
        >
          +10
        </button>
      </span>
    </li>
  );
}

export function ArmyPanel() {
  const hero = useGameStore((s) => s.hero);
  const castleId = useGameStore((s) => s.castle.castleId);
  const bootstrap = useGameStore((s) => s.bootstrap);
  const displayGold = useDisplayGoldInt();
  const resources = useGameStore((s) => s.player.resources);

  const stacks = hero?.units ?? [];
  const shopUnits = bootstrap?.shopUnits ?? [];

  // Best-effort name lookup for counter hints: prefer the hero's stack rows
  // (always populated), fall back to the castle shop catalog, then to a stub.
  const unitNameById = (id: number): string => {
    const fromArmy = stacks.find((s) => s.unitId === id);
    if (fromArmy) return fromArmy.name;
    const fromShop = shopUnits.find((s) => s.unitId === id);
    if (fromShop) return fromShop.name;
    return `Unit ${id}`;
  };

  const buy = (unitTypeId: number, qty: number) => {
    if (!castleId || !bootstrap) return;
    try {
      send(MsgUnitBuy, { castleId, unitTypeId, qty });
    } catch {
      /* disconnected */
    }
  };

  const build = (buildingCode: string) => {
    if (!castleId) return;
    try {
      send(MsgCastleBuild, { castleId, buildingCode });
    } catch {
      /* disconnected */
    }
  };

  if (castleId == null) {
    return (
      <section className="hud-panel">
        <h2>Army</h2>
        <p className="muted">—</p>
      </section>
    );
  }

  const stackIds = new Set(stacks.map((s) => s.unitId));
  const recruitOnly = shopUnits.filter((u) => !stackIds.has(u.unitId));

  return (
    <section className="hud-panel">
      <h2>Garrison</h2>
      <ul className="army-list">
        {stacks.map((stack) => (
          <UnitRow
            key={stack.unitId}
            stack={stack}
            displayGold={displayGold}
            onBuy={buy}
            unitNameById={unitNameById}
          />
        ))}
        {stacks.length === 0 ? (
          <li className="muted army-empty">No troops in retinue</li>
        ) : null}
      </ul>
      {recruitOnly.length > 0 ? (
        <>
          <h3 className="hud-subhead">⚔ Recruit at Keep</h3>
          <ul className="army-list">
            {recruitOnly.map((unit) => (
              <UnitRow
                key={`shop-${unit.unitId}`}
                stack={unit}
                displayGold={displayGold}
                onBuy={buy}
                unitNameById={unitNameById}
              />
            ))}
          </ul>
        </>
      ) : null}
      <h3 className="hud-subhead">✦ Construct</h3>
      <div className="build-menu">
        <button type="button" onClick={() => build("defense")}>
          Bulwark (+1)
        </button>
        <button type="button" onClick={() => build("barracks")}>
          Barracks Tier
        </button>
        <button type="button" onClick={() => build("academy")}>
          Academy Tier
        </button>
      </div>
      {resources ? (
        <p className="hud-meta">
          Wood {Math.floor(resources.wood)} · Stone {Math.floor(resources.stone)}
        </p>
      ) : null}
    </section>
  );
}
