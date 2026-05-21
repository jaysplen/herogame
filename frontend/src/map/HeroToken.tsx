import { Circle, Group, Star } from "react-konva";

interface HeroTokenProps {
  x: number;
  y: number;
}

/**
 * Hero marker: gold-leaf medallion with a glowing star sigil.
 * Drawn above nodes during travel interpolation.
 */
export function HeroToken({ x, y }: HeroTokenProps) {
  return (
    <Group x={x} y={y} listening={false}>
      {/* Outer aura */}
      <Circle radius={22} fill="#f3d27a" opacity={0.18} />
      {/* Dark base ring */}
      <Circle radius={16} fill="#1a1308" stroke="#0a0805" strokeWidth={1.5} />
      {/* Gold medallion */}
      <Circle
        radius={13}
        fillRadialGradientStartPoint={{ x: -4, y: -4 }}
        fillRadialGradientStartRadius={0}
        fillRadialGradientEndPoint={{ x: 0, y: 0 }}
        fillRadialGradientEndRadius={13}
        fillRadialGradientColorStops={[0, "#fff1b8", 0.55, "#f3d27a", 1, "#8a5d18"]}
        stroke="#3a2810"
        strokeWidth={1.3}
        shadowColor="#f3d27a"
        shadowBlur={12}
        shadowOpacity={0.8}
      />
      {/* Inner highlight */}
      <Circle x={-4} y={-4} radius={4} fill="#ffffff" opacity={0.45} />
      {/* Star sigil */}
      <Star
        numPoints={5}
        innerRadius={3.5}
        outerRadius={7.5}
        fill="#5a3f10"
        stroke="#3a2810"
        strokeWidth={0.6}
        y={0}
      />
    </Group>
  );
}
