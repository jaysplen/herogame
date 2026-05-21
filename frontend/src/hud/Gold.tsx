import { useDisplayGoldInt } from "./useDisplayGold";
import { useGameStore } from "../state/store";

export function Gold() {
  const display = useDisplayGoldInt();
  const goldPerMin = useGameStore((s) => s.goldAnchor?.goldPerMin);
  const resources = useGameStore((s) => s.player.resources);

  if (display == null) {
    return (
      <section className="hud-panel">
        <h2>Gold</h2>
        <p className="muted">—</p>
      </section>
    );
  }

  return (
    <section className="hud-panel">
      <h2>Gold</h2>
      <p className="hud-gold-value">{display}</p>
      {goldPerMin != null ? (
        <p className="hud-meta">+{goldPerMin} / min (estimate between ticks)</p>
      ) : null}
      {resources ? (
        <ul className="resource-mini">
          <li>M {Math.floor(resources.metal)}</li>
          <li>Gm {Math.floor(resources.gems)}</li>
          <li>C {Math.floor(resources.coal)}</li>
          <li>W {Math.floor(resources.wood)}</li>
          <li>S {Math.floor(resources.stone)}</li>
        </ul>
      ) : null}
    </section>
  );
}
