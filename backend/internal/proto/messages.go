package proto

// Client → server message types.
const (
	TypeHello       = "hello"
	TypeMoveRequest = "move.request"
	TypeUnitBuy     = "unit.buy"
	TypeCastleBuild = "castle.build"
)

// MoveArrivedPayload is emitted when a hero finishes travel.
type MoveArrivedPayload struct {
	HeroID int64 `json:"heroId"`
	NodeID int64 `json:"nodeId"`
}

// Server → client message types.
const (
	TypeHelloAck       = "hello.ack"
	TypeMoveUpdate     = "move.update"
	TypeMoveArrived    = "move.arrived"
	TypeCombatResolved = "combat.resolved"
	TypeCastleTick     = "castle.tick"
	TypeHeroState      = "hero.state"
	TypeCreepState     = "creep.state"
	TypeResourceState  = "resource.state"
	TypeObjectiveState = "objective.state"
	TypeError          = "error"
)

// HelloPayload is the client handshake (architecture.md §7.2).
type HelloPayload struct {
	PlayerID int64 `json:"playerId"`
}

// MoveRequestPayload is a client move command.
type MoveRequestPayload struct {
	HeroID       int64 `json:"heroId"`
	TargetNodeID int64 `json:"targetNodeId"`
}

// MoveUpdatePayload confirms or broadcasts a movement order.
type MoveUpdatePayload struct {
	HeroID        int64 `json:"heroId"`
	FromNodeID    int64 `json:"fromNodeId"`
	ToNodeID      int64 `json:"toNodeId"`
	DepartAt      int64 `json:"departAt"`
	ArriveAt      int64 `json:"arriveAt"`
	TravelSeconds int   `json:"travelSeconds"`
}

// UnitBuyPayload purchases units at a castle.
type UnitBuyPayload struct {
	CastleID   int64 `json:"castleId"`
	UnitTypeID int64 `json:"unitTypeId"`
	Qty        int   `json:"qty"`
}

// CastleBuildPayload upgrades one castle building.
type CastleBuildPayload struct {
	CastleID      int64  `json:"castleId"`
	BuildingCode  string `json:"buildingCode"`
}

// CombatLogEntry is one round in a combat log (game_rules.md §6.3).
type CombatLogEntry struct {
	Round           int    `json:"round"`
	Side            string `json:"side"`
	Damage          int    `json:"damage"`
	DefenderHPAfter *int   `json:"defenderHpAfter,omitempty"`
	AttackerHPAfter *int   `json:"attackerHpAfter,omitempty"`
}

// CombatResolvedPayload is emitted after auto-combat (architecture.md §7.2).
type CombatResolvedPayload struct {
	HeroID     int64            `json:"heroId"`
	CreepID    int64            `json:"creepId"`
	EnemyHeroID *int64          `json:"enemyHeroId,omitempty"`
	Outcome    string           `json:"outcome"`
	GoldReward int32            `json:"goldReward"`
	Casualties int              `json:"casualties"`
	ConvertedUnits int          `json:"convertedUnits"`
	Log        []CombatLogEntry `json:"log"`
}

// CastleTickPayload is a throttled economy update.
type CastleTickPayload struct {
	CastleID   int64   `json:"castleId"`
	Gold       float64 `json:"gold"`
	GoldPerMin int32  `json:"goldPerMin"`
}

// MapNodeDTO is a node in the bootstrap map snapshot.
type MapNodeDTO struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
	X    int32  `json:"x"`
	Y    int32  `json:"y"`
	Kind string `json:"kind"`
}

// MapEdgeDTO is an edge in the bootstrap map snapshot.
type MapEdgeDTO struct {
	ID            int64 `json:"id"`
	FromNodeID    int64 `json:"fromNodeId"`
	ToNodeID      int64 `json:"toNodeId"`
	DistanceUnits int32 `json:"distanceUnits"`
}

// MapSnapshot is the full static map for the client.
type MapSnapshot struct {
	Nodes []MapNodeDTO `json:"nodes"`
	Edges []MapEdgeDTO `json:"edges"`
}

// HeroUnitStackDTO is one row of the hero army or castle shop catalog.
type HeroUnitStackDTO struct {
	UnitID   int64  `json:"unitId"`
	Code     string `json:"code"`
	Name     string `json:"name"`
	Qty      int    `json:"qty"`
	CostGold int32  `json:"costGold"`
	CostMetal int32 `json:"costMetal"`
	CostGems  int32 `json:"costGems"`
	CostCoal  int32 `json:"costCoal"`
	CostWood  int32 `json:"costWood"`
	CostStone int32 `json:"costStone"`
	Tier      int32 `json:"tier"`
	Faction   string `json:"faction"`
}

