import { onMounted, onUnmounted } from 'vue'
import { wsClient } from '@/utils/websocket'
import type { Message, Session } from '@/types'

export interface WebSocketEventHandlers {
  onMessageReceived?: (message: Message) => void
  onMessageDelivered?: (messageId: string) => void
  onSessionConnect?: (session: Session) => void
  onSessionDisconnect?: (sessionId: string) => void
}

/**
 * Composable for managing WebSocket event subscriptions
 * Automatically handles cleanup on component unmount
 */
export function useWebSocketEvents(handlers: WebSocketEventHandlers) {
  const wrappedHandlers = {
    message_received: handlers.onMessageReceived
      ? (data: { message: Message }) => handlers.onMessageReceived!(data.message)
      : undefined,
    message_delivered: handlers.onMessageDelivered
      ? (data: { message_id: string }) => handlers.onMessageDelivered!(data.message_id)
      : undefined,
    session_connect: handlers.onSessionConnect
      ? (data: { session: Session }) => handlers.onSessionConnect!(data.session)
      : undefined,
    session_disconnect: handlers.onSessionDisconnect
      ? (data: { session_id: string }) => handlers.onSessionDisconnect!(data.session_id)
      : undefined
  }

  onMounted(() => {
    if (wrappedHandlers.message_received) {
      wsClient.on('message_received', wrappedHandlers.message_received)
    }
    if (wrappedHandlers.message_delivered) {
      wsClient.on('message_delivered', wrappedHandlers.message_delivered)
    }
    if (wrappedHandlers.session_connect) {
      wsClient.on('session_connect', wrappedHandlers.session_connect)
    }
    if (wrappedHandlers.session_disconnect) {
      wsClient.on('session_disconnect', wrappedHandlers.session_disconnect)
    }
  })

  onUnmounted(() => {
    if (wrappedHandlers.message_received) {
      wsClient.off('message_received', wrappedHandlers.message_received)
    }
    if (wrappedHandlers.message_delivered) {
      wsClient.off('message_delivered', wrappedHandlers.message_delivered)
    }
    if (wrappedHandlers.session_connect) {
      wsClient.off('session_connect', wrappedHandlers.session_connect)
    }
    if (wrappedHandlers.session_disconnect) {
      wsClient.off('session_disconnect', wrappedHandlers.session_disconnect)
    }
  })
}
