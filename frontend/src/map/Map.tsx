import { useCallback, useMemo, useRef, useState } from "react";
import { Group, Layer, Stage } from "react-konva";
import { send } from "../net/ws";
import { MsgMoveRequest } from "../proto/types";
import type { CreepStateDTO, MapNodeDTO } from "../proto/messages";
import { useGameStore, useServerNow } from "../state/store";
import { Edge } from "./Edge";
import { HeroToken } from "./HeroToken";
import { MapBackground } from "./MapBackground";
import { Node } from "./Node";
import {
  moveProgress,
  neighborIds,
  nodeMap,
  travelPos,
  uniqueLogicalEdges,
} from "./geometry";

const STAGE_WIDTH = 1200;
const STAGE_HEIGHT = 820;

function clampChance(n: number): number {
  return Math.max(5, Math.min(95, Math.round(n)));
}

function creepPosition(
  creep: CreepStateDTO,
  nodesById: Map<number, MapNodeDTO>,
  nowMs: number,
): { x: number; y: number } | null {
  if (creep.fromNodeId && creep.toNodeId && creep.departAt && creep.arriveAt) {
    const from = nodesById.get(creep.fromNodeId);
    const to = nodesById.get(creep.toNodeId);
    if (from && to) {
      const t = moveProgress(nowMs, creep.departAt, creep.arriveAt);
      // Follow the same road curve the Edge.tsx component draws.
      return travelPos(from, to, t);
    }
  }
  const node = nodesById.get(creep.nodeId);
  if (!node) return null;
  return { x: node.x, y: node.y };
}

