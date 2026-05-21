import { Group, Line } from "react-konva";
import type { MapNodeDTO } from "../proto/messages";

interface EdgeProps {
  from: MapNodeDTO;
  to: MapNodeDTO;
}

/**
 * Stylized stone-and-rune road between nodes.
 * Renders an outer dark shadow under a warmer gold-tinted track.
 */
export function Edge({ from, to }: EdgeProps) {
  const points = [from.x, from.y, to.x, to.y];

  return (
    <Group listening={false}>
      {/* Shadow underlay */}
      <Line
        points={points}
        stroke="#0a0805"
        strokeWidth={7}
        lineCap="round"
        opacity={0.7}
      />
      {/* Stone base */}
      <Line
        points={points}
        stroke="#3a2c1a"
        strokeWidth={5}
        lineCap="round"
      />
      {/* Warm worn-path inner stripe */}
      <Line
        points={points}
        stroke="#7a5e36"
        strokeWidth={2.5}
        lineCap="round"
        opacity={0.85}
      />
      {/* Subtle ley-line highlight */}
      <Line
        points={points}
        stroke="#d4a548"
        strokeWidth={0.8}
        lineCap="round"
        opacity={0.35}
        dash={[6, 10]}
      />
    </Group>
  );
}
