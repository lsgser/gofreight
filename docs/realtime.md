<!--
| --------------------------------------------------------------------------
| Realtime
| --------------------------------------------------------------------------
|
| Official Gofreight documentation consumed by the docs site and linked
| from the CLI welcome page.
|
| Keep examples aligned with the current CLI and version; sync copies to
| gofreight-web when publishing.
|
| Covers concepts, tutorials, and reference material for this topic.
|

-->

<!--
| --------------------------------------------------------------------------
| Real-time WebSockets
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

# Real-time WebSockets

Gofreight ships a **socket.io-style** WebSocket layer for rooms, events, and broadcasts — no third-party real-time server required for single-process apps.

## Quick start

### 1. Mount the socket endpoint

```go
// bootstrap/app.go
app.MountSocket("/socket") // default path: /socket
```

### 2. Handle events on the server

```go
import "encoding/json"

app.Channels.OnConnect(func(c *channels.Connection) {
    c.Join("chat:lobby")
})

app.Channels.On("chat:message", func(c *channels.Connection, raw json.RawMessage) {
    var msg struct{ Text string `json:"text"` }
    _ = json.Unmarshal(raw, &msg)

    app.Channels.To("chat:lobby").Emit("chat:message", map[string]string{
        "text": msg.Text,
        "id":   c.ID,
    })
})
```

### 3. Connect from the browser (TypeScript)

Copy `channels/gofreight-socket.ts` from the framework repo, or import it from your frontend bundle:

```typescript
import { GofreightSocket } from './gofreight-socket'

const socket = new GofreightSocket('/socket')

socket.on('connected', ({ id }) => console.log('connected', id))
socket.on('chat:message', (msg) => appendMessage(msg))

document.querySelector('form')?.addEventListener('submit', (e) => {
  e.preventDefault()
  const input = e.target.querySelector('input')
  socket.emit('chat:message', { text: input.value, user: 'You' })
  input.value = ''
})
```

## Server API

| Method | Description |
|--------|-------------|
| `app.MountSocket(path)` | Register WebSocket handler (default `/socket`) |
| `app.Channels.On(event, fn)` | Handle client-emitted events |
| `app.Channels.OnConnect(fn)` | Run when a client connects |
| `app.Channels.OnDisconnect(fn)` | Run when a client disconnects |
| `app.Channels.To(room).Emit(event, data)` | Broadcast to a room |
| `app.Channels.Emit(room, event, data)` | Same as `To(room).Emit(...)` |
| `app.Channels.Broadcast(channel, event, data)` | Legacy alias for `Emit` |

### Connection helpers

```go
c.Join("chat:lobby")   // subscribe to a room
c.Leave("chat:lobby")  // unsubscribe
c.Emit("typing", data) // emit to server handlers (same as client emit)
c.ID                   // unique connection id
c.Meta                 // attach custom metadata per connection
```

## Wire protocol

All messages are JSON text frames.

**Server → client (on connect):**

```json
{ "type": "connected", "data": { "id": "conn-a1b2c3d4" } }
```

**Client → server (join a room):**

```json
{ "type": "join", "room": "chat:lobby" }
```

**Client → server (leave a room):**

```json
{ "type": "leave", "room": "chat:lobby" }
```

**Client → server (emit event):**

```json
{ "type": "emit", "event": "chat:message", "data": { "text": "Hello" } }
```

**Server → client (broadcast):**

```json
{ "room": "chat:lobby", "event": "chat:message", "data": { "text": "Hello", "id": "conn-a1b2c3d4" } }
```

### Legacy channel subscribe

Channel subscribe via JSON message still works:

```json
{ "action": "subscribe", "channel": "posts" }
```

Use `app.Channels.Broadcast("posts", "created", payload)` to push events.

## TypeScript client API

`GofreightSocket` mirrors socket.io ergonomics:

```typescript
const socket = new GofreightSocket('/socket', {
  autoConnect: true,       // connect immediately (default)
  reconnect: true,         // reconnect on disconnect (default)
  reconnectDelayMs: 1500,
})

socket.connect()
socket.disconnect()

socket.on('connect', () => {})
socket.on('disconnect', () => {})
socket.on('connected', ({ id }) => {})   // server assigned id
socket.on('chat:message', (data) => {})

socket.emit('chat:message', { text: 'Hi' })
socket.join('chat:lobby')
socket.leave('chat:lobby')
socket.subscribe('posts')   // legacy channel subscribe

socket.id          // set after 'connected' event
socket.connected   // boolean
```

## Broadcasting from controllers

After creating a post, notify subscribers:

```go
func (c *PostsController) Store(w http.ResponseWriter, r *http.Request) {
    // ... save post ...

    c.App.Channels.To("posts").Emit("created", map[string]any{
        "id":    post.ID,
        "title": post.Title,
    })

    c.Redirect("/posts")
}
```

Clients join the room first:

```typescript
socket.join('posts')
socket.on('created', (post) => refreshFeed(post))
```

## Demo app

The `demoapp` includes a live chat example wired in `bootstrap/app.go`:

- Endpoint: `ws://localhost:5000/socket`
- Room: `chat:lobby`
- Events: `chat:message`, `chat:history`

Run `gofreight serve` inside `demoapp/` and connect with the TypeScript client.

## Production notes

- The built-in hub is **in-memory** by default — suitable for single-process deployments and development.
- For **multi-instance broadcast**, set `REDIS_URL` in `.env`. New apps call `app.UseRedisBroadcast(config.ResolveRedisURL())` when Redis is configured. Events publish to Redis pub/sub and fan out to all connected clients on every node.
- Set `CheckOrigin` appropriately before production (currently permissive for local dev).
- WebSocket routes skip CSRF for the upgrade handshake; protect sensitive rooms with session/JWT checks in `OnConnect`.

## See also

- [Features — real-time](features.md#11-real-time-websockets)
- [Tutorial: Real-time chat](https://github.com/lsgser/gofreight-web/blob/main/src/content/docs/tutorial-realtime.md)
