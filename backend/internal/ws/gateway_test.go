package ws_test

import (
	"context"
	"encoding/json"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	"github.com/herogame/backend/internal/proto"
	"github.com/herogame/backend/internal/store"
	"github.com/herogame/backend/internal/ws"
	"log/slog"
)

func testStore(t *testing.T) *store.Store {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://herogame:herogame@localhost:5432/herogame?sslmode=disable"
	}
	_, file, _, _ := runtime.Caller(0)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(file), "..", ".."))
	_ = os.Setenv("MIGRATIONS_DIR", filepath.Join(backendRoot, "migrations"))
	if err := store.MigrateUp(dsn); err != nil {
		t.Skip("postgres unavailable:", err)
	}
	st, err := store.New(context.Background(), dsn)
	if err != nil {
		t.Skip("postgres connect:", err)
	}
	return st
}

func TestHelloAck(t *testing.T) {
	st := testStore(t)
	defer st.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	gw := ws.NewGateway(st, logger)
	srv := httptest.NewServer(ws.Handler(gw))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	hello, err := proto.NewEnvelope(proto.TypeHello, proto.HelloPayload{PlayerID: 1}, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := conn.WriteJSON(hello); err != nil {
		t.Fatal(err)
	}
	_ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	var ack proto.Envelope
	if err := conn.ReadJSON(&ack); err != nil {
		t.Fatal(err)
	}
	if ack.Type != proto.TypeHelloAck {
		t.Fatalf("type = %q", ack.Type)
	}
	var payload proto.HelloAckPayload
	if err := ack.DecodePayload(&payload); err != nil {
		t.Fatal(err)
	}
	if payload.PlayerID != 1 || payload.HeroID != 1 || len(payload.MapSnapshot.Nodes) != 6 {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestHelloUnknownPlayer(t *testing.T) {
	st := testStore(t)
	defer st.Close()

	logger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	gw := ws.NewGateway(st, logger)
	srv := httptest.NewServer(ws.Handler(gw))
	defer srv.Close()

	wsURL := "ws" + strings.TrimPrefix(srv.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()

	hello, _ := proto.NewEnvelope(proto.TypeHello, proto.HelloPayload{PlayerID: 9999}, 2)
	_ = conn.WriteJSON(hello)

	var env proto.Envelope
	_ = conn.ReadJSON(&env)
	if env.Type != proto.TypeError {
		t.Fatalf("type = %q", env.Type)
	}
	var errPayload proto.ErrorPayload
	_ = json.Unmarshal(env.Payload, &errPayload)
	if errPayload.Code != proto.CodeHelloUnknownPlayer {
		t.Fatalf("code = %q", errPayload.Code)
	}
}
