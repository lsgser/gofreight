package channels

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/websocket"
)

func TestHubEmitToRoom(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(hub.Handler())
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	var connected wireEnvelope
	if err := conn.ReadJSON(&connected); err != nil {
		t.Fatalf("read connected: %v", err)
	}
	if connected.Type != "connected" {
		t.Fatalf("expected connected, got %q", connected.Type)
	}

	join, _ := json.Marshal(wireEnvelope{Type: "join", Room: "chat"})
	if err := conn.WriteMessage(websocket.TextMessage, join); err != nil {
		t.Fatalf("join: %v", err)
	}
	time.Sleep(50 * time.Millisecond)

	hub.Emit("chat", "message", map[string]string{"text": "hello"})

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read message: %v", err)
	}
	if msg.Event != "message" {
		t.Fatalf("expected message event, got %q", msg.Event)
	}
}

func TestHubOnEvent(t *testing.T) {
	hub := NewHub()
	received := make(chan json.RawMessage, 1)
	hub.On("ping", func(c *Connection, data json.RawMessage) {
		received <- data
		c.Hub.To("lobby").Emit("pong", map[string]string{"ok": "true"})
	})

	server := httptest.NewServer(hub.Handler())
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	_, _, _ = conn.ReadMessage() // connected

	join, _ := json.Marshal(wireEnvelope{Type: "join", Room: "lobby"})
	_ = conn.WriteMessage(websocket.TextMessage, join)

	emit, _ := json.Marshal(wireEnvelope{Type: "emit", Event: "ping", Data: json.RawMessage(`{"n":1}`)})
	if err := conn.WriteMessage(websocket.TextMessage, emit); err != nil {
		t.Fatalf("emit ping: %v", err)
	}

	select {
	case data := <-received:
		if string(data) != `{"n":1}` {
			t.Fatalf("unexpected payload: %s", data)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler not called")
	}

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read pong: %v", err)
	}
	if msg.Event != "pong" {
		t.Fatalf("expected pong, got %q", msg.Event)
	}
}

func TestLegacySubscribeBroadcast(t *testing.T) {
	hub := NewHub()
	server := httptest.NewServer(http.HandlerFunc(hub.Handler()))
	defer server.Close()

	wsURL := "ws" + strings.TrimPrefix(server.URL, "http")
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	_, _, _ = conn.ReadMessage()

	sub, _ := json.Marshal(map[string]string{"action": "subscribe", "channel": "posts"})
	_ = conn.WriteMessage(websocket.TextMessage, sub)
	time.Sleep(50 * time.Millisecond)

	hub.Broadcast("posts", "created", map[string]int{"id": 1})

	conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	var msg Message
	if err := conn.ReadJSON(&msg); err != nil {
		t.Fatalf("read broadcast: %v", err)
	}
	if msg.Event != "created" {
		t.Fatalf("expected created, got %q", msg.Event)
	}
}
