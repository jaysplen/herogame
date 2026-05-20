/** Wire envelope (architecture.md §7.1). */
export interface Envelope<T = unknown> {
  type: string;
  payload: T;
  seq: number;
  serverTime: number;
}

export function parseEnvelope(raw: string): Envelope<unknown> {
  const env = JSON.parse(raw) as Envelope<unknown>;
  if (typeof env.type !== "string" || typeof env.serverTime !== "number") {
    throw new Error("invalid envelope");
  }
  return env;
}

export function encodeEnvelope<T>(type: string, payload: T, seq: number): string {
  return JSON.stringify({
    type,
    payload,
    seq,
    serverTime: Date.now(),
  });
}

export function decodePayload<T>(env: Envelope<unknown>): T {
  if (typeof env.payload === "string") {
    return JSON.parse(env.payload) as T;
  }
  return env.payload as T;
}
