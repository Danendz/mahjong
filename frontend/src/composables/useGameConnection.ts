import { useWebSocket } from './useWebSocket'
import { useUserStore } from '../stores/user'
import { useRoomStore } from '../stores/room'
import { useGameStore } from '../stores/game'
import type { ServerMessage, ClientMessage } from '../types/generated'

let initialized = false
let joinResolve: (() => void) | null = null

export function useGameConnection() {
  const ws = useWebSocket()
  const userStore = useUserStore()
  const roomStore = useRoomStore()
  const gameStore = useGameStore()

  function init() {
    if (initialized) return
    initialized = true

    ws.onMessage((msg: ServerMessage) => {
      switch (msg.type) {
        case 'room_joined':
          userStore.setSeat(msg.your_seat!)
          gameStore.yourSeat = msg.your_seat!
          roomStore.setRoom(msg.code!, msg.room_id!, msg.config!, msg.players!)
          if (joinResolve) {
            joinResolve()
            joinResolve = null
          }
          break

        case 'player_joined':
          roomStore.addPlayer({
            seat: msg.seat!,
            nickname: msg.nickname!,
            ready: false,
            connected: true,
          })
          break

        case 'player_left':
          roomStore.removePlayer(msg.seat!)
          break

        case 'player_ready':
          roomStore.setPlayerReady(msg.seat!)
          break

        case 'config_updated':
          roomStore.updateConfig(msg.config!)
          break

        case 'game_started':
          roomStore.setPlaying()
          gameStore.handleGameStarted(msg as any)
          break

        case 'your_turn':
          gameStore.handleYourTurn(msg as any)
          break

        case 'tile_discarded':
          gameStore.handleTileDiscarded(msg.seat!, msg.tile!, msg.wall_remaining!)
          break

        case 'reaction_prompt':
          gameStore.handleReactionPrompt(
            msg.tile!,
            msg.from_seat!,
            msg.available_actions as any,
            msg.chi_options as any,
            msg.time_limit!,
          )
          break

        case 'action_resolved':
          gameStore.handleActionResolved(
            msg.seat!, msg.action!, msg.tiles_revealed!, msg.next_turn_seat!,
          )
          break

        case 'gang_result':
          // Handled via action_resolved or game_state
          break

        case 'round_end':
          gameStore.handleRoundEnd(msg as any)
          break

        case 'game_state':
          gameStore.handleGameState(msg as any)
          break

        case 'player_disconnected':
          gameStore.setDisconnected(msg.seat!)
          break

        case 'player_reconnected':
          gameStore.setReconnected(msg.seat!)
          break

        case 'error':
          console.error(`Server error [${msg.code}]: ${msg.message}`)
          break
      }
    })

    ws.connect()
  }

  function send(msg: ClientMessage) {
    ws.send(msg)
  }

  function joinRoom(code: string, nickname: string, sessionToken: string): Promise<void> {
    return new Promise((resolve) => {
      joinResolve = resolve
      send({
        type: 'join_room',
        code,
        nickname,
        session_token: sessionToken,
      })
    })
  }

  function leaveRoom() {
    send({ type: 'leave_room' })
  }

  function toggleReady() {
    send({ type: 'player_ready' })
  }

  function startGame() {
    send({ type: 'start_game' })
  }

  function discard(tile: string) {
    gameStore.discardTile(tile as any)
    send({ type: 'discard', tile: tile as any })
  }

  function declarePong() {
    gameStore.clearReaction()
    send({ type: 'pong' })
  }

  function declareChi(tiles: [string, string]) {
    gameStore.clearReaction()
    send({ type: 'chi', tiles: tiles as any })
  }

  function declareGang(gangType: 'open' | 'closed' | 'add', tile: string) {
    gameStore.clearReaction()
    send({ type: 'gang', gang_type: gangType, tile: tile as any })
  }

  function declareHu() {
    gameStore.clearReaction()
    send({ type: 'hu' })
  }

  function declarePass() {
    gameStore.clearReaction()
    send({ type: 'pass' })
  }

  return {
    status: ws.status,
    init,
    send,
    joinRoom,
    leaveRoom,
    toggleReady,
    startGame,
    discard,
    declarePong,
    declareChi,
    declareGang,
    declareHu,
    declarePass,
  }
}
