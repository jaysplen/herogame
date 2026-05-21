import {
  Circle,
  Group,
  Line,
  Path,
  Rect,
  RegularPolygon,
  Text,
} from "react-konva";
import type { MapNodeDTO } from "../proto/messages";

/* ----------------------------------------------------------------------------
 * Node icon variants — chosen by node.kind and keywords in node.name so the
 * watercolor map reads like a HoMM cartouche: forests have trees, caves have
 * arches, castles have banners, etc.
 * -------------------------------------------------------------------------- */

type Variant =
  | "castle-player"
  | "castle-enemy"
  | "forest"
  | "caves"
  | "quarry"
  | "coal"
  | "ruin"
  | "ridge"
  | "lumber"
  | "field"
  | "pass"
  | "gate"
  | "creep-wolf"
  | "creep-bandit"
  | "marsh"
  | "crossing"
  | "wild-generic";

function variantFor(node: MapNodeDTO): Variant {
  const n = node.name.toLowerCase();
  if (node.kind === "castle") {
    // Player keep (low id) uses cool palette; enemy keep uses warm
    return node.id === 1 || /ironkeep/.test(n) ? "castle-player" : "castle-enemy";
  }
  if (node.kind === "creep") {
    if (/wolf/.test(n)) return "creep-wolf";
    return "creep-bandit";
  }
  if (/forest/.test(n)) return "forest";
  if (/cave/.test(n)) return "caves";
  if (/quarry/.test(n)) return "quarry";
  if (/coal|pit/.test(n)) return "coal";
  if (/ruin|watch/.test(n)) return "ruin";
  if (/ridge|stone/.test(n)) return "ridge";
  if (/lumber|wood/.test(n)) return "lumber";
  if (/field|golden|wheat/.test(n)) return "field";
  if (/pass/.test(n)) return "pass";
  if (/gate/.test(n)) return "gate";
  if (/marsh|mercury|swamp/.test(n)) return "marsh";
  if (/cross/.test(n)) return "crossing";
  return "wild-generic";
}

/* ----------------------------------------------------------------------------
 * Icon primitives
 * -------------------------------------------------------------------------- */

interface IconProps {
  reachable: boolean;
}

function CastleIcon({ reachable, faction }: IconProps & { faction: "player" | "enemy" }) {
  const wall = faction === "player" ? "#7a8aa8" : "#7a6e60";
  const wallShade = faction === "player" ? "#3a4258" : "#3a3024";
  const banner = faction === "player" ? "#3a7ac4" : "#c83232";
  const stroke = reachable ? "#f3d27a" : "#3a2410";
  return (
    <Group>
      {/* Base walls */}
      <Rect x={-18} y={-2} width={36} height={18} fill={wall} stroke={stroke} strokeWidth={1.2} />
      {/* Crenellations */}
      <Rect x={-18} y={-6} width={6} height={4} fill={wall} stroke={stroke} strokeWidth={0.8} />
      <Rect x={-8} y={-6} width={6} height={4} fill={wall} stroke={stroke} strokeWidth={0.8} />
      <Rect x={2} y={-6} width={6} height={4} fill={wall} stroke={stroke} strokeWidth={0.8} />
      <Rect x={12} y={-6} width={6} height={4} fill={wall} stroke={stroke} strokeWidth={0.8} />
      {/* Wall shading */}
      <Rect x={-18} y={10} width={36} height={6} fill={wallShade} opacity={0.55} />
      {/* Gate */}
      <Path data="M -4 16 L -4 6 Q 0 1 4 6 L 4 16 Z" fill="#2a1a08" />
      {/* Tower */}
      <Rect x={-4} y={-22} width={8} height={16} fill={wall} stroke={stroke} strokeWidth={1.2} />
      <Rect x={-5} y={-24} width={2} height={4} fill={wall} stroke={stroke} strokeWidth={0.6} />
      <Rect x={-1} y={-24} width={2} height={4} fill={wall} stroke={stroke} strokeWidth={0.6} />
      <Rect x={3} y={-24} width={2} height={4} fill={wall} stroke={stroke} strokeWidth={0.6} />
      {/* Banner pole + flag */}
      <Line points={[0, -26, 0, -36]} stroke="#3a2410" strokeWidth={1} />
      <Path data="M 0 -36 L 10 -33 L 0 -30 Z" fill={banner} stroke="#3a2410" strokeWidth={0.6} />
    </Group>
  );
}

function ForestIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Three tree canopies */}
      <Circle x={-9} y={2} radius={9} fill="#6b8a4a" stroke={stroke} strokeWidth={0.8} />
      <Circle x={9} y={2} radius={9} fill="#6b8a4a" stroke={stroke} strokeWidth={0.8} />
      <Circle x={0} y={-6} radius={10} fill="#7d9a58" stroke={stroke} strokeWidth={0.8} />
      {/* Trunks */}
      <Rect x={-10.5} y={9} width={3} height={5} fill="#3a2410" />
      <Rect x={7.5} y={9} width={3} height={5} fill="#3a2410" />
      <Rect x={-1.5} y={2} width={3} height={6} fill="#3a2410" />
    </Group>
  );
}

function CavesIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Hill */}
      <Path
        data="M -18 12 Q -8 -14 0 -14 Q 8 -14 18 12 Z"
        fill="#86796a"
        stroke={stroke}
        strokeWidth={1}
      />
      {/* Cave mouth */}
      <Path
        data="M -7 12 Q -7 -2 0 -2 Q 7 -2 7 12 Z"
        fill="#1a1208"
        stroke={stroke}
        strokeWidth={0.8}
      />
      {/* Sparkle gem */}
      <Path data="M 0 4 L 2 6 L 0 8 L -2 6 Z" fill="#e87fc8" />
    </Group>
  );
}

function QuarryIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Stone blocks */}
      <Rect x={-12} y={-2} width={10} height={8} fill="#b8b0a4" stroke={stroke} strokeWidth={0.8} />
      <Rect x={-2} y={-6} width={10} height={12} fill="#cdc4b4" stroke={stroke} strokeWidth={0.8} />
      <Rect x={8} y={-2} width={8} height={8} fill="#a89e90" stroke={stroke} strokeWidth={0.8} />
      <Rect x={-7} y={6} width={14} height={4} fill="#8e8576" stroke={stroke} strokeWidth={0.8} />
      {/* Pickaxe */}
      <Line points={[-10, -10, 6, -2]} stroke="#3a2410" strokeWidth={1.4} />
      <Path data="M -14 -12 Q -10 -16 -4 -10" stroke="#3a2410" strokeWidth={1.6} fill="none" />
    </Group>
  );
}

function CoalIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Pit rim */}
      <Path
        data="M -16 6 Q -8 12 0 12 Q 8 12 16 6 Q 16 0 0 -2 Q -16 0 -16 6 Z"
        fill="#5a4838"
        stroke={stroke}
        strokeWidth={1}
      />
      {/* Dark hole */}
      <Path
        data="M -10 4 Q -4 8 0 8 Q 4 8 10 4 Q 10 0 0 -1 Q -10 0 -10 4 Z"
        fill="#0e0805"
      />
      {/* Coal chunks */}
      <Circle x={-4} y={6} radius={1.6} fill="#2a2018" />
      <Circle x={3} y={7} radius={1.4} fill="#2a2018" />
      <Circle x={-1} y={9} radius={1.2} fill="#2a2018" />
    </Group>
  );
}

function RuinIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Broken tower */}
      <Path
        data="M -8 14 L -8 -10 L -4 -14 L -4 -6 L -1 -10 L -1 0 L 4 0 L 4 -12 L 8 -10 L 8 14 Z"
        fill="#8a7e6a"
        stroke={stroke}
        strokeWidth={1}
      />
      {/* Window slit */}
      <Rect x={-1} y={2} width={2} height={4} fill="#1a1208" />
      {/* Rubble */}
      <Circle x={-12} y={14} radius={2} fill="#6e6354" stroke={stroke} strokeWidth={0.6} />
      <Circle x={12} y={14} radius={2.5} fill="#6e6354" stroke={stroke} strokeWidth={0.6} />
    </Group>
  );
}

function RidgeIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Three peaks */}
      <Path data="M -16 12 L -6 -8 L 2 0 L 8 -12 L 16 12 Z" fill="#9aa0a8" stroke={stroke} strokeWidth={1} />
      {/* Snow */}
      <Path data="M -6 -8 L -2 -2 L -10 2 Z" fill="#fff" opacity={0.85} />
      <Path data="M 8 -12 L 12 -6 L 4 -4 Z" fill="#fff" opacity={0.8} />
      {/* Shadow */}
      <Path data="M -16 12 L -6 -8 L -2 -2 L -10 2 L -4 8 Z" fill="#6e7480" opacity={0.45} />
    </Group>
  );
}

function LumberIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Stacked logs */}
      <Rect x={-14} y={4} width={28} height={6} fill="#8a5d18" stroke={stroke} strokeWidth={0.8} />
      <Circle x={-14} y={7} radius={3} fill="#d2a060" stroke={stroke} strokeWidth={0.6} />
      <Circle x={14} y={7} radius={3} fill="#d2a060" stroke={stroke} strokeWidth={0.6} />
      <Rect x={-10} y={-2} width={20} height={6} fill="#a06d2a" stroke={stroke} strokeWidth={0.8} />
      <Circle x={-10} y={1} radius={3} fill="#e2b070" stroke={stroke} strokeWidth={0.6} />
      <Circle x={10} y={1} radius={3} fill="#e2b070" stroke={stroke} strokeWidth={0.6} />
      {/* Axe */}
      <Line points={[0, -14, -3, -4]} stroke="#3a2410" strokeWidth={1.4} />
      <Path data="M 0 -16 L 6 -14 L 4 -10 L -2 -12 Z" fill="#9aa0a8" stroke={stroke} strokeWidth={0.6} />
    </Group>
  );
}

function FieldIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#5a3f10";
  return (
    <Group>
      {/* Field disc */}
      <Path
        data="M -16 6 Q 0 -6 16 6 Q 0 14 -16 6 Z"
        fill="#e8c860"
        stroke={stroke}
        strokeWidth={1}
      />
      {/* Wheat */}
      {[-10, -4, 2, 8].map((cx) => (
        <Group key={cx} x={cx} y={-2}>
          <Line points={[0, 6, 0, -8]} stroke="#a8842c" strokeWidth={0.8} />
          <Path data="M 0 -8 q 2 -1 0 -3 q -2 -1 0 -3 q 2 -1 0 -3" stroke="#5a3f10" strokeWidth={0.6} fill="none" />
        </Group>
      ))}
    </Group>
  );
}

function PassIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Two peaks framing path */}
      <Path data="M -18 12 L -6 -10 L 0 6 Z" fill="#8a8a96" stroke={stroke} strokeWidth={0.8} />
      <Path data="M 0 6 L 6 -10 L 18 12 Z" fill="#8a8a96" stroke={stroke} strokeWidth={0.8} />
      {/* Path */}
      <Path data="M -10 14 Q 0 8 10 14" stroke="#c8a878" strokeWidth={3} fill="none" />
      <Path data="M -10 14 Q 0 8 10 14" stroke="#5a3f10" strokeWidth={0.6} fill="none" />
    </Group>
  );
}

function GateIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Posts */}
      <Rect x={-14} y={-12} width={5} height={26} fill="#7a6e60" stroke={stroke} strokeWidth={1} />
      <Rect x={9} y={-12} width={5} height={26} fill="#7a6e60" stroke={stroke} strokeWidth={1} />
      {/* Arch */}
      <Path data="M -14 -12 Q 0 -22 14 -12" stroke="#7a6e60" strokeWidth={3} fill="none" />
      <Path data="M -14 -12 Q 0 -22 14 -12" stroke={stroke} strokeWidth={0.6} fill="none" />
      {/* Hanging banner */}
      <Path data="M -4 -14 L 4 -14 L 4 -4 L 0 -7 L -4 -4 Z" fill="#8e1d1d" stroke={stroke} strokeWidth={0.6} />
    </Group>
  );
}

function WolfIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Head */}
      <Path data="M -10 4 Q -8 -10 0 -10 Q 8 -10 10 4 Q 8 10 0 10 Q -8 10 -10 4 Z"
            fill="#5e554c" stroke={stroke} strokeWidth={1} />
      {/* Ears */}
      <Path data="M -8 -6 L -10 -14 L -4 -10 Z" fill="#5e554c" stroke={stroke} strokeWidth={0.6} />
      <Path data="M 8 -6 L 10 -14 L 4 -10 Z" fill="#5e554c" stroke={stroke} strokeWidth={0.6} />
      {/* Eyes */}
      <Circle x={-3} y={-2} radius={1.4} fill="#e8c860" />
      <Circle x={3} y={-2} radius={1.4} fill="#e8c860" />
      {/* Snout */}
      <Path data="M 0 4 L -3 8 L 3 8 Z" fill="#2a1a08" />
      <Line points={[-1.5, 8, -1.5, 11]} stroke={stroke} strokeWidth={0.6} />
      <Line points={[1.5, 8, 1.5, 11]} stroke={stroke} strokeWidth={0.6} />
    </Group>
  );
}

function BanditIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Crossed swords */}
      <Line points={[-12, -12, 12, 12]} stroke="#3a2410" strokeWidth={2.5} />
      <Line points={[12, -12, -12, 12]} stroke="#3a2410" strokeWidth={2.5} />
      <Line points={[-10, -10, 10, 10]} stroke="#c0c4cc" strokeWidth={1.4} />
      <Line points={[10, -10, -10, 10]} stroke="#c0c4cc" strokeWidth={1.4} />
      {/* Pommels */}
      <Circle x={-12} y={-12} radius={2.4} fill="#d4a548" stroke={stroke} strokeWidth={0.5} />
      <Circle x={12} y={-12} radius={2.4} fill="#d4a548" stroke={stroke} strokeWidth={0.5} />
      {/* Skull center */}
      <Circle radius={5.5} fill="#ede2c4" stroke="#2a1a08" strokeWidth={0.8} />
      <Circle x={-1.6} y={-0.5} radius={1.1} fill="#2a1a08" />
      <Circle x={1.6} y={-0.5} radius={1.1} fill="#2a1a08" />
      <Path data="M -1.8 2.2 L 0 1.2 L 1.8 2.2 L 0 3.2 Z" fill="#2a1a08" />
    </Group>
  );
}

function MarshIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Water puddle */}
      <Path
        data="M -16 6 Q -8 12 0 8 Q 8 12 16 6 Q 12 -2 0 0 Q -12 -2 -16 6 Z"
        fill="#7a8aa8"
        stroke={stroke}
        strokeWidth={0.8}
      />
      {/* Ripples */}
      <Path data="M -10 5 Q -5 3 0 5" stroke="#cdd6e0" strokeWidth={0.6} fill="none" />
      <Path data="M 2 7 Q 7 5 12 7" stroke="#cdd6e0" strokeWidth={0.6} fill="none" />
      {/* Reeds */}
      <Line points={[-6, 0, -8, -10]} stroke="#5a6a3a" strokeWidth={0.8} />
      <Line points={[-4, 0, -3, -10]} stroke="#5a6a3a" strokeWidth={0.8} />
      <Line points={[6, 0, 8, -10]} stroke="#5a6a3a" strokeWidth={0.8} />
      <Line points={[8, 0, 5, -12]} stroke="#5a6a3a" strokeWidth={0.8} />
    </Group>
  );
}

function CrossingIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      {/* Signpost */}
      <Rect x={-1.5} y={-14} width={3} height={24} fill="#3a2410" stroke={stroke} strokeWidth={0.6} />
      {/* Arrow signs */}
      <Path data="M -14 -10 L -2 -10 L -2 -4 L -14 -4 L -16 -7 Z" fill="#a08560" stroke={stroke} strokeWidth={0.6} />
      <Path data="M 14 -2 L 2 -2 L 2 4 L 14 4 L 16 1 Z" fill="#a08560" stroke={stroke} strokeWidth={0.6} />
      {/* Roof */}
      <RegularPolygon sides={4} rotation={45} radius={3} x={0} y={-14} fill="#d4a548" stroke={stroke} strokeWidth={0.4} />
    </Group>
  );
}

function WildIcon({ reachable }: IconProps) {
  const stroke = reachable ? "#f3d27a" : "#2a1a08";
  return (
    <Group>
      <Circle radius={11} fill="#7a6e54" stroke={stroke} strokeWidth={1} />
      <Path data="M -6 4 L -2 -2 L 2 2 L 6 -4" stroke="#3a2410" strokeWidth={1.2} fill="none" />
    </Group>
  );
}

