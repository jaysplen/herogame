/** Client → server message types (mirrors backend/internal/proto). */
export const MsgHello = "hello" as const;
export const MsgMoveRequest = "move.request" as const;
export const MsgUnitBuy = "unit.buy" as const;

/** Server → client message types. */
export const MsgHelloAck = "hello.ack" as const;
export const MsgMoveUpdate = "move.update" as const;
export const MsgMoveArrived = "move.arrived" as const;
export const MsgCombatResolved = "combat.resolved" as const;
export const MsgCastleTick = "castle.tick" as const;
export const MsgHeroState = "hero.state" as const;
export const MsgError = "error" as const;

export type ClientMessageType = typeof MsgHello | typeof MsgMoveRequest | typeof MsgUnitBuy;
export type ServerMessageType =
  | typeof MsgHelloAck
  | typeof MsgMoveUpdate
  | typeof MsgMoveArrived
  | typeof MsgCombatResolved
  | typeof MsgCastleTick
  | typeof MsgHeroState
  | typeof MsgError;
