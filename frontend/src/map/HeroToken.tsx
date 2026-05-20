import { Circle, Group, Star } from "react-konva";

interface HeroTokenProps {
  x: number;
  y: number;
}

/** Hero marker drawn above nodes during travel interpolation. */
export function HeroToken({ x, y }: HeroTokenProps) {
  return (
    <Group x={x} y={y} listening={false}>
      <Circle radius={14} fill="#e9c46a" stroke="#1a1b1e" strokeWidth={2} />
      <Star
        numPoints={5}
        innerRadius={4}
        outerRadius={8}
        fill="#1a1b1e"
        y={-1}
      />
    </Group>
  );
}
