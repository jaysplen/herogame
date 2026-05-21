import { Group, Path } from "react-konva";
import type { MapNodeDTO } from "../proto/messages";

interface EdgeProps {
  from: MapNodeDTO;
  to: MapNodeDTO;
}

/**
 * Deterministic 0..1 hash so road wobble is stable per edge.
 * Same wobble on every render — no jitter.
 */
function hash01(a: number, b: number): number {
  const lo = Math.min(a, b);
  const hi = Math.max(a, b);
  let x = ((lo * 73856093) ^ (hi * 19349663)) >>> 0;
  x = (x ^ (x >>> 13)) >>> 0;
  x = (x * 0x5bd1e995) >>> 0;
  return (x & 0xffff) / 0xffff;
}

/**
 * Cubic-bezier path between two nodes with a deterministic perpendicular
 * bend — produces a hand-drawn winding road instead of a ruler-straight line.
 */
function curvedPathData(from: MapNodeDTO, to: MapNodeDTO): string {
  const dx = to.x - from.x;
  const dy = to.y - from.y;
  const len = Math.hypot(dx, dy) || 1;
  // Unit perpendicular
  const px = -dy / len;
  const py = dx / len;

  const h = hash01(from.id, to.id);
  const sign = h < 0.5 ? -1 : 1;
  // Curve magnitude: 8..28px depending on edge length & hash
  const mag = sign * (8 + h * 20 + len * 0.06);

  // Two control points pulled off the chord at 1/3 and 2/3 with the same offset
  const c1x = from.x + dx * 0.33 + px * mag;
  const c1y = from.y + dy * 0.33 + py * mag;
  const c2x = from.x + dx * 0.66 + px * mag * 0.85;
  const c2y = from.y + dy * 0.66 + py * mag * 0.85;

  return `M ${from.x} ${from.y} C ${c1x} ${c1y}, ${c2x} ${c2y}, ${to.x} ${to.y}`;
}

/**
 * Hand-drawn winding road between two nodes.
 * Renders an outer shadow, a stone underlay, a warm worn track,
 * and a dashed ley-line accent — all sharing the same curve.
 */
export function Edge({ from, to }: EdgeProps) {
  const d = curvedPathData(from, to);

  return (
    <Group listening={false}>
      {/* Shadow underlay */}
      <Path data={d} stroke="#1a0e05" strokeWidth={9} lineCap="round" opacity={0.55} />
      {/* Earth base */}
      <Path data={d} stroke="#3a2c1a" strokeWidth={6} lineCap="round" />
      {/* Worn path */}
      <Path data={d} stroke="#a08560" strokeWidth={3} lineCap="round" opacity={0.95} />
      {/* Inner highlight */}
      <Path data={d} stroke="#e0c890" strokeWidth={1.2} lineCap="round" opacity={0.6} />
      {/* Subtle dashed gold ley accent */}
      <Path
        data={d}
        stroke="#d4a548"
        strokeWidth={0.6}
        lineCap="round"
        opacity={0.45}
        dash={[5, 9]}
      />
    </Group>
  );
}
