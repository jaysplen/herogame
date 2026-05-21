package proto

// Stable error codes for error envelopes (architecture.md §7.4).
const (
	CodeHelloTimeout        = "HELLO_TIMEOUT"
	CodeHelloUnknownPlayer  = "HELLO_UNKNOWN_PLAYER"
	CodeHelloInvalidPayload = "HELLO_INVALID_PAYLOAD"
	CodeMoveInvalidEdge     = "MOVE_INVALID_EDGE"
	CodeMoveHeroInFlight    = "MOVE_HERO_IN_FLIGHT"
	CodeMoveHeroRespawning  = "MOVE_HERO_RESPAWNING"
	CodeMoveGracePeriod     = "MOVE_GRACE_PERIOD"
	CodeBuyInsufficientGold = "BUY_INSUFFICIENT_GOLD"
	CodeBuyTierLocked       = "BUY_TIER_LOCKED"
	CodeBuyWrongFaction     = "BUY_WRONG_FACTION"
	CodeBuildInvalid        = "BUILD_INVALID"
	CodeBuildInsufficientResources = "BUILD_INSUFFICIENT_RESOURCES"
	CodeUnknownMessage      = "UNKNOWN_MESSAGE"
	CodeInternal            = "INTERNAL_ERROR"
)

// ErrorPayload is the shape of error envelope payloads.
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	RefSeq  *int64 `json:"refSeq,omitempty"`
}

// NewErrorPayload builds a client-facing error payload.
func NewErrorPayload(code, message string, refSeq *int64) ErrorPayload {
	return ErrorPayload{Code: code, Message: message, RefSeq: refSeq}
}
