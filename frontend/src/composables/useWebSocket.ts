import { ref, readonly } from 'vue'
import type { ClientMessage, ServerMessage } from '../types/generated'

const WS_URL = import.meta.env.VITE_WS_URL || 'ws://localhost:8080/ws'

export type ConnectionStatus = 'disconnected' | 'connecting' | 'connected' | 'reconnecting'

const status = ref<ConnectionStatus>('disconnected')
let socket: WebSocket | null = null
let reconnectAttempts = 0
let reconnectTimer: ReturnType<typeof setTimeout> | null = null
let messageHandler: ((msg: ServerMessage) => void) | null = null

const MAX_RECONNECT_DELAY = 30_000

function getReconnectDelay(): number {
  // Exponential backoff: 0, 1s, 2s, 4s, 8s, ... capped at 30s
  if (reconnectAttempts === 0) return 0
  return Math.min(1000 * Math.pow(2, reconnectAttempts - 1), MAX_RECONNECT_DELAY)
}

export function useWebSocket() {
  function connect() {
    if (socket?.readyState === WebSocket.OPEN) return

    status.value = reconnectAttempts > 0 ? 'reconnecting' : 'connecting'

    socket = new WebSocket(WS_URL)

    socket.onopen = () => {
      status.value = 'connected'
      reconnectAttempts = 0
    }

    socket.onmessage = (event) => {
      try {
        const msg: ServerMessage = JSON.parse(event.data)
        if (messageHandler) {
          messageHandler(msg)
        }
      } catch (e) {
        console.error('Failed to parse message:', e)
      }
    }

    socket.onclose = () => {
      status.value = 'disconnected'
      socket = null
      scheduleReconnect()
    }

    socket.onerror = () => {
      socket?.close()
    }
  }

  function disconnect() {
    if (reconnectTimer) {
      clearTimeout(reconnectTimer)
      reconnectTimer = null
    }
    reconnectAttempts = 0
    socket?.close()
    socket = null
    status.value = 'disconnected'
  }

  function send(msg: ClientMessage) {
    if (socket?.readyState !== WebSocket.OPEN) {
      console.warn('WebSocket not connected, cannot send:', msg.type)
      return
    }
    socket.send(JSON.stringify(msg))
  }

  function onMessage(handler: (msg: ServerMessage) => void) {
    messageHandler = handler
  }

  function scheduleReconnect() {
    reconnectAttempts++
    const delay = getReconnectDelay()
    reconnectTimer = setTimeout(() => {
      connect()
    }, delay)
  }

  return {
    status: readonly(status),
    connect,
    disconnect,
    send,
    onMessage,
  }
}
