import { useWebSocket } from './useWebSocket'
import { useToast } from './useToast'
import { useUserStore } from '../stores/user'
import { useRoomStore } from '../stores/room'
import { useGameStore } from '../stores/game'
import { usePlayerName } from './usePlayerName'
import { i18n } from '../i18n'
import type { ServerMessage, ClientMessage, BotDifficulty, RoomConfig } from '../types/generated'

let initialized = false
let joinResolve: (() => void) | null = null

export function useGameConnection() {
  const ws = useWebSocket()
  const userStore = useUserStore()
  const roomStore = useRoomStore()
  const gameStore = useGameStore()
  const toast = useToast()
  const { playerName } = usePlayerName()

  // Show a toast when our pending reaction was beaten by a higher-priority
  // action (or by hu ending the round). Caller passes the winning seat and
  // their action; we look up the pending reaction type from the store.
  function notifyPreempted(winningSeat: number, winningAction: string) {
    const pending = gameStore.pendingReaction
    if (!pending) return
    const t = i18n.global.t
    const name = playerName(winningSeat)
    const their = t(`game.actionShort.${winningAction}`, winningAction)
    const yours = t(`game.actionShort.${pending}`, pending)
    toast.show(t('game.preempted', { name, their, yours }), 'warn')
    gameStore.clearPendingReaction()
  }

  function init() {
    if (initialized) return
    initialized = true

    ws.onReconnect(() => {
      const code = roomStore.code
      const token = userStore.sessionToken
      const nickname = userStore.nickname
      if (code && token) {
        send({
          type: 'join_room',
          code,
          nickname,
          session_token: token,
        })
      }
    })

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
          roomStore.updateConfig(msg.config!, (msg as any).players)
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
          // A new reaction round started — any pending reaction from the
          // previous discard is moot now. Clear silently.
          gameStore.clearPendingReaction()
          gameStore.handleReactionPrompt(
            msg.tile!,
            msg.from_seat!,
            msg.available_actions as any,
            msg.chi_options as any,
            msg.time_limit!,
            msg.hu_score_preview,
          )
          break

        case 'action_resolved': {
          const winningSeat = msg.seat!
          if (gameStore.pendingReaction) {
            if (winningSeat === gameStore.yourSeat) {
              gameStore.clearPendingReaction()
            } else {
              notifyPreempted(winningSeat, msg.action!)
            }
          }
          gameStore.handleActionResolved(
            winningSeat, msg.action!, msg.tiles_revealed!, msg.next_turn_seat!,
          )
          break
        }

        case 'gang_result':
          // Handled via action_resolved or game_state
          break

        case 'round_end':
          if (gameStore.pendingReaction && msg.winner_seat != null && msg.winner_seat !== gameStore.yourSeat) {
            notifyPreempted(msg.winner_seat, 'hu')
          } else {
            gameStore.clearPendingReaction()
          }
          gameStore.handleRoundEnd(msg as any)
          break

        case 'game_state':
          roomStore.setPlaying()
          gameStore.handleGameState(msg as any)
          break

        case 'player_disconnected':
          gameStore.setDisconnected(msg.seat!)
          break

        case 'player_reconnected':
          gameStore.setReconnected(msg.seat!)
          break

        case 'bot_added':
          roomStore.addPlayer({
            seat: msg.seat,
            nickname: msg.nickname,
            ready: true,
            connected: true,
            is_bot: true,
            difficulty: msg.difficulty,
          })
          break

        case 'bot_removed':
          roomStore.removePlayer(msg.seat)
          break

        case 'bot_difficulty_changed':
          roomStore.updateBotDifficulty(msg.seat, msg.difficulty)
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
    gameStore.setPendingReaction('pong')
    gameStore.clearReaction()
    send({ type: 'pong' })
  }

  function declareChi(tiles: [string, string]) {
    gameStore.setPendingReaction('chi')
    gameStore.clearReaction()
    send({ type: 'chi', tiles: tiles as any })
  }

  function declareGang(gangType: 'open' | 'closed' | 'add', tile: string) {
    // Only open gang (reaction phase) competes with other players' priorities.
    // Closed/add gang happens on your own turn — no preemption possible.
    if (gangType === 'open') {
      gameStore.setPendingReaction('gang')
    }
    gameStore.clearReaction()
    send({ type: 'gang', gang_type: gangType, tile: tile as any })
  }

  function declareHu() {
    gameStore.setPendingReaction('hu')
    gameStore.clearReaction()
    send({ type: 'hu' })
  }

  function declarePass() {
    gameStore.clearReaction()
    send({ type: 'pass' })
  }

  function addBot(targetSeat: number, difficulty?: BotDifficulty) {
    send({ type: 'add_bot', target_seat: targetSeat, difficulty })
  }

  function removeBot(targetSeat: number) {
    send({ type: 'remove_bot', target_seat: targetSeat })
  }

  function setBotDifficulty(targetSeat: number, difficulty: BotDifficulty) {
    send({ type: 'set_bot_difficulty', target_seat: targetSeat, difficulty })
  }

  function configureRoom(config: RoomConfig) {
    send({ type: 'configure_room', config })
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
    addBot,
    removeBot,
    setBotDifficulty,
    configureRoom,
  }
}
