import { useDisplayGoldInt } from "./useDisplayGold";
import { useGameStore } from "../state/store";

export function Gold() {
  const display = useDisplayGoldInt();
  const goldPerMin = useGameStore((s) => s.goldAnchor?.goldPerMin);

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
    </section>
  );
}
