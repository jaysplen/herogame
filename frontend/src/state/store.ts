import { useEffect, useState } from "react";
import { create } from "zustand";
import type { Envelope } from "../proto/envelope";
import { decodePayload } from "../proto/envelope";
import type { ErrorPayload } from "../proto/errors";
import { CASTLE_GOLD_PER_MIN_DEFAULT } from "../hud/constants";
import {
  MsgCastleTick,
  MsgCombatResolved,
  MsgHelloAck,
  MsgHeroState,
  MsgMoveArrived,
  MsgMoveUpdate,
  MsgError,
} from "../proto/types";
import type {
  CastleTickPayload,
  CombatResolvedPayload,
  HelloAckPayload,
  HeroStatePayload,
  MapSnapshot,
  MoveUpdatePayload,
} from "../proto/messages";

export type ConnectionStatus =
  | "disconnected"
  | "connecting"
  | "connected"
  | "error";

export interface ConnectionSlice {
  status: ConnectionStatus;
  error: string | null;
}

export interface PlayerSlice {
  playerId: number | null;
  gold: number | null;
}

export interface CastleSlice {
  castleId: number | null;
  gold: number | null;
  goldPerMin: number | null;
}

/** Authoritative gold snapshot for client extrapolation (architecture.md §7.3). */
export interface GoldAnchor {
  gold: number;
  atMs: number;
  goldPerMin: number;
}

interface GameState {
  connection: ConnectionSlice;
  clockSkewMs: number;
  goldAnchor: GoldAnchor | null;
  player: PlayerSlice;
  hero: HeroStatePayload | null;
  castle: CastleSlice;
  map: MapSnapshot | null;
  inFlight: MoveUpdatePayload | null;
  bootstrap: HelloAckPayload | null;
  lastCombat: CombatResolvedPayload | null;
  combatModalOpen: boolean;
  lastError: ErrorPayload | null;
  lastEnvelope: Envelope<unknown> | null;

  setConnection: (patch: Partial<ConnectionSlice>) => void;
  applyEnvelope: (env: Envelope<unknown>) => void;
  dismissCombatModal: () => void;
  reset: () => void;
}

const initialCastle: CastleSlice = {
  castleId: null,
  gold: null,
  goldPerMin: null,
};

const initialPlayer: PlayerSlice = {
  playerId: null,
  gold: null,
};

function anchorFromGold(
  gold: number,
  goldPerMin: number,
  serverTimeMs: number,
): GoldAnchor {
  return { gold, atMs: serverTimeMs, goldPerMin };
}

export const useGameStore = create<GameState>((set, get) => ({
  connection: { status: "disconnected", error: null },
  clockSkewMs: 0,
  goldAnchor: null,
  player: { ...initialPlayer },
  hero: null,
  castle: { ...initialCastle },
  map: null,
  inFlight: null,
  bootstrap: null,
  lastCombat: null,
  combatModalOpen: false,
  lastError: null,
  lastEnvelope: null,

  setConnection: (patch) =>
    set((s) => ({
      connection: { ...s.connection, ...patch },
    })),

  dismissCombatModal: () => set({ combatModalOpen: false }),

  applyEnvelope: (env) => {
    const skew = env.serverTime - Date.now();
    set({ clockSkewMs: skew, lastEnvelope: env });

    switch (env.type) {
      case MsgHelloAck: {
        const ack = decodePayload<HelloAckPayload>(env);
        const gpm = CASTLE_GOLD_PER_MIN_DEFAULT;
        set({
          bootstrap: ack,
          player: { playerId: ack.playerId, gold: ack.gold },
          hero: ack.heroState,
          castle: {
            castleId: ack.castleId,
            gold: ack.gold,
            goldPerMin: gpm,
          },
          goldAnchor: anchorFromGold(ack.gold, gpm, env.serverTime),
          map: ack.mapSnapshot,
          inFlight: ack.inFlight ?? null,
          connection: { status: "connected", error: null },
        });
        break;
      }
      case MsgHeroState: {
        const hero = decodePayload<HeroStatePayload>(env);
        set({ hero });
        break;
      }
      case MsgCastleTick: {
        const tick = decodePayload<CastleTickPayload>(env);
        const { player } = get();
        set({
          player: { ...player, gold: tick.gold },
          castle: {
            castleId: tick.castleId,
            gold: tick.gold,
            goldPerMin: tick.goldPerMin,
          },
          goldAnchor: anchorFromGold(tick.gold, tick.goldPerMin, env.serverTime),
        });
        break;
      }
      case MsgMoveUpdate: {
        const move = decodePayload<MoveUpdatePayload>(env);
        set({ inFlight: move });
        break;
      }
      case MsgMoveArrived: {
        const arrived = decodePayload<{ heroId: number; nodeId: number }>(env);
        const { hero } = get();
        set({
          inFlight: null,
          hero: hero
            ? { ...hero, currentNodeId: arrived.nodeId }
            : hero,
        });
        break;
      }
      case MsgCombatResolved: {
        const combat = decodePayload<CombatResolvedPayload>(env);
        set({
          lastCombat: combat,
          combatModalOpen: true,
        });
        break;
      }
      case MsgError: {
        const err = decodePayload<ErrorPayload>(env);
        set({ lastError: err });
        break;
      }
      default:
        break;
    }
  },

  reset: () =>
    set({
      connection: { status: "disconnected", error: null },
      clockSkewMs: 0,
      goldAnchor: null,
      player: { ...initialPlayer },
      hero: null,
      castle: { ...initialCastle },
      map: null,
      inFlight: null,
      bootstrap: null,
      lastCombat: null,
      combatModalOpen: false,
      lastError: null,
      lastEnvelope: null,
    }),
}));

/** Authoritative server clock: Date.now() + skew (architecture.md §10). */
export function serverNow(): number {
  return Date.now() + useGameStore.getState().clockSkewMs;
}

/** Reactive hook for UI that needs live server time. */
export function useServerNow(intervalMs = 100): number {
  const skew = useGameStore((s) => s.clockSkewMs);
  const [now, setNow] = useState(() => Date.now() + skew);

  useEffect(() => {
    setNow(Date.now() + skew);
    const id = setInterval(() => {
      setNow(Date.now() + useGameStore.getState().clockSkewMs);
    }, intervalMs);
    return () => clearInterval(id);
  }, [skew, intervalMs]);

  return now;
}
