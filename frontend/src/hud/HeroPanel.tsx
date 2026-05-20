import { useEffect, useMemo } from "react";
import { useGameStore, useServerNow } from "../state/store";

export function HeroPanel() {
  const hero = useGameStore((s) => s.hero);
  const respawnUntilMs = useGameStore((s) => s.respawnUntilMs);
  const serverNow = useServerNow(500);

  useEffect(() => {
    if (respawnUntilMs != null && serverNow >= respawnUntilMs) {
      useGameStore.setState({ respawnUntilMs: null });
    }
  }, [respawnUntilMs, serverNow]);

  const respawnSecondsLeft = useMemo(() => {
    if (respawnUntilMs == null || serverNow >= respawnUntilMs) return 0;
    return Math.ceil((respawnUntilMs - serverNow) / 1000);
  }, [respawnUntilMs, serverNow]);

  if (!hero) {
    return (
      <section className="hud-panel">
        <h2>Hero</h2>
        <p className="muted">—</p>
      </section>
    );
  }

  return (
    <section className="hud-panel">
      <h2>Hero</h2>
      {respawnSecondsLeft > 0 ? (
        <span className="badge badge-respawn">
          Respawning {respawnSecondsLeft}s
        </span>
      ) : null}
      <dl className="hero-stats">
        <div>
          <dt>Speed</dt>
          <dd>{hero.speedEffective.toFixed(1)}</dd>
        </div>
        <div>
          <dt>Upkeep</dt>
          <dd>{hero.upkeepGoldPerHour.toFixed(1)} / hr</dd>
        </div>
        <div>
          <dt>Node</dt>
          <dd>{hero.currentNodeId}</dd>
        </div>
      </dl>
    </section>
  );
}
