import { useMemo } from "react";
import { useGameStore, useServerNow } from "../state/store";

/** Extrapolate gold from last authoritative tick (architecture.md §7.3). */
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
