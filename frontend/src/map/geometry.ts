import type { MapEdgeDTO, MapNodeDTO } from "../proto/messages";

/** One undirected edge (24 logical edges from 48 DB rows). */
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

/* ----------------------------------------------------------------------------
 * Shared road geometry — cubic Bezier curve per edge.
 *
 * Both Edge.tsx (drawing the road) and Map.tsx (hero / creep movement
 * interpolation) use these helpers so the hero token follows the road
 * instead of cutting the chord.
 *
 * Each curve is built from a perpendicular offset off the chord. The
 * BENDS table assigns a hand-tuned magnitude and side to every logical
 * edge so the 24 roads fan out cleanly around dense junctions
 * (e.g. nodes 5, 9, 10, 15, 16) and don't collide. Unknown edges fall
 * back to a deterministic hash-based bend.
 * -------------------------------------------------------------------------- */

export interface BezCurve {
  p0x: number; p0y: number;
  c1x: number; c1y: number;
  c2x: number; c2y: number;
  p3x: number; p3y: number;
}

interface Bend {
  /** Perpendicular offset in pixels off the chord. */
  mag: number;
  /** +1 bends toward CCW perpendicular of (from→to); -1 bends toward CW. */
  sign: -1 | 1;
}

/**
 * Hand-tuned bends keyed by "min-max" node id. Designed against the
 * node layout in migration 00005_reposition_map_nodes.sql (1200x820
 * stage) so adjacent roads splay apart instead of stacking.
 *
 * The sign is interpreted as if the road is traveled in the canonical
 * direction (min id → max id). edgeCurve() reuses the same curve when
 * the hero travels the road backwards, so the visual stays put.
 */
const BENDS: Record<string, Bend> = {
  "1-2":   { mag: 26,  sign:  1 },  // Ironkeep → Moss Crossing, arcs north along grassland
  "2-3":   { mag: 38,  sign: -1 },  // Moss Crossing → North Forest, swings west around the woods
  "2-4":   { mag: 26,  sign:  1 },  // Moss Crossing → South Quarry, arcs east
  "2-5":   { mag: 30,  sign:  1 },  // Moss Crossing → Bandit Camp, dips south
  "3-7":   { mag: 22,  sign:  1 },  // North Forest → Gem Caves, dips south through the trees
  "4-8":   { mag: 24,  sign: -1 },  // South Quarry → Coal Pit, arcs north along the shore
  "5-7":   { mag: 48,  sign: -1 },  // Bandit Camp → Gem Caves, sweeps west around the wood
  "5-8":   { mag: 38,  sign: -1 },  // Bandit Camp → Coal Pit, sweeps west
  "5-9":   { mag: 32,  sign:  1 },  // Bandit Camp → Ruined Watch, climbs NW
  "5-10":  { mag: 32,  sign:  1 },  // Bandit Camp → Old Lumber, drops SE
  "9-11":  { mag: 26,  sign: -1 },  // Ruined Watch → Stone Ridge, climbs over the ridge
  "10-12": { mag: 30,  sign:  1 },  // Old Lumber → Golden Field, dips south through fields
  "9-15":  { mag: 36,  sign: -1 },  // Ruined Watch → Wolf Den, hugs the west slope
  "10-15": { mag: 28,  sign:  1 },  // Old Lumber → Wolf Den, arcs east
  "15-16": { mag: 22,  sign:  1 },  // Wolf Den → Mercury Marsh, dips south
  "11-16": { mag: 48,  sign:  1 },  // Stone Ridge → Mercury Marsh, large eastern arc
  "12-16": { mag: 48,  sign: -1 },  // Golden Field → Mercury Marsh, large eastern arc
  "11-13": { mag: 30,  sign:  1 },  // Stone Ridge → East Pass, drops S then climbs
  "12-14": { mag: 30,  sign: -1 },  // Golden Field → South Gate, arcs north
  "6-13":  { mag: 28,  sign: -1 },  // Sunspire → East Pass, mountain pass arc
  "6-14":  { mag: 38,  sign:  1 },  // Sunspire → South Gate, sweeps east outside the realm
  "6-16":  { mag: 64,  sign:  1 },  // Sunspire → Mercury Marsh, long southern detour
  "9-16":  { mag: 64,  sign: -1 },  // Ruined Watch → Mercury Marsh, long northern detour
  "10-16": { mag: 64,  sign:  1 },  // Old Lumber → Mercury Marsh, long southern detour
};

function defaultBend(fromId: number, toId: number, len: number): Bend {
  const lo = Math.min(fromId, toId);
  const hi = Math.max(fromId, toId);
  let x = ((lo * 73856093) ^ (hi * 19349663)) >>> 0;
  x = (x ^ (x >>> 13)) >>> 0;
  x = (x * 0x5bd1e995) >>> 0;
  const h = (x & 0xffff) / 0xffff;
  return { mag: 14 + h * 18 + len * 0.04, sign: h < 0.5 ? -1 : 1 };
}

/**
 * Cubic Bezier curve for the road between two nodes. Always built in
 * the canonical direction (min id → max id) so both rendering and
 * traversal share the same geometry.
 */
export function edgeCurve(
  from: { id: number; x: number; y: number },
  to: { id: number; x: number; y: number },
): BezCurve {
  // Normalize: a is the lower id, b is the higher id.
  const swap = from.id > to.id;
  const a = swap ? to : from;
  const b = swap ? from : to;

  const dx = b.x - a.x;
  const dy = b.y - a.y;
  const len = Math.hypot(dx, dy) || 1;
  const px = -dy / len;
  const py = dx / len;

  const key = `${a.id}-${b.id}`;
  const bend = BENDS[key] ?? defaultBend(a.id, b.id, len);

  const offset = bend.sign * bend.mag;

  const c1x = a.x + dx * 0.30 + px * offset;
  const c1y = a.y + dy * 0.30 + py * offset;
  const c2x = a.x + dx * 0.70 + px * offset * 0.94;
  const c2y = a.y + dy * 0.70 + py * offset * 0.94;

  return {
    p0x: a.x, p0y: a.y,
    c1x, c1y, c2x, c2y,
    p3x: b.x, p3y: b.y,
  };
}

/**
 * Position along a cubic Bezier at parameter t ∈ [0, 1]. Standard
 * de Casteljau formula. Used by Map.tsx to place the hero token along
 * the same curve the Edge.tsx component renders.
 */
export function bezierPos(c: BezCurve, t: number): { x: number; y: number } {
  const u = 1 - t;
  const uu = u * u;
  const tt = t * t;
  const x =
    uu * u * c.p0x +
    3 * uu * t * c.c1x +
    3 * u * tt * c.c2x +
    tt * t * c.p3x;
  const y =
    uu * u * c.p0y +
    3 * uu * t * c.c1y +
    3 * u * tt * c.c2y +
    tt * t * c.p3y;
  return { x, y };
}

/** SVG / Konva "M ... C ..." path string for the same curve. */
export function pathD(c: BezCurve): string {
  return `M ${c.p0x} ${c.p0y} C ${c.c1x} ${c.c1y}, ${c.c2x} ${c.c2y}, ${c.p3x} ${c.p3y}`;
}

/**
 * Position of a unit at progress `t` traveling from→to. Internally
 * normalizes to the canonical edge direction so the path matches the
 * road drawn by Edge.tsx regardless of travel direction.
 */
export function travelPos(
  from: { id: number; x: number; y: number },
  to: { id: number; x: number; y: number },
  t: number,
): { x: number; y: number } {
  const curve = edgeCurve(from, to);
  const adj = from.id <= to.id ? t : 1 - t;
  return bezierPos(curve, adj);
}
