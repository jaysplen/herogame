import { Line } from "react-konva";
import type { MapNodeDTO } from "../proto/messages";

interface EdgeProps {
  from: MapNodeDTO;
  to: MapNodeDTO;
}

export function Edge({ from, to }: EdgeProps) {
  return (
    <Line
      points={[from.x, from.y, to.x, to.y]}
      stroke="#5c6370"
      strokeWidth={3}
      lineCap="round"
    />
  );
}