export function MapView() {
  const deadAnnounced = useRef(false);
  const map = useGameStore((s) => s.map);
  const hero = useGameStore((s) => s.hero);
  const bootstrap = useGameStore((s) => s.bootstrap);
  const inFlight = useGameStore((s) => s.inFlight);
  const creeps = useGameStore((s) => s.creeps);
  const objective = useGameStore((s) => s.objective);
  const resourceNodes = useGameStore((s) => s.resourceNodes);
  const creepTooltip = useCallback(
    (c: (typeof creeps)[0]) => {
      const army = c.qty;
      const atk = c.attack * c.qty;
      const def = c.defense * c.qty;
      const hp = c.hp * c.qty;
      const heroScore =
        (hero?.armySize ?? 0) * (hero?.speedEffective ?? 1);
      const enemyScore = army * (c.attack + c.defense + c.hp / 4);
      const winChance = clampChance(
        (heroScore / Math.max(1, heroScore + enemyScore)) * 100,
      );
      return `${c.name}: ${army} | A:${atk} D:${def} HP:${hp} | Win~${winChance}%`;
    },
    [creeps, hero?.armySize, hero?.speedEffective],
  );

  const creepByNode = useMemo(() => {
    const m = new Map<number, string>();
    for (const c of creeps) {
      m.set(c.nodeId, creepTooltip(c));
    }
    return m;
  }, [creeps, creepTooltip]);

  const resourceByNode = useMemo(() => {
    const m = new Map<number, string>();
    for (const r of resourceNodes) {
      const owner = r.ownerPlayerId ? `P${r.ownerPlayerId}` : "neutral";
      m.set(r.nodeId, `${r.resourceType}+${r.perMin}/min (${owner})`);
    }
    return m;
  }, [resourceNodes]);
  const serverNow = useServerNow(33);
  const [toast, setToast] = useState<string | null>(null);

  const showToast = useCallback((message: string) => {
    setToast(message);
    window.setTimeout(() => setToast(null), 2500);
  }, []);

  const nodesById = useMemo(
    () => (map ? nodeMap(map.nodes) : new Map<number, MapNodeDTO>()),
    [map],
  );

  const logicalEdges = useMemo(
    () => (map ? uniqueLogicalEdges(map.edges) : []),
    [map],
  );

  const currentNodeId = hero?.currentNodeId ?? null;
  const neighbors = useMemo(() => {
    if (!map || currentNodeId == null) return new Set<number>();
    return new Set(neighborIds(currentNodeId, map.edges));
  }, [map, currentNodeId]);

  const heroPosition = useMemo(() => {
    if (!map || !hero) return null;
    if (inFlight) {
      const from = nodesById.get(inFlight.fromNodeId);
      const to = nodesById.get(inFlight.toNodeId);
      if (!from || !to) return null;
      const t = moveProgress(serverNow, inFlight.departAt, inFlight.arriveAt);
      // Hero follows the curved road, not the straight chord.
      return travelPos(from, to, t);
    }
    const node = nodesById.get(hero.currentNodeId);
    return node ? { x: node.x, y: node.y } : null;
  }, [map, hero, inFlight, serverNow, nodesById]);

  const handleNodeSelect = useCallback(
    (targetId: number) => {
      if (!hero || !bootstrap) return;
      if (targetId === hero.currentNodeId) return;
      if (hero.respawnUntil && serverNow < hero.respawnUntil) {
        const left = Math.ceil((hero.respawnUntil - serverNow) / 1000);
        const msg = `Hero is dead (${left}s)`;
        showToast(msg);
        window.alert(msg);
        return;
      }
      if (hero.respawnUntil && serverNow >= hero.respawnUntil) {
        if (!deadAnnounced.current) {
          showToast("Go!");
          window.alert("Go!");
          deadAnnounced.current = true;
        }
      }
      if (!hero.respawnUntil || serverNow < hero.respawnUntil) {
        deadAnnounced.current = false;
      }

      if (inFlight) {
        showToast("Hero is moving");
        return;
      }

      // UX hint only; server validates edges and timing on move.request.
      if (!neighbors.has(targetId)) {
        showToast("Not connected to that node");
        return;
      }

      try {
        send(MsgMoveRequest, {
          heroId: bootstrap.heroId,
          targetNodeId: targetId,
        });
      } catch {
        showToast("Not connected to server");
      }
    },
    [hero, bootstrap, inFlight, neighbors, showToast, serverNow, deadAnnounced],
  );

  if (!map || !hero) {
    return <p className="muted">Map loads after hello.ack…</p>;
  }

  return (
    <div className="map-wrap">
      {toast ? <div className="toast">{toast}</div> : null}
      {objective ? (
        <div className="objective-banner">
          Objective: eliminate enemy hero {objective.targetHeroKills}x · Progress{" "}
          {objective.enemyHeroKills}/{objective.targetHeroKills}
        </div>
      ) : null}
      <Stage width={STAGE_WIDTH} height={STAGE_HEIGHT}>
        <Layer>
          <MapBackground width={STAGE_WIDTH} height={STAGE_HEIGHT} />
          {logicalEdges.map((e) => {
            const from = nodesById.get(e.fromNodeId);
            const to = nodesById.get(e.toNodeId);
            if (!from || !to) return null;
            return (
              <Edge
                key={`${e.fromNodeId}-${e.toNodeId}`}
                from={from}
                to={to}
              />
            );
          })}
          {map.nodes.map((node) => (
            <Node
              key={node.id}
              node={node}
              reachable={!inFlight && neighbors.has(node.id)}
              tooltip={creepByNode.get(node.id) ?? resourceByNode.get(node.id)}
              onSelect={handleNodeSelect}
            />
          ))}
          {creeps.map((creep) => {
            const pos = creepPosition(creep, nodesById, serverNow);
            if (!pos) return null;
            return (
              <Group
                key={`creep-${creep.id}`}
                x={pos.x + 14}
                y={pos.y - 14}
                onMouseEnter={() => showToast(creepTooltip(creep))}
              >
                <HeroToken x={0} y={0} />
              </Group>
            );
          })}
          {heroPosition ? (
            <HeroToken x={heroPosition.x} y={heroPosition.y} />
          ) : null}
        </Layer>
      </Stage>
    </div>
  );
}
