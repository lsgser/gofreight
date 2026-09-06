/**
 * GofreightSocket — simplified socket.io-style client for Gofreight WebSockets.
 *
 * Usage:
 *   const socket = new GofreightSocket('/socket')
 *   socket.on('connected', ({ id }) => console.log(id))
 *   socket.join('chat:lobby')
 *   socket.on('chat:message', (msg) => render(msg))
 *   socket.emit('chat:message', { text: 'Hello' })
 */

export type GofreightSocketOptions = {
  autoConnect?: boolean
  reconnect?: boolean
  reconnectDelayMs?: number
  /** Override host (e.g. `localhost:5000`). Defaults to `window.location.host`. */
  host?: string
  /** Force secure WebSocket (`wss:`). Auto-detected from the page when omitted. */
  secure?: boolean
}

type WireEnvelope = {
  type?: string
  event?: string
  room?: string
  channel?: string
  data?: unknown
}

type EventHandler = (data: unknown) => void

export class GofreightSocket {
  private url: string
  private ws: WebSocket | null = null
  private handlers = new Map<string, Set<EventHandler>>()
  private opts: Required<Pick<GofreightSocketOptions, 'autoConnect' | 'reconnect' | 'reconnectDelayMs'>> & {
    host?: string
    secure?: boolean
  }
  private _id: string | undefined
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null
  private manualClose = false

  constructor(url: string, options: GofreightSocketOptions = {}) {
    this.url = url
    this.opts = {
      autoConnect: options.autoConnect ?? true,
      reconnect: options.reconnect ?? true,
      reconnectDelayMs: options.reconnectDelayMs ?? 1500,
      host: options.host,
      secure: options.secure,
    }
    if (this.opts.autoConnect) {
      this.connect()
    }
  }

  get id(): string | undefined {
    return this._id
  }

  get connected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN
  }

  connect(): this {
    this.manualClose = false
    this.ws = new WebSocket(this.resolveURL())

    this.ws.onopen = () => {
      this.emitLocal('connect', undefined)
    }

    this.ws.onmessage = (event) => {
      try {
        const msg = JSON.parse(String(event.data)) as WireEnvelope & MessageEventPayload
        this.handleIncoming(msg)
      } catch {
        // ignore malformed frames
      }
    }

    this.ws.onclose = () => {
      this.emitLocal('disconnect', undefined)
      if (!this.manualClose && this.opts.reconnect) {
        this.scheduleReconnect()
      }
    }

    this.ws.onerror = () => {
      this.emitLocal('error', undefined)
    }

    return this
  }

  disconnect(): void {
    this.manualClose = true
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
      this.reconnectTimer = null
    }
    this.ws?.close()
    this.ws = null
  }

  on(event: string, handler: EventHandler): this {
    if (!this.handlers.has(event)) {
      this.handlers.set(event, new Set())
    }
    this.handlers.get(event)!.add(handler)
    return this
  }

  off(event: string, handler?: EventHandler): this {
    if (!handler) {
      this.handlers.delete(event)
      return this
    }
    this.handlers.get(event)?.delete(handler)
    return this
  }

  join(room: string): this {
    this.send({ type: 'join', room })
    return this
  }

  leave(room: string): this {
    this.send({ type: 'leave', room })
    return this
  }

  /** Emit an event to the server (handled by Hub.On). */
  emit(event: string, data?: unknown): this {
    this.send({ type: 'emit', event, data })
    return this
  }

  /** Subscribe to a channel (legacy Action Cable-style). */
  subscribe(channel: string): this {
    this.send({ action: 'subscribe', channel })
    return this
  }

  private resolveURL(): string {
    if (this.url.startsWith('ws://') || this.url.startsWith('wss://')) {
      return this.url
    }

    const path = this.url.startsWith('/') ? this.url : `/${this.url}`

    if (typeof window === 'undefined') {
      throw new Error('GofreightSocket: pass an absolute ws:// or wss:// URL outside the browser')
    }

    const secure = this.opts.secure ?? window.location.protocol === 'https:'
    const protocol = secure ? 'wss:' : 'ws:'
    const host = this.opts.host ?? window.location.host
    return `${protocol}//${host}${path}`
  }

  private send(payload: Record<string, unknown>): void {
    if (!this.connected) {
      return
    }
    this.ws!.send(JSON.stringify(payload))
  }

  private handleIncoming(msg: WireEnvelope & MessageEventPayload): void {
    if (msg.type === 'connected') {
      const data = msg.data as { id?: string } | undefined
      this._id = data?.id
      this.emitLocal('connected', msg.data)
      return
    }

    const event = msg.event
    if (event) {
      this.emitLocal(event, msg.data)
    }
  }

  private emitLocal(event: string, data: unknown): void {
    const set = this.handlers.get(event)
    if (!set) {
      return
    }
    for (const handler of set) {
      handler(data)
    }
  }

  private scheduleReconnect(): void {
    if (this.reconnectTimer) {
      return
    }
    this.reconnectTimer = setTimeout(() => {
      this.reconnectTimer = null
      this.connect()
    }, this.opts.reconnectDelayMs)
  }
}

type MessageEventPayload = {
  channel?: string
  event?: string
  data?: unknown
}

export default GofreightSocket
