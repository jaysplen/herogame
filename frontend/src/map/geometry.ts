import type { MapEdgeDTO, MapNodeDTO } from "../proto/messages";

/** One undirected edge (7 logical edges from 14 DB rows). */
export interface LogicalEdge {
  fromNodeId: number;
  toNodeId: number;
}

export function uniqueLogicalEdges(edges: MapEdgeDTO[]): LogicalEdge[] {
  const seen = new Set<string>();
  const out: LogicalEdge[] = [];
  for (const e of edges) {
    const a = Math.min(e.fromNodeId, e.toNodeId);
    const b = Math.max(e.fromNodeId, e.toNodeId);
    const key = `${a}-${b}`;
    if (seen.has(key)) continue;
    seen.add(key);
    out.push({ fromNodeId: a, toNodeId: b });
  }
  return out;
}

export function neighborIds(nodeId: number, edges: MapEdgeDTO[]): number[] {
  const ids = new Set<number>();
  for (const e of edges) {
    if (e.fromNodeId === nodeId) ids.add(e.toNodeId);
    if (e.toNodeId === nodeId) ids.add(e.fromNodeId);
  }
  return [...ids];
}

export function nodeMap(nodes: MapNodeDTO[]): Map<number, MapNodeDTO> {
  return new Map(nodes.map((n) => [n.id, n]));
}

export function lerp(a: number, b: number, t: number): number {
  return a + (b - a) * t;
}

/**
 * Visual progress along a move using server-provided departAt/arriveAt and skew-corrected clock.
 * Does not compute travel duration — that is server-only (architecture.md §10).
 */
export function moveProgress(
  serverNowMs: number,
  departAt: number,
  arriveAt: number,
): number {
  const span = arriveAt - departAt;
  if (span <= 0) return 1;
  return Math.min(1, Math.max(0, (serverNowMs - departAt) / span));
}
