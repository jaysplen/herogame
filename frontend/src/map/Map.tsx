import { useCallback, useMemo, useState } from "react";
import { Layer, Stage } from "react-konva";
import { send } from "../net/ws";
import { MsgMoveRequest } from "../proto/types";
import type { MapNodeDTO } from "../proto/messages";
import { useGameStore, useServerNow } from "../state/store";
import { Edge } from "./Edge";
import { HeroToken } from "./HeroToken";
import { Node } from "./Node";
import {
  lerp,
  moveProgress,
  neighborIds,
  nodeMap,
  uniqueLogicalEdges,
} from "./geometry";

const STAGE_WIDTH = 820;
const STAGE_HEIGHT = 620;

export function MapView() {
  const map = useGameStore((s) => s.map);
  const hero = useGameStore((s) => s.hero);
  const bootstrap = useGameStore((s) => s.bootstrap);
  const inFlight = useGameStore((s) => s.inFlight);
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
      return {
        x: lerp(from.x, to.x, t),
        y: lerp(from.y, to.y, t),
      };
    }
    const node = nodesById.get(hero.currentNodeId);
    return node ? { x: node.x, y: node.y } : null;
  }, [map, hero, inFlight, serverNow, nodesById]);

  const handleNodeSelect = useCallback(
    (targetId: number) => {
      if (!hero || !bootstrap) return;
      if (targetId === hero.currentNodeId) return;

      if (inFlight) {
        showToast("Hero is moving");
        return;
      }

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
    [hero, bootstrap, inFlight, neighbors, showToast],
  );

  if (!map || !hero) {
    return <p className="muted">Map loads after hello.ack…</p>;
  }

  return (
    <div className="map-wrap">
      {toast ? <div className="toast">{toast}</div> : null}
      <Stage width={STAGE_WIDTH} height={STAGE_HEIGHT}>
        <Layer>
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
              onSelect={handleNodeSelect}
            />
          ))}
          {heroPosition ? (
            <HeroToken x={heroPosition.x} y={heroPosition.y} />
          ) : null}
        </Layer>
      </Stage>
    </div>
  );
}