function RenderVariant({ v, reachable }: { v: Variant; reachable: boolean }) {
  switch (v) {
    case "castle-player": return <CastleIcon reachable={reachable} faction="player" />;
    case "castle-enemy":  return <CastleIcon reachable={reachable} faction="enemy" />;
    case "forest":        return <ForestIcon reachable={reachable} />;
    case "caves":         return <CavesIcon reachable={reachable} />;
    case "quarry":        return <QuarryIcon reachable={reachable} />;
    case "coal":          return <CoalIcon reachable={reachable} />;
    case "ruin":          return <RuinIcon reachable={reachable} />;
    case "ridge":         return <RidgeIcon reachable={reachable} />;
    case "lumber":        return <LumberIcon reachable={reachable} />;
    case "field":         return <FieldIcon reachable={reachable} />;
    case "pass":          return <PassIcon reachable={reachable} />;
    case "gate":          return <GateIcon reachable={reachable} />;
    case "creep-wolf":    return <WolfIcon reachable={reachable} />;
    case "creep-bandit":  return <BanditIcon reachable={reachable} />;
    case "marsh":         return <MarshIcon reachable={reachable} />;
    case "crossing":      return <CrossingIcon reachable={reachable} />;
    case "wild-generic":
    default:              return <WildIcon reachable={reachable} />;
  }
}

/* ----------------------------------------------------------------------------
 * Node component
 * -------------------------------------------------------------------------- */

export interface NodeProps {
  node: MapNodeDTO;
  reachable: boolean;
  onSelect: (nodeId: number) => void;
  tooltip?: string;
}

const HIT_RADIUS = 28;
const LABEL_OFFSET = 30;

export function Node({ node, reachable, onSelect, tooltip }: NodeProps) {
  const v = variantFor(node);

  return (
    <Group
      x={node.x}
      y={node.y}
      onClick={() => onSelect(node.id)}
      onTap={() => onSelect(node.id)}
    >
      {/* Reachable glow halo */}
      {reachable ? (
        <Circle
          radius={HIT_RADIUS + 4}
          fill="#f3d27a"
          opacity={0.22}
          listening={false}
        />
      ) : null}

      {/* Soft parchment disc behind the icon for legibility on the watercolor */}
      <Circle
        radius={HIT_RADIUS - 4}
        fill="#f4e6c2"
        opacity={0.78}
        stroke={reachable ? "#d4a548" : "#5a3f10"}
        strokeWidth={reachable ? 1.6 : 0.9}
        shadowColor="#000"
        shadowBlur={6}
        shadowOpacity={0.35}
        shadowOffsetY={1.5}
        listening={false}
      />

      {/* Big transparent hit circle so the user can click anywhere on/near the icon */}
      <Circle radius={HIT_RADIUS} opacity={0} />

      {/* Variant illustration */}
      <Group listening={false}>
        <RenderVariant v={v} reachable={reachable} />
      </Group>

      {/* Tooltip text */}
      {tooltip ? (
        <Text
          text={tooltip}
          fontSize={11}
          fontFamily="Spectral, Georgia, serif"
          fill="#f3d27a"
          width={170}
          offsetX={85}
          y={-HIT_RADIUS - 22}
          align="center"
          shadowColor="#000"
          shadowBlur={4}
          shadowOpacity={0.85}
          listening={false}
        />
      ) : null}

      {/* Name label on a small parchment ribbon */}
      <Group y={LABEL_OFFSET} listening={false}>
        <Rect
          x={-58}
          y={-2}
          width={116}
          height={16}
          fill="#f4e6c2"
          opacity={0.88}
          stroke="#5a3f10"
          strokeWidth={0.6}
          cornerRadius={2}
        />
        <Text
          text={node.name}
          fontSize={11}
          fontFamily="Cinzel, serif"
          fontStyle="bold"
          fill="#3a2410"
          width={120}
          offsetX={60}
          y={1}
          align="center"
          listening={false}
        />
      </Group>
    </Group>
  );
}

export const NODE_RADIUS = HIT_RADIUS;
