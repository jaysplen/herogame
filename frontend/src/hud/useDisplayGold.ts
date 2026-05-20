import { useMemo } from "react";
import { useGameStore, useServerNow } from "../state/store";

/**
 * Display-only gold estimate between server castle.tick events (architecture.md §7.3).
 * Authoritative balance changes only on the server tick engine / buy / combat.
 */
export function extrapolateGold(
  anchorGold: number,
  anchorAtMs: number,
  goldPerMin: number,
  nowMs: number,
): number {
  const elapsedSec = (nowMs - anchorAtMs) / 1000;
  const goldPerSec = goldPerMin / 60;
  return anchorGold + goldPerSec * elapsedSec;
}

export function useDisplayGold(): number | null {
  const anchor = useGameStore((s) => s.goldAnchor);
  const serverNowMs = useServerNow(1000);

  return useMemo(() => {
    if (!anchor) return null;
    return extrapolateGold(
      anchor.gold,
      anchor.atMs,
      anchor.goldPerMin,
      serverNowMs,
    );
  }, [anchor, serverNowMs]);
}

/** Integer gold for HUD display. */
export function useDisplayGoldInt(): number | null {
  const v = useDisplayGold();
  return v == null ? null : Math.floor(v);
}
