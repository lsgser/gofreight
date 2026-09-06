package channels

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

// EventHandler handles a client-emitted event.
type EventHandler func(c *Connection, data json.RawMessage)

// Hub manages WebSocket connections, rooms, and events (socket.io-style).
type Hub struct {
	mu           sync.RWMutex
	rooms        map[string]map[*Connection]bool
	connections  map[*Connection]bool
	handlers     map[string]EventHandler
	onConnect    func(*Connection)
	onDisconnect func(*Connection)
}

// Connection is a connected WebSocket client.
type Connection struct {
	Hub     *Hub
	ID      string
	conn    *websocket.Conn
	send    chan []byte
	rooms   map[string]bool
	Meta    map[string]any
}

// Client is an alias for Connection (backward compatibility).
type Client = Connection

// Message is a broadcast payload (legacy channel format).
type Message struct {
	Channel string `json:"channel,omitempty"`
	Room    string `json:"room,omitempty"`
	Event   string `json:"event"`
	Data    any    `json:"data"`
}

type wireEnvelope struct {
	Type    string          `json:"type"`
	Action  string          `json:"action"`
	Event   string          `json:"event"`
	Room    string          `json:"room"`
	Channel string          `json:"channel"`
	Data    json.RawMessage `json:"data"`
}

type connectedPayload struct {
	ID string `json:"id"`
}

// RoomEmitter targets emits to a single room.
type RoomEmitter struct {
	hub  *Hub
	room string
}

// NewHub creates a WebSocket hub.
func NewHub() *Hub {
	return &Hub{
		rooms:       make(map[string]map[*Connection]bool),
		connections: make(map[*Connection]bool),
		handlers:    make(map[string]EventHandler),
	}
}

// On registers a handler for client-emitted events.
func (h *Hub) On(event string, handler EventHandler) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.handlers[event] = handler
}

// OnConnect registers a callback when a client connects.
func (h *Hub) OnConnect(fn func(*Connection)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onConnect = fn
}

// OnDisconnect registers a callback when a client disconnects.
func (h *Hub) OnDisconnect(fn func(*Connection)) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.onDisconnect = fn
}

// To returns an emitter scoped to a room.
func (h *Hub) To(room string) *RoomEmitter {
	return &RoomEmitter{hub: h, room: room}
}

// Emit sends an event to all connections in a room.
func (h *Hub) Emit(room, event string, data any) {
	h.emit(room, event, data)
}

// Emit sends an event to all connections in the emitter's room.
func (re *RoomEmitter) Emit(event string, data any) {
	re.hub.emit(re.room, event, data)
}

// Broadcast sends an event to all subscribers on a channel (legacy alias for Emit).
func (h *Hub) Broadcast(channel, event string, data any) {
	h.Emit(channel, event, data)
}

// Subscribe adds a client to a channel (legacy alias for Join).
func (h *Hub) Subscribe(client *Connection, channel string) {
	client.Join(channel)
}

func (h *Hub) emit(room, event string, data any) {
	payload, err := json.Marshal(Message{Room: room, Channel: room, Event: event, Data: data})
	if err != nil {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()
	for conn := range h.rooms[room] {
		select {
		case conn.send <- payload:
		default:
		}
	}
}

// Join adds the connection to a room.
func (c *Connection) Join(room string) {
	if room == "" {
		return
	}
	c.Hub.mu.Lock()
	defer c.Hub.mu.Unlock()
	if c.Hub.rooms[room] == nil {
		c.Hub.rooms[room] = make(map[*Connection]bool)
	}
	c.Hub.rooms[room][c] = true
	c.rooms[room] = true
}

// Leave removes the connection from a room.
func (c *Connection) Leave(room string) {
	if room == "" {
		return
	}
	c.Hub.mu.Lock()
	defer c.Hub.mu.Unlock()
	delete(c.Hub.rooms[room], c)
	delete(c.rooms, room)
	if len(c.Hub.rooms[room]) == 0 {
		delete(c.Hub.rooms, room)
	}
}

// Emit sends an event to the server (handled by Hub.On handlers).
func (c *Connection) Emit(event string, data any) {
	raw, err := json.Marshal(data)
	if err != nil {
		return
	}
	c.Hub.dispatch(c, event, raw)
}

func (c *Connection) sendJSON(v any) {
	payload, err := json.Marshal(v)
	if err != nil {
		return
	}
	select {
	case c.send <- payload:
	default:
	}
}

func (h *Hub) dispatch(c *Connection, event string, data json.RawMessage) {
	h.mu.RLock()
	handler := h.handlers[event]
	h.mu.RUnlock()
	if handler != nil {
		handler(c, data)
	}
}

func (h *Hub) addConnection(c *Connection) {
	h.mu.Lock()
	h.connections[c] = true
	onConnect := h.onConnect
	h.mu.Unlock()

	c.sendJSON(wireEnvelope{Type: "connected", Data: mustRaw(connectedPayload{ID: c.ID})})
	if onConnect != nil {
		onConnect(c)
	}
}

func (h *Hub) removeConnection(c *Connection) {
	h.mu.Lock()
	delete(h.connections, c)
	for room := range c.rooms {
		delete(h.rooms[room], c)
		if len(h.rooms[room]) == 0 {
			delete(h.rooms, room)
		}
	}
	onDisconnect := h.onDisconnect
	h.mu.Unlock()

	if onDisconnect != nil {
		onDisconnect(c)
	}
}

// Handler upgrades HTTP to WebSocket and handles the socket protocol.
func (h *Hub) Handler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &Connection{
			Hub:   h,
			ID:    newConnectionID(),
			conn:  conn,
			send:  make(chan []byte, 64),
			rooms: make(map[string]bool),
			Meta:  make(map[string]any),
		}

		h.addConnection(client)
		go client.writePump()
		client.readPump()
	}
}

func (c *Connection) readPump() {
	defer func() {
		c.Hub.removeConnection(c)
		close(c.send)
		c.conn.Close()
	}()

	for {
		_, raw, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
		c.handleMessage(raw)
	}
}

func (c *Connection) handleMessage(raw []byte) {
	var env wireEnvelope
	if json.Unmarshal(raw, &env) != nil {
		return
	}

	// Legacy Action Cable-style subscribe.
	if env.Action == "subscribe" {
		room := env.Channel
		if room != "" {
			c.Join(room)
		}
		return
	}

	switch env.Type {
	case "join":
		c.Join(env.Room)
	case "leave":
		c.Leave(env.Room)
	case "emit":
		if env.Event != "" {
			c.Hub.dispatch(c, env.Event, env.Data)
		}
	default:
		if env.Event != "" {
			c.Hub.dispatch(c, env.Event, env.Data)
		}
	}
}

func (c *Connection) writePump() {
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

// Mount registers the WebSocket endpoint on a router.
func (h *Hub) Mount(r interface {
	Mount(prefix string, handler http.Handler)
}, path string) {
	r.Mount(path, http.HandlerFunc(h.Handler()))
}

func newConnectionID() string {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "conn-unknown"
	}
	return "conn-" + hex.EncodeToString(b[:])
}

func mustRaw(v any) json.RawMessage {
	raw, err := json.Marshal(v)
	if err != nil {
		return json.RawMessage(`{}`)
	}
	return raw
}
