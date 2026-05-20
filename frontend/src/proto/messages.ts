export interface HelloPayload {
  playerId: number;
}

export interface MoveRequestPayload {
  heroId: number;
  targetNodeId: number;
}

export interface UnitBuyPayload {
  castleId: number;
  unitTypeId: number;
  qty: number;
}

export interface MoveArrivedPayload {
  heroId: number;
  nodeId: number;
}

export interface MoveUpdatePayload {
  heroId: number;
  fromNodeId: number;
  toNodeId: number;
  departAt: number;
  arriveAt: number;
  travelSeconds: number;
}

export interface CombatLogEntry {
  round: number;
  side: string;
  damage: number;
  defenderHpAfter?: number;
  attackerHpAfter?: number;
}

export interface CombatResolvedPayload {
  heroId: number;
  creepId: number;
  outcome: "win" | "loss" | string;
  goldReward: number;
  casualties: number;
  log: CombatLogEntry[];
}

export interface CastleTickPayload {
  castleId: number;
  gold: number;
  goldPerMin: number;
}

export interface MapNodeDTO {
  id: number;
  name: string;
  x: number;
  y: number;
  kind: string;
}

export interface MapEdgeDTO {
  id: number;
  fromNodeId: number;
  toNodeId: number;
  distanceUnits: number;
}

export interface MapSnapshot {
  nodes: MapNodeDTO[];
  edges: MapEdgeDTO[];
}

export interface HeroStatePayload {
  heroId: number;
  currentNodeId: number;
  armySize: number;
  upkeepGoldPerHour: number;
  speedEffective: number;
  /** Server epoch ms; hero cannot move while serverNow < respawnUntil. */
  respawnUntil?: number;
}

export interface HelloAckPayload {
  playerId: number;
  heroId: number;
  castleId: number;
  gold: number;
  mapSnapshot: MapSnapshot;
  heroState: HeroStatePayload;
  /** Present when hero has an in-flight movement_order (reconnect). */
  inFlight?: MoveUpdatePayload;
}
