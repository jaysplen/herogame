export const CodeHelloTimeout = "HELLO_TIMEOUT";
export const CodeHelloUnknownPlayer = "HELLO_UNKNOWN_PLAYER";
export const CodeHelloInvalidPayload = "HELLO_INVALID_PAYLOAD";
export const CodeMoveInvalidEdge = "MOVE_INVALID_EDGE";
export const CodeMoveHeroInFlight = "MOVE_HERO_IN_FLIGHT";
export const CodeMoveHeroRespawning = "MOVE_HERO_RESPAWNING";
export const CodeBuyInsufficientGold = "BUY_INSUFFICIENT_GOLD";
export const CodeUnknownMessage = "UNKNOWN_MESSAGE";
export const CodeInternal = "INTERNAL_ERROR";

export interface ErrorPayload {
  code: string;
  message: string;
  refSeq?: number;
}
