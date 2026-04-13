type MessageHandler = (data: any) => void

class WebSocketClient {
  private ws: WebSocket | null = null
  private _url: string
  private handlers: Map<string, MessageHandler[]> = new Map()
  private reconnectTimer: number | null = null
  private reconnectAttempts = 0
  private maxReconnectAttempts = 5
  private heartbeatTimer: number | null = null
  private heartbeatInterval = 30000 // 30 seconds

  constructor(url: string) {
    this._url = url
  }

  connect() {
    // 防止重复连接：OPEN 或 CONNECTING 状态都跳过
    if (this.ws?.readyState === WebSocket.OPEN || this.ws?.readyState === WebSocket.CONNECTING) {
      return
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:'
    const token = localStorage.getItem('token')

    // Build WebSocket URL with token for authentication
    let wsUrl = `${protocol}//${window.location.host}${this._url}`
    if (token) {
      wsUrl += `?token=${encodeURIComponent(token)}`
    }

    this.ws = new WebSocket(wsUrl)

    this.ws.onopen = () => {
      console.log('WebSocket connected')
      this.reconnectAttempts = 0
      this.startHeartbeat()
    }

    this.ws.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        const handlers = this.handlers.get(data.type) || []
        handlers.forEach(handler => handler(data))

        // Also trigger 'any' handlers
        const anyHandlers = this.handlers.get('*') || []
        anyHandlers.forEach(handler => handler(data))
      } catch (e) {
        console.error('Failed to parse WebSocket message:', e)
      }
    }

    this.ws.onclose = (event) => {
      console.log('WebSocket disconnected', event.code, event.reason)
      this.stopHeartbeat()

      // Don't reconnect if authentication failed (401)
      if (event.code === 1008) {
        console.error('WebSocket authentication failed')
        return
      }

      this.scheduleReconnect()
    }

    this.ws.onerror = (error) => {
      console.error('WebSocket error:', error)
    }
  }

  private startHeartbeat() {
    this.stopHeartbeat()
    this.heartbeatTimer = window.setInterval(() => {
      if (this.ws?.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify({ type: 'ping' }))
      }
    }, this.heartbeatInterval)
  }

  private stopHeartbeat() {
    if (this.heartbeatTimer) {
      clearInterval(this.heartbeatTimer)
      this.heartbeatTimer = null
    }
  }

  private scheduleReconnect() {
    if (this.reconnectAttempts >= this.maxReconnectAttempts) {
      return
    }

    this.reconnectAttempts++
    const delay = Math.min(1000 * Math.pow(2, this.reconnectAttempts), 30000)

    this.reconnectTimer = window.setTimeout(() => {
      this.connect()
    }, delay)
  }

  on(type: string, handler: MessageHandler) {
    if (!this.handlers.has(type)) {
      this.handlers.set(type, [])
    }
    this.handlers.get(type)!.push(handler)
  }

  off(type: string, handler: MessageHandler) {
    const handlers = this.handlers.get(type)
    if (handlers) {
      const index = handlers.indexOf(handler)
      if (index > -1) {
        handlers.splice(index, 1)
      }
    }
  }

  removeAllHandlers() {
    this.handlers.clear()
  }

  disconnect() {
    this.stopHeartbeat()
    if (this.reconnectTimer) {
      clearTimeout(this.reconnectTimer)
    }
    if (this.ws) {
      this.ws.close()
      this.ws = null
    }
  }
}

export const wsClient = new WebSocketClient('/ws')
