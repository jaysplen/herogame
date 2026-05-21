import { send } from "../net/ws";
import { MsgCastleBuild, MsgUnitBuy } from "../proto/types";
import type { HeroUnitStackDTO } from "../proto/messages";
import { useGameStore } from "../state/store";
import { useDisplayGoldInt } from "./useDisplayGold";

function UnitRow({
  stack,
  displayGold,
  onBuy,
}: {
  stack: HeroUnitStackDTO;
  displayGold: number | null;
  onBuy: (unitTypeId: number, qty: number) => void;
}) {
  const canBuy1 =
    displayGold != null && displayGold >= stack.costGold;
  const canBuy10 =
    displayGold != null && displayGold >= stack.costGold * 10;

  return (
    <li className="army-row">
      <span className="army-name">
        {stack.name} × {stack.qty}
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
      <h2>Army</h2>
      <ul className="army-list">
        {stacks.map((stack) => (
          <UnitRow
            key={stack.unitId}
            stack={stack}
            displayGold={displayGold}
            onBuy={buy}
          />
        ))}
        {stacks.length === 0 ? (
          <li className="muted army-empty">No units in army</li>
        ) : null}
      </ul>
      {recruitOnly.length > 0 ? (
        <>
          <h3 className="hud-subhead">Recruit at castle</h3>
          <ul className="army-list">
            {recruitOnly.map((unit) => (
              <UnitRow
                key={`shop-${unit.unitId}`}
                stack={unit}
                displayGold={displayGold}
                onBuy={buy}
              />
            ))}
          </ul>
        </>
      ) : null}
      <h3 className="hud-subhead">Castle build</h3>
      <div className="build-menu">
        <button type="button" onClick={() => build("defense")}>
          Defense (+1)
        </button>
        <button type="button" onClick={() => build("barracks")}>
          Barracks tier
        </button>
        <button type="button" onClick={() => build("academy")}>
          Academy tier
        </button>
      </div>
      {resources ? (
        <p className="hud-meta">
          Build res: W {Math.floor(resources.wood)} · S {Math.floor(resources.stone)}
        </p>
      ) : null}
    </section>
  );
}
