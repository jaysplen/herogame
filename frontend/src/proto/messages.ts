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

export interface CastleBuildPayload {
  castleId: number;
  buildingCode: string;
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
  enemyHeroId?: number;
  outcome: "win" | "loss" | string;
  goldReward: number;
  casualties: number;
  convertedUnits: number;
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

export interface HeroUnitStackDTO {
  unitId: number;
  code: string;
  name: string;
  qty: number;
  costGold: number;
  costMetal: number;
  costGems: number;
  costCoal: number;
  costWood: number;
  costStone: number;
  tier: number;
  faction: string;
}

export interface ResourceBagDTO {
  gold: number;
  metal: number;
  gems: number;
  coal: number;
  wood: number;
  stone: number;
}

export interface ResourceNodeDTO {
  id: number;
  nodeId: number;
  resourceType: string;
  perMin: number;
  ownerPlayerId?: number;
}

export interface CastleStateDTO {
  castleId: number;
  playerId: number;
  nodeId: number;
  faction: string;
  defenseBonus: number;
  barracksTier: number;
  academyTier: number;
}

export interface CreepStateDTO {
  id: number;
  name: string;
  nodeId: number;
  qty: number;
  alive: boolean;
  attack: number;
  defense: number;
  hp: number;
  graceUntil?: number;
  fromNodeId?: number;
  toNodeId?: number;
  departAt?: number;
  arriveAt?: number;
}

export interface CreepStatePayload {
  creeps: CreepStateDTO[];
}

export interface ResourceStatePayload {
  playerId: number;
  resources: ResourceBagDTO;
  resourceNodes: ResourceNodeDTO[];
}

export interface ObjectiveStatePayload {
  playerId: number;
  enemyPlayerId: number;
  enemyHeroKills: number;
  targetHeroKills: number;
}

export interface HeroStatePayload {
  heroId: number;
  playerId: number;
  currentNodeId: number;
  armySize: number;
  units: HeroUnitStackDTO[];
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
  resources: ResourceBagDTO;
  mapSnapshot: MapSnapshot;
  heroState: HeroStatePayload;
  /** Castle recruit catalog (qty 0 per row). */
  shopUnits: HeroUnitStackDTO[];
  castles: CastleStateDTO[];
  creeps: CreepStateDTO[];
  resourceNodes: ResourceNodeDTO[];
  objective: ObjectiveStatePayload;
  /** Present when hero has an in-flight movement_order (reconnect). */
  inFlight?: MoveUpdatePayload;
}
