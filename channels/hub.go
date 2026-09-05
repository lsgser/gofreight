package channels

import (
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// Hub manages WebSocket channels (Action Cable / Laravel Echo style).
type Hub struct {
	mu       sync.RWMutex
	clients  map[string]map[*Client]bool
	channels map[string]map[*Client]bool
}

// Client is a connected WebSocket client.
type Client struct {
	hub      *Hub
	conn     *websocket.Conn
	send     chan []byte
	channels map[string]bool
}

// NewHub creates a channel hub.
func NewHub() *Hub {
	return &Hub{
		clients:  make(map[string]map[*Client]bool),
		channels: make(map[string]map[*Client]bool),
	}
}

// Message is a channel broadcast payload.
type Message struct {
	Channel string `json:"channel"`
	Event   string `json:"event"`
	Data    any    `json:"data"`
}

// Subscribe adds a client to a channel.
func (h *Hub) Subscribe(client *Client, channel string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.channels[channel] == nil {
		h.channels[channel] = make(map[*Client]bool)
	}
	h.channels[channel][client] = true
	client.channels[channel] = true
}

// Broadcast sends an event to all subscribers on a channel.
func (h *Hub) Broadcast(channel, event string, data any) {
	msg, _ := json.Marshal(Message{Channel: channel, Event: event, Data: data})
	h.mu.RLock()
	defer h.mu.RUnlock()
	for client := range h.channels[channel] {
		select {
		case client.send <- msg:
		default:
		}
	}
}

// Handler upgrades HTTP to WebSocket and handles channel protocol.
func (h *Hub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}
		client := &Client{hub: h, conn: conn, send: make(chan []byte, 64), channels: make(map[string]bool)}
		go client.writePump()
		client.readPump()
	}
}

func (c *Client) readPump() {
	defer c.conn.Close()
	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		var cmd struct {
			Action  string `json:"action"`
			Channel string `json:"channel"`
		}
		if json.Unmarshal(raw, &cmd) != nil {
			continue
		}
		if cmd.Action == "subscribe" && cmd.Channel != "" {
			c.hub.Subscribe(c, cmd.Channel)
		}
	}
}

func (c *Client) writePump() {
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

// Mount registers the WebSocket endpoint.
func (h *Hub) Mount(r interface {
	Mount(prefix string, handler http.Handler)
}, path string) {
	r.Mount(path, http.HandlerFunc(h.Handler()))
}
