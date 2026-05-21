import { Circle, Group, Ring, Text } from "react-konva";
import type { MapNodeDTO } from "../proto/messages";

const RADIUS = 24;

interface NodeStyle {
  fill: string;
  fillShade: string;
  stroke: string;
  glow: string;
  symbol?: string;
  symbolColor?: string;
}

const KIND_STYLES: Record<string, NodeStyle> = {
  castle: {
    fill: "#4a6d9a",
    fillShade: "#1e3052",
    stroke: "#a8c6e8",
    glow: "#7aaee0",
    symbol: "♜",
    symbolColor: "#f0e6c8",
  },
  wild: {
    fill: "#3d7a52",
    fillShade: "#1a3a26",
    stroke: "#a4e2b8",
    glow: "#7acc8e",
    symbol: "❦",
    symbolColor: "#d6f0d6",
  },
  creep: {
    fill: "#9b2d2d",
    fillShade: "#3e0c0c",
    stroke: "#f0a878",
    glow: "#e87060",
    symbol: "☠",
    symbolColor: "#ffe2bc",
  },
};

function styleFor(kind: string): NodeStyle {
  return (
    KIND_STYLES[kind] ?? {
      fill: "#4a3a26",
      fillShade: "#1a1308",
      stroke: "#a08454",
      glow: "#c8a060",
    }
  );
}

export interface NodeProps {
  node: MapNodeDTO;
  reachable: boolean;
  onSelect: (nodeId: number) => void;
  tooltip?: string;
}

export function Node({ node, reachable, onSelect, tooltip }: NodeProps) {
  const s = styleFor(node.kind);
  const haloColor = reachable ? "#f3d27a" : s.glow;

  return (
    <Group
      x={node.x}
      y={node.y}
      onClick={() => onSelect(node.id)}
      onTap={() => onSelect(node.id)}
    >
      {/* Outer glow halo */}
      <Circle
        radius={RADIUS + (reachable ? 6 : 3)}
        fill={haloColor}
        opacity={reachable ? 0.22 : 0.08}
        listening={false}
      />
      {/* Stone base (dark shade for depth) */}
      <Circle radius={RADIUS + 2} fill={s.fillShade} listening={false} />
      {/* Main jewel face with radial gradient */}
      <Circle
        radius={RADIUS}
        fillRadialGradientStartPoint={{ x: -RADIUS / 3, y: -RADIUS / 3 }}
        fillRadialGradientStartRadius={0}
        fillRadialGradientEndPoint={{ x: 0, y: 0 }}
        fillRadialGradientEndRadius={RADIUS}
        fillRadialGradientColorStops={[0, s.fill, 1, s.fillShade]}
        stroke={reachable ? "#f3d27a" : s.stroke}
        strokeWidth={reachable ? 3 : 2}
        shadowBlur={reachable ? 14 : 6}
        shadowColor={haloColor}
        shadowOpacity={reachable ? 0.9 : 0.5}
      />
      {/* Inner gold ornament ring */}
      <Ring
        innerRadius={RADIUS - 5}
        outerRadius={RADIUS - 3}
        fill="#d4a548"
        opacity={0.55}
        listening={false}
      />
      {/* Specular highlight */}
      <Circle
        x={-RADIUS / 3}
        y={-RADIUS / 3}
        radius={RADIUS / 3}
        fill="#ffffff"
        opacity={0.16}
        listening={false}
      />
      {/* Kind symbol */}
      {s.symbol ? (
        <Text
          text={s.symbol}
          fontSize={20}
          fontStyle="bold"
          fill={s.symbolColor ?? "#f3d27a"}
          width={RADIUS * 2}
          offsetX={RADIUS}
          offsetY={11}
          align="center"
          shadowColor="#000"
          shadowBlur={3}
          shadowOpacity={0.6}
          listening={false}
        />
      ) : null}
      {/* Tooltip text */}
      {tooltip ? (
        <Text
          text={tooltip}
          fontSize={11}
          fontFamily="Spectral, Georgia, serif"
          fill="#f3d27a"
          width={160}
          offsetX={80}
          y={-RADIUS - 22}
          align="center"
          shadowColor="#000"
          shadowBlur={4}
          shadowOpacity={0.85}
          listening={false}
        />
      ) : null}
      {/* Node name label */}
      <Text
        text={node.name}
        fontSize={12}
        fontFamily="Cinzel, serif"
        fontStyle="bold"
        fill="#ede2c4"
        width={140}
        offsetX={70}
        y={RADIUS + 8}
        align="center"
        shadowColor="#000"
        shadowBlur={3}
        shadowOpacity={0.85}
        listening={false}
      />
    </Group>
  );
}

export const NODE_RADIUS = RADIUS;
