<!--
| --------------------------------------------------------------------------
| Tutorial-realtime
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
| Tutorial-realtime
| --------------------------------------------------------------------------
|
| Official Gofreight documentation (published on the docs site).
|

-->

<p align="center">
  <img src="assets/gofreight-icon.png" alt="Gofreight" width="48">
</p>

# Real-time WebSockets

Gofreight includes a **socket.io-style** WebSocket layer for rooms, events, and broadcasts. Use it for live chat, notifications, dashboards, and collaborative features.

## What you'll build

A live chat room where multiple browser tabs exchange messages in real time over WebSockets.

## Prerequisites

- Completed [Your First App](tutorial-first-app.md)
- Go 1.22+

## Step 1 — Mount the socket endpoint

In `bootstrap/app.go`:

```go
import "github.com/lsgser/gofreight/channels"

func wireRealtime(app *application.Application) {
    app.MountSocket("/socket")
}
```

Call `wireRealtime(app)` from your `Application()` bootstrap function.

The WebSocket URL becomes `ws://localhost:5000/socket`.

## Step 2 — Handle events on the server

```go
import "encoding/json"

func wireRealtime(app *application.Application) {
    app.Channels.OnConnect(func(c *channels.Connection) {
        c.Join("chat:lobby")
    })

    app.Channels.On("chat:message", func(c *channels.Connection, raw json.RawMessage) {
        var payload struct {
            Text string `json:"text"`
            User string `json:"user"`
        }
        if json.Unmarshal(raw, &payload) != nil || payload.Text == "" {
            return
        }
        if payload.User == "" {
            payload.User = "guest"
        }

        app.Channels.To("chat:lobby").Emit("chat:message", map[string]string{
            "text": payload.Text,
            "user": payload.User,
            "id":   c.ID,
        })
    })

    app.MountSocket("/socket")
}
```

## Step 3 — Add the TypeScript client

Copy `gofreight-socket.ts` from the framework repo (`channels/gofreight-socket.ts`) into your frontend `public/` or `assets/` folder.

```typescript
import { GofreightSocket } from './gofreight-socket'

const socket = new GofreightSocket('/socket')

socket.on('connected', ({ id }) => {
  console.log('Connected as', id)
  socket.emit('chat:history', {})
})

socket.on('chat:message', (msg: { text: string; user: string }) => {
  const li = document.createElement('li')
  li.textContent = `${msg.user}: ${msg.text}`
  document.querySelector('#messages')?.appendChild(li)
})

document.querySelector('#chat-form')?.addEventListener('submit', (e) => {
  e.preventDefault()
  const input = document.querySelector<HTMLInputElement>('#chat-input')
  if (!input?.value.trim()) return
  socket.emit('chat:message', { text: input.value, user: 'You' })
  input.value = ''
})
```

## Step 4 — HTML page

```html
<ul id="messages"></ul>
<form id="chat-form">
  <input id="chat-input" type="text" placeholder="Say hello…" autocomplete="off" />
  <button type="submit">Send</button>
</form>
<script type="module" src="/assets/chat.js"></script>
```

Open two browser tabs to the page — messages appear instantly in both.

## Client API reference

| Method | Description |
|--------|-------------|
| `new GofreightSocket('/socket')` | Connect to the server |
| `socket.on(event, fn)` | Listen for server events |
| `socket.emit(event, data)` | Send an event to the server |
| `socket.join(room)` | Join a room |
| `socket.leave(room)` | Leave a room |
| `socket.id` | Connection id (after `connected` event) |
| `socket.connected` | Whether the socket is open |

## Broadcasting from controllers

Notify clients after a database change:

```go
app.Channels.To("posts").Emit("created", map[string]any{
    "id":    post.ID,
    "title": post.Title,
})
```

Clients subscribe with:

```typescript
socket.join('posts')
socket.on('created', (post) => refreshFeed(post))
```

## Demo app

The included `demoapp` has a working chat backend in `bootstrap/app.go`. Run:

```bash
cd demoapp
gofreight serve
```

Connect to `ws://localhost:5000/socket` and join room `chat:lobby`.

## Next steps

- Read the full [Real-time WebSockets](realtime.md) guide
- Add JWT checks in `OnConnect` for private rooms
- See [Security](security.md) for production hardening
