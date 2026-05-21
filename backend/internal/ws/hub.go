package ws

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/herogame/backend/internal/proto"
)

// Hub tracks active WebSocket clients for PoC-wide broadcast.
type Hub struct {
	mu      sync.RWMutex
	clients map[*Client]struct{}
	logger  *slog.Logger
}

// NewHub creates an empty client hub.
func NewHub(logger *slog.Logger) *Hub {
	return &Hub{
		clients: make(map[*Client]struct{}),
		logger:  logger,
	}
}

// Register adds a connected client.
func (h *Hub) Register(c *Client) {
	h.mu.Lock()
	h.clients[c] = struct{}{}
	h.mu.Unlock()
}

// Unregister removes a client and closes its send channel.
func (h *Hub) Unregister(c *Client) {
	h.mu.Lock()
	if _, ok := h.clients[c]; ok {
		delete(h.clients, c)
		close(c.send)
	}
	h.mu.Unlock()
}

// Broadcast sends an envelope to every connected client (PoC: full map visibility).
func (h *Hub) Broadcast(env proto.Envelope) {
	data, err := json.Marshal(env)
	if err != nil {
		h.logger.Error("broadcast marshal", slog.String("error", err.Error()))
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		select {
		case c.send <- data:
		default:
			h.logger.Warn("client send buffer full, dropping message",
				slog.Int64("player_id", c.playerID))
		}
	}
}

// BroadcastToPlayer sends an envelope only to sessions of one player.
func (h *Hub) BroadcastToPlayer(playerID int64, env proto.Envelope) {
	data, err := json.Marshal(env)
	if err != nil {
		h.logger.Error("broadcast marshal", slog.String("error", err.Error()))
		return
	}
	h.mu.RLock()
	defer h.mu.RUnlock()
	for c := range h.clients {
		if c.playerID != playerID {
			continue
		}
		select {
		case c.send <- data:
		default:
			h.logger.Warn("client send buffer full, dropping message",
				slog.Int64("player_id", c.playerID))
		}
	}
}