// ResourceBagDTO is the full player resource wallet.
type ResourceBagDTO struct {
	Gold  float64 `json:"gold"`
	Metal float64 `json:"metal"`
	Gems  float64 `json:"gems"`
	Coal  float64 `json:"coal"`
	Wood  float64 `json:"wood"`
	Stone float64 `json:"stone"`
}

// ResourceNodeDTO describes one conquerable resource node.
type ResourceNodeDTO struct {
	ID            int64  `json:"id"`
	NodeID        int64  `json:"nodeId"`
	ResourceType  string `json:"resourceType"`
	PerMin        int32  `json:"perMin"`
	OwnerPlayerID *int64 `json:"ownerPlayerId,omitempty"`
}

// CastleStateDTO exposes progression/building values to clients.
type CastleStateDTO struct {
	CastleID      int64  `json:"castleId"`
	PlayerID      int64  `json:"playerId"`
	NodeID        int64  `json:"nodeId"`
	Faction       string `json:"faction"`
	DefenseBonus  int32  `json:"defenseBonus"`
	BarracksTier  int32  `json:"barracksTier"`
	AcademyTier   int32  `json:"academyTier"`
}

// CreepStateDTO is one roaming neutral army state.
type CreepStateDTO struct {
	ID           int64  `json:"id"`
	Name         string `json:"name"`
	NodeID       int64  `json:"nodeId"`
	Qty          int32  `json:"qty"`
	Alive        bool   `json:"alive"`
	Attack       int32  `json:"attack"`
	Defense      int32  `json:"defense"`
	HP           int32  `json:"hp"`
	GraceUntil   *int64 `json:"graceUntil,omitempty"`
	FromNodeID   *int64 `json:"fromNodeId,omitempty"`
	ToNodeID     *int64 `json:"toNodeId,omitempty"`
	DepartAt     *int64 `json:"departAt,omitempty"`
	ArriveAt     *int64 `json:"arriveAt,omitempty"`
}

// CreepStatePayload broadcasts all alive creeps.
type CreepStatePayload struct {
	Creeps []CreepStateDTO `json:"creeps"`
}

// ResourceStatePayload broadcasts node ownership and wallet updates.
type ResourceStatePayload struct {
	PlayerID      int64             `json:"playerId"`
	Resources     ResourceBagDTO    `json:"resources"`
	ResourceNodes []ResourceNodeDTO `json:"resourceNodes"`
}

// ObjectiveStatePayload tracks elimination objective progress.
type ObjectiveStatePayload struct {
	PlayerID        int64 `json:"playerId"`
	EnemyPlayerID   int64 `json:"enemyPlayerId"`
	EnemyHeroKills  int32 `json:"enemyHeroKills"`
	TargetHeroKills int32 `json:"targetHeroKills"`
}

// HeroStatePayload is the hero slice of hello.ack / hero.state.
type HeroStatePayload struct {
	HeroID            int64              `json:"heroId"`
	PlayerID          int64              `json:"playerId"`
	CurrentNodeID     int64              `json:"currentNodeId"`
	ArmySize          int                `json:"armySize"`
	Units             []HeroUnitStackDTO `json:"units"`
	UpkeepGoldPerHour float64            `json:"upkeepGoldPerHour"`
	SpeedEffective    float64            `json:"speedEffective"`
	// RespawnUntil is server epoch ms; set while hero cannot move after defeat.
	RespawnUntil *int64 `json:"respawnUntil,omitempty"`
}

// HelloAckPayload is the bootstrap snapshot after handshake.
type HelloAckPayload struct {
	PlayerID    int64            `json:"playerId"`
	HeroID      int64            `json:"heroId"`
	CastleID    int64            `json:"castleId"`
	Gold        float64          `json:"gold"`
	Resources   ResourceBagDTO   `json:"resources"`
	MapSnapshot MapSnapshot      `json:"mapSnapshot"`
	HeroState   HeroStatePayload `json:"heroState"`
	// ShopUnits is the castle recruit catalog (qty 0); same shape as army stacks.
	ShopUnits []HeroUnitStackDTO `json:"shopUnits"`
	Castles   []CastleStateDTO   `json:"castles"`
	Creeps    []CreepStateDTO    `json:"creeps"`
	ResourceNodes []ResourceNodeDTO `json:"resourceNodes"`
	Objective ObjectiveStatePayload `json:"objective"`
	// InFlight is set when the hero has an active movement_order (reconnect replay).
	InFlight *MoveUpdatePayload `json:"inFlight,omitempty"`
}
