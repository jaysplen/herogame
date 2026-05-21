import { Circle, Group, Text } from "react-konva";
import type { MapNodeDTO } from "../proto/messages";

const RADIUS = 22;

const KIND_COLORS: Record<string, { fill: string; stroke: string }> = {
  castle: { fill: "#3d5a80", stroke: "#98c1d9" },
  wild: { fill: "#2d6a4f", stroke: "#95d5b2" },
  creep: { fill: "#9b2226", stroke: "#f4a261" },
};

function colorsFor(kind: string) {
  return KIND_COLORS[kind] ?? { fill: "#444", stroke: "#888" };
}

export interface NodeProps {
  node: MapNodeDTO;
  reachable: boolean;
  onSelect: (nodeId: number) => void;
  tooltip?: string;
}

export function Node({ node, reachable, onSelect, tooltip }: NodeProps) {
  const colors = colorsFor(node.kind);

  return (
    <Group
      x={node.x}
      y={node.y}
      onClick={() => onSelect(node.id)}
      onTap={() => onSelect(node.id)}
    >
      <Circle
        radius={RADIUS}
        fill={colors.fill}
        stroke={reachable ? "#f9dc5c" : colors.stroke}
        strokeWidth={reachable ? 3 : 2}
        shadowBlur={reachable ? 8 : 0}
        shadowColor="#f9dc5c"
      />
      {tooltip ? (
        <Text
          text={tooltip}
          fontSize={10}
          fill="#f4a261"
          width={150}
          offsetX={75}
          y={-RADIUS - 18}
          align="center"
        />
      ) : null}
      <Text
        text={node.name}
        fontSize={11}
        fill="#e8e6e3"
        width={120}
        offsetX={60}
        y={RADIUS + 6}
        align="center"
      />
    </Group>
  );
}

export const NODE_RADIUS = RADIUS;
