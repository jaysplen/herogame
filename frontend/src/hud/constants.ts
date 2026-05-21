/**
 * Unit catalog mirror (game_rules.md §10, docs/future_features.md §1).
 *
 * The server is the source of truth (the `units` table), but the client keeps
 * a tiny mirror so it can render counter hints and pick per-unit icons before
 * the bootstrap snapshot arrives.
 */

export const PIKEMAN_UNIT_ID = 1;
export const ARCHER_UNIT_ID = 6;
export const CAVALRY_UNIT_ID = 7;

export const PIKEMAN_NAME = "Pikeman";
export const ARCHER_NAME = "Archer";
export const CAVALRY_NAME = "Cavalry";

export const PIKEMAN_COST_GOLD = 50;
export const ARCHER_COST_GOLD = 70;
export const CAVALRY_COST_GOLD = 110;

export const CASTLE_GOLD_PER_MIN_DEFAULT = 60;
export const RESPAWN_LOCKOUT_MS = 60_000;

/**
 * Rock-paper-scissors counter triangle. attackerUnitId -> defenderUnitId it
 * counters for +25% damage. Used only for HUD hints; the server applies the
 * actual multiplier in internal/combat/resolve.go.
 */
export const COUNTER_TARGET: Readonly<Record<number, number>> = Object.freeze({
  [PIKEMAN_UNIT_ID]: CAVALRY_UNIT_ID,
  [ARCHER_UNIT_ID]: PIKEMAN_UNIT_ID,
  [CAVALRY_UNIT_ID]: ARCHER_UNIT_ID,
});

/**
 * Glyph used in HUD lists; falls back to a generic shield for unknown units.
 */
export function unitGlyph(unitId: number): string {
  switch (unitId) {
    case PIKEMAN_UNIT_ID:
      return "\u{1F6E1}"; // shield
    case ARCHER_UNIT_ID:
      return "\u{1F3F9}"; // bow
    case CAVALRY_UNIT_ID:
      return "\u{1F40E}"; // horse
    default:
      return "⚔"; // crossed swords
  }
}

/**
 * Short hint shown next to the unit name describing what it counters.
 * Empty string for units that aren't in the triangle.
 */
export function counterHint(unitId: number, nameOf: (id: number) => string): string {
  const targetId = COUNTER_TARGET[unitId];
  if (!targetId) return "";
  return "vs " + nameOf(targetId);
}
