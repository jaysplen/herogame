package proto

// Client → server message types.
const (
	TypeHello       = "hello"
	TypeMoveRequest = "move.request"
	TypeUnitBuy     = "unit.buy"
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
	TypeError          = "error"
)

// HelloPayload is the client handshake (architecture.md §7.2).
type HelloPayload struct {
	PlayerID int64 `json:"playerId"`
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

// HeroStatePayload is the hero slice of hello.ack / hero.state.
type HeroStatePayload struct {
	HeroID            int64   `json:"heroId"`
	CurrentNodeID     int64   `json:"currentNodeId"`
	ArmySize          int     `json:"armySize"`
	UpkeepGoldPerHour float64 `json:"upkeepGoldPerHour"`
	SpeedEffective    float64 `json:"speedEffective"`
}

// HelloAckPayload is the bootstrap snapshot after handshake.
type HelloAckPayload struct {
	PlayerID    int64            `json:"playerId"`
	HeroID      int64            `json:"heroId"`
	CastleID    int64            `json:"castleId"`
	Gold        float64          `json:"gold"`
	MapSnapshot MapSnapshot      `json:"mapSnapshot"`
	HeroState   HeroStatePayload `json:"heroState"`
}
