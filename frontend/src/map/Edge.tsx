import { Group, Path } from "react-konva";
import type { MapNodeDTO } from "../proto/messages";
import { edgeCurve, pathD } from "./geometry";

interface EdgeProps {
  from: MapNodeDTO;
  to: MapNodeDTO;
}

/**
 * Hand-drawn winding road between two nodes.
 *
 * Renders an outer shadow, a stone underlay, a warm worn track, and
 * a dashed ley-line accent — all sharing the same cubic Bezier curve
 * produced by `edgeCurve` (geometry.ts). The hero token interpolates
 * along the same curve so it actually follows the road instead of
 * cutting the chord.
 */
export function Edge({ from, to }: EdgeProps) {
  const d = pathD(edgeCurve(from, to));

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
