import { send } from "../net/ws";
import { MsgUnitBuy } from "../proto/types";
import { useGameStore } from "../state/store";
import { PIKEMAN_COST_GOLD, PIKEMAN_NAME, PIKEMAN_UNIT_ID } from "./constants";
import { useDisplayGoldInt } from "./useDisplayGold";

export function ArmyPanel() {
  const hero = useGameStore((s) => s.hero);
  const castleId = useGameStore((s) => s.castle.castleId);
  const bootstrap = useGameStore((s) => s.bootstrap);
  const displayGold = useDisplayGoldInt();

  const qty = hero?.armySize ?? 0;
  const canBuy1 =
    castleId != null &&
    displayGold != null &&
    displayGold >= PIKEMAN_COST_GOLD;
  const canBuy10 =
    castleId != null &&
    displayGold != null &&
    displayGold >= PIKEMAN_COST_GOLD * 10;

  const buy = (amount: number) => {
    if (!castleId || !bootstrap) return;
    try {
      send(MsgUnitBuy, {
        castleId,
        unitTypeId: PIKEMAN_UNIT_ID,
        qty: amount,
      });
    } catch {
      /* disconnected */
    }
  };

  return (
    <section className="hud-panel">
      <h2>Army</h2>
      <ul className="army-list">
        <li className="army-row">
          <span className="army-name">
            {PIKEMAN_NAME} × {qty}
          </span>
          <span className="army-actions">
            <button
              type="button"
              disabled={!canBuy1}
              onClick={() => buy(1)}
              title={`${PIKEMAN_COST_GOLD} gold`}
            >
              +1
            </button>
            <button
              type="button"
              disabled={!canBuy10}
              onClick={() => buy(10)}
              title={`${PIKEMAN_COST_GOLD * 10} gold`}
            >
              +10
            </button>
          </span>
        </li>
      </ul>
      <p className="hud-meta">{PIKEMAN_COST_GOLD} gold each</p>
    </section>
  );
}
