import { useMemo } from "react";
import { useGameStore, useServerNow } from "../state/store";

export function HeroPanel() {
  const hero = useGameStore((s) => s.hero);
  const serverNow = useServerNow(500);

  const respawnSecondsLeft = useMemo(() => {
    const until = hero?.respawnUntil;
    if (until == null || serverNow >= until) return 0;
    return Math.ceil((until - serverNow) / 1000);
  }, [hero?.respawnUntil, serverNow]);

  const spawnMessage = useMemo(() => {
    if (!hero?.respawnUntil) return null;
    if (serverNow < hero.respawnUntil) {
      return `Hero is dead (${Math.ceil((hero.respawnUntil - serverNow) / 1000)}s)`;
    }
    return "Go!";
  }, [hero?.respawnUntil, serverNow]);

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
          ☠ Resurrecting {respawnSecondsLeft}s
        </span>
      ) : null}
      {spawnMessage ? <p className="spawn-msg">{spawnMessage}</p> : null}
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
          <dt>Locus</dt>
          <dd>#{hero.currentNodeId}</dd>
        </div>
      </dl>
    </section>
  );
}
