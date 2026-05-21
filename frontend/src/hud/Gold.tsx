import { useDisplayGoldInt } from "./useDisplayGold";
import { useGameStore } from "../state/store";

function CoinIcon() {
  return (
    <svg
      className="coin-icon"
      viewBox="0 0 32 32"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <defs>
        <radialGradient id="coinGrad" cx="35%" cy="30%" r="75%">
          <stop offset="0%" stopColor="#fff1b8" />
          <stop offset="45%" stopColor="#f3d27a" />
          <stop offset="85%" stopColor="#b8862c" />
          <stop offset="100%" stopColor="#6b4a14" />
        </radialGradient>
      </defs>
      <circle cx="16" cy="16" r="13" fill="url(#coinGrad)" stroke="#3a2810" strokeWidth="1.5" />
      <circle cx="16" cy="16" r="9.5" fill="none" stroke="#8a5d18" strokeWidth="0.8" />
      <text
        x="16"
        y="21"
        textAnchor="middle"
        fontFamily="Cinzel, serif"
        fontWeight="700"
        fontSize="13"
        fill="#5a3f10"
      >
        ⚜
      </text>
    </svg>
  );
}

export function Gold() {
  const display = useDisplayGoldInt();
  const goldPerMin = useGameStore((s) => s.goldAnchor?.goldPerMin);
  const resources = useGameStore((s) => s.player.resources);

  if (display == null) {
    return (
      <section className="hud-panel">
        <h2>Treasury</h2>
        <p className="muted">—</p>
      </section>
    );
  }

  return (
    <section className="hud-panel">
      <h2>Treasury</h2>
      <div className="hud-gold-row">
        <CoinIcon />
        <p className="hud-gold-value">{display}</p>
      </div>
      {goldPerMin != null ? (
        <p className="hud-meta">+{goldPerMin} per minute</p>
      ) : null}
      {resources ? (
        <ul className="resource-mini" aria-label="Resources">
          <li data-kind="wood" data-icon="❦" title="Wood">
            {Math.floor(resources.wood)}
          </li>
          <li data-kind="stone" data-icon="◈" title="Stone">
            {Math.floor(resources.stone)}
          </li>
          <li data-kind="metal" data-icon="⚒" title="Metal">
            {Math.floor(resources.metal)}
          </li>
          <li data-kind="coal" data-icon="◆" title="Coal">
            {Math.floor(resources.coal)}
          </li>
          <li data-kind="gems" data-icon="✦" title="Gems">
            {Math.floor(resources.gems)}
          </li>
        </ul>
      ) : null}
    </section>
  );
}
