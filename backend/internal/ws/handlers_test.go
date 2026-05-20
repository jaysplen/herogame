package ws_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/redisx"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/tick"
	"github.com/herogame/backend/internal/ws"
	"log/slog"
)

func testGateway(t *testing.T) (*store.Store, *redisx.Client, *ws.Gateway) {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable"
	}
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		redisURL = "redis://localhost:6379/0"
	}
	_, file, _, _ := runtime.Caller(0)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	_ = os.Setenv("MIGRATIONS_DIR", filepath.Join(backendRoot, "migrations"))
	if err := store.MigrateUp(dsn); err != nil {
		t.Skip("postgres:", err)
	}
	ctx := context.Background()
	st, err := store.New(ctx, dsn)
	if err != nil {
		t.Skip("postgres:", err)
	}
	rdb, err := redisx.New(ctx, redisURL)
	if err != nil {
		st.Close()
		t.Skip("redis:", err)
	}
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	hub := ws.NewHub(logger)
	eng := tick.NewEngine(st, rdb, hub, logger)
	_ = eng.Arrivals().Rehydrate(ctx)
	gw := ws.NewGateway(st, rdb, eng.Arrivals(), hub, logger)
	return st, rdb, gw
}

func dialHello(t *testing.T, url string, playerID int64) *websocket.Conn {
	t.Helper()
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		t.Fatal(err)
	}
	hello, _ := proto.NewEnvelope(proto.TypeHello, proto.HelloPayload{PlayerID: playerID}, 1)
	if err := conn.WriteJSON(hello); err != nil {
		t.Fatal(err)
	}
	var ack proto.Envelope
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatal(err)
	}
	if ack.Type != proto.TypeHelloAck {
		t.Fatalf("expected hello.ack, got %q", ack.Type)
	}
	return conn
}

func TestMoveRequestBroadcastToTwoClients(t *testing.T) {
	st, rdb, gw := testGateway(t)
	defer st.Close()
	defer rdb.Close()

	srv := httptest.NewServer(ws.Handler(gw))
	defer srv.Close()
	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")

	conn1 := dialHello(t, wsURL, 1)
	defer conn1.Close()
	conn2 := dialHello(t, wsURL, 2)
	defer conn2.Close()

	var wg sync.WaitGroup
	var mu sync.Mutex
	moveUpdates := 0

	readUntilMove := func(conn *websocket.Conn, label string) {
		defer wg.Done()
		_ = conn.SetReadDeadline(time.Now().Add(3 * time.Second))
		for {
			var env proto.Envelope
			if err := conn.ReadJSON(&env); err != nil {
				t.Errorf("%s read: %v", label, err)
				return
			}
			if env.Type == proto.TypeMoveUpdate {
				mu.Lock()
				moveUpdates++
				mu.Unlock()
				var p proto.MoveUpdatePayload
				_ = json.Unmarshal(env.Payload, &p)
				if p.HeroID != 1 || p.ToNodeID != 2 {
					t.Errorf("%s payload: %+v", label, p)
				}
				return
			}
		}
	}

	wg.Add(2)
	go readUntilMove(conn1, "c1")
	go readUntilMove(conn2, "c2")

	move, _ := proto.NewEnvelope(proto.TypeMoveRequest, proto.MoveRequestPayload{
		HeroID: 1, TargetNodeID: 2,
	}, 10)
	if err := conn1.WriteJSON(move); err != nil {
		t.Fatal(err)
	}

	wg.Wait()
	mu.Lock()
	n := moveUpdates
	mu.Unlock()
	if n != 2 {
		t.Fatalf("move.update count = %d, want 2", n)
	}

	// reset hero position for other tests
	ctx := context.Background()
	_, _ = st.Pool().Exec(ctx, `UPDATE heroes SET current_node_id = 1 WHERE id = 1`)
	_, _ = st.Pool().Exec(ctx, `DELETE FROM movement_orders WHERE hero_id = 1 AND status = 'in_flight'`)
	_ = rdb.Underlying().Del(ctx, "arrivals:zset")
}
