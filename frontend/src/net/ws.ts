import { encodeEnvelope, parseEnvelope } from "../proto/envelope";
import { MsgHello } from "../proto/types";
import type { HelloPayload } from "../proto/messages";
import { useGameStore } from "../state/store";

const DEFAULT_URL = "ws://localhost:8080/ws";
const DEFAULT_PLAYER_ID = 1;

let socket: WebSocket | null = null;
let seq = 0;

function nextSeq(): number {
  seq += 1;
  return seq;
}

export function wsUrl(): string {
  const fromEnv = import.meta.env.VITE_WS_URL as string | undefined;
  return fromEnv?.trim() || DEFAULT_URL;
}

export function connect(playerId = DEFAULT_PLAYER_ID): void {
  disconnect();
  const store = useGameStore.getState();
  store.reset();
  store.setConnection({ status: "connecting", error: null });

  const ws = new WebSocket(wsUrl());
  socket = ws;

  ws.onopen = () => {
    const hello: HelloPayload = { playerId };
    ws.send(encodeEnvelope(MsgHello, hello, nextSeq()));
  };

  ws.onmessage = (ev) => {
    try {
      const envelope = parseEnvelope(
        typeof ev.data === "string" ? ev.data : String(ev.data),
      );
      useGameStore.getState().applyEnvelope(envelope);
    } catch (err) {
      useGameStore.getState().setConnection({
        status: "error",
        error: err instanceof Error ? err.message : "parse error",
      });
    }
  };

  ws.onerror = () => {
    useGameStore.getState().setConnection({
      status: "error",
      error: "websocket error",
    });
  };

  ws.onclose = () => {
    socket = null;
    const { connection } = useGameStore.getState();
    if (connection.status !== "error") {
      useGameStore.getState().setConnection({ status: "disconnected", error: null });
    }
  };
}

export function disconnect(): void {
  if (socket) {
    socket.close();
    socket = null;
  }
}

export function send<T>(type: string, payload: T): void {
  if (!socket || socket.readyState !== WebSocket.OPEN) {
    throw new Error("websocket not connected");
  }
  socket.send(encodeEnvelope(type, payload, nextSeq()));
}

export function isConnected(): boolean {
  return socket?.readyState === WebSocket.OPEN;
}
