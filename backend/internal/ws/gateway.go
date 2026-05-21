package ws

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/arrivals"
)

const (
	helloTimeoutSec = 5
	closeHelloTimeout = 4000
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// PoC: allow all origins (tighten post-PoC).
		return true
	},
}

// Gateway serves the WebSocket endpoint.
type Gateway struct {
	hub    *Hub
	store  *store.Store
	router *Router
	logger *slog.Logger
}

// NewGateway wires the WS hub and router. If hub is nil a new hub is created.
func NewGateway(st *store.Store, rdb *redisx.Client, sched *arrivals.Scheduler, hub *Hub, logger *slog.Logger) *Gateway {
	if hub == nil {
		hub = NewHub(logger)
	}
	return &Gateway{
		hub:    hub,
		store:  st,
		router: NewRouter(st, rdb, sched, hub),
		logger: logger,
	}
}

// Hub exposes the client hub for broadcasts (tick engine, BETA-002+).
func (g *Gateway) Hub() *Hub {
	return g.hub
}

func parseViewAsPlayerID(raw string) int64 {
	if raw == "" {
		return 0
	}
	var id int64
	_, _ = fmt.Sscanf(raw, "%d", &id)
	return id
}

// ServeHTTP upgrades GET /ws and runs the connection lifecycle.
func (g *Gateway) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		g.logger.Error("ws upgrade", slog.String("error", err.Error()))
		return
	}

	client := &Client{
		hub:    g.hub,
		conn:   conn,
		send:   make(chan []byte, sendBufferSize),
		viewForPlayerID: parseViewAsPlayerID(r.URL.Query().Get("viewAsPlayer")),
		logger: g.logger,
	}
	g.hub.Register(client)

	go client.writePump()
	g.runConnection(r.Context(), client)
	g.hub.Unregister(client)
	conn.Close()
}

func (g *Gateway) runConnection(ctx context.Context, c *Client) {
	c.conn.SetReadLimit(maxMessageSize)
	_ = c.conn.SetReadDeadline(time.Now().Add(helloTimeoutSec * time.Second))
	c.conn.SetPongHandler(func(string) error {
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	// Handshake: first message must be hello within 5s.
	_, raw, err := c.conn.ReadMessage()
	if err != nil {
		g.closeWithReason(c, closeHelloTimeout, proto.CodeHelloTimeout)
		return
	}

	env, err := proto.ParseInbound(raw)
	if err != nil || env.Type != proto.TypeHello {
		g.sendErrorEnvelope(c, proto.CodeHelloInvalidPayload, "expected hello envelope", nil)
		g.closeWithReason(c, closeHelloTimeout, proto.CodeHelloTimeout)
		return
	}

	ackData, playerID, err := g.router.HandleHello(ctx, env)
	if err != nil {
		if err == errUnknownPlayer {
			_ = c.conn.WriteMessage(websocket.TextMessage, ackData)
			g.closeWithReason(c, websocket.ClosePolicyViolation, proto.CodeHelloUnknownPlayer)
			return
		}
		g.logger.Error("hello failed", slog.String("error", err.Error()))
		g.sendErrorEnvelope(c, proto.CodeInternal, "hello failed", &env.Seq)
		return
	}

	c.playerID = playerID
	if c.viewForPlayerID <= 0 {
		c.viewForPlayerID = playerID
	}
	c.enqueue(ackData)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))

	// Post-handshake read loop.
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			return
		}
		_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))

		env, err := proto.ParseInbound(raw)
		if err != nil {
			g.sendErrorEnvelope(c, proto.CodeHelloInvalidPayload, "invalid json envelope", nil)
			continue
		}
		if err := g.router.Handle(ctx, c, env); err != nil {
			g.logger.Error("handle message", slog.String("type", env.Type), slog.String("error", err.Error()))
		}
	}
}

func (g *Gateway) sendErrorEnvelope(c *Client, code, message string, refSeq *int64) {
	env, err := proto.NewEnvelope(proto.TypeError, proto.NewErrorPayload(code, message, refSeq), 0)
	if err != nil {
		return
	}
	data, _ := json.Marshal(env)
	c.enqueue(data)
}

func (g *Gateway) closeWithReason(c *Client, code int, reason string) {
	_ = c.conn.WriteMessage(
		websocket.CloseMessage,
		websocket.FormatCloseMessage(code, reason),
	)
}

// Handler returns an http.Handler that requires a configured store.
func Handler(gw *Gateway) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if gw.store == nil {
			http.Error(w, "database not configured", http.StatusServiceUnavailable)
			return
		}
		gw.ServeHTTP(w, r)
	})
}
