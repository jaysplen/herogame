import type { CombatResolvedPayload } from "../proto/messages";

interface CombatModalProps {
  combat: CombatResolvedPayload | null;
  open: boolean;
  onClose: () => void;
}

export function CombatModal({ combat, open, onClose }: CombatModalProps) {
  if (!open || !combat) return null;

  return (
    <div
      className="modal-backdrop"
      role="dialog"
      aria-modal="true"
      data-testid="combat-modal"
    >
      <div className="modal">
        <header className="modal-header">
          <h2>
            Combat —{" "}
            <span className={combat.outcome === "win" ? "win" : "loss"}>
              {combat.outcome}
            </span>
          </h2>
          <button type="button" className="modal-close" onClick={onClose}>
            ×
          </button>
        </header>
        <p className="modal-summary">
          Casualties: {combat.casualties}
          {combat.goldReward > 0 ? ` · Gold +${combat.goldReward}` : null}
          {combat.convertedUnits > 0
            ? ` · Converted +${combat.convertedUnits}`
            : null}
        </p>
        <div className="modal-log">
          <table>
            <thead>
              <tr>
                <th>Rnd</th>
                <th>Side</th>
                <th>Dmg</th>
                <th>HP after</th>
              </tr>
            </thead>
            <tbody>
              {combat.log.map((entry, i) => (
                <tr key={i}>
                  <td>{entry.round}</td>
                  <td>{entry.side}</td>
                  <td>{entry.damage}</td>
                  <td>
                    {entry.defenderHpAfter ?? entry.attackerHpAfter ?? "—"}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
          {combat.log.length === 0 ? (
            <p className="muted">No combat rounds (instant resolution).</p>
          ) : null}
        </div>
      </div>
    </div>
  );
}
