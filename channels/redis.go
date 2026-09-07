package channels

import (
	"context"
	"encoding/json"

	"github.com/redis/go-redis/v9"
)

type broadcastMessage struct {
	Room  string `json:"room"`
	Event string `json:"event"`
	Data  any    `json:"data"`
}

// RedisBroadcaster publishes WebSocket events across instances via Redis pub/sub.
type RedisBroadcaster struct {
	client  *redis.Client
	channel string
	hub     *Hub
}

// NewRedisBroadcaster connects a hub to Redis for horizontal scaling.
func NewRedisBroadcaster(hub *Hub, redisURL, channel string) (*RedisBroadcaster, error) {
	opts, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	if channel == "" {
		channel = "gofreight:broadcast"
	}
	rb := &RedisBroadcaster{
		client:  redis.NewClient(opts),
		channel: channel,
		hub:     hub,
	}
	go rb.listen()
	return rb, nil
}

// Publish emits an event to Redis and the local hub.
func (rb *RedisBroadcaster) Publish(room, event string, data any) error {
	payload, err := json.Marshal(broadcastMessage{Room: room, Event: event, Data: data})
	if err != nil {
		return err
	}
	if err := rb.client.Publish(context.Background(), rb.channel, payload).Err(); err != nil {
		return err
	}
	rb.hub.emitLocal(room, event, data)
	return nil
}

func (rb *RedisBroadcaster) listen() {
	sub := rb.client.Subscribe(context.Background(), rb.channel)
	for msg := range sub.Channel() {
		var bm broadcastMessage
		if json.Unmarshal([]byte(msg.Payload), &bm) != nil {
			continue
		}
		rb.hub.emitLocal(bm.Room, bm.Event, bm.Data)
	}
}

// UseRedisBroadcast configures the hub to publish via Redis.
func (h *Hub) UseRedisBroadcast(rb *RedisBroadcaster) {
	if h == nil || rb == nil {
		return
	}
	h.publish = rb.Publish
}
