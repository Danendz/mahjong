import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type { PlayerInfo, RoomConfig, BotDifficulty } from '../types/generated'

export const useRoomStore = defineStore('room', () => {
  const roomId = ref('')
  const code = ref('')
  const players = ref<PlayerInfo[]>([])
  const config = ref<RoomConfig>({
    score_cap: 500,
    open_call_mode: 'koukou',
    turn_timer: 15,
    reaction_timer: 8,
    num_rounds: 8,
    zimo_only: false,
    dealer_continuation: false,
  })
  const status = ref<'lobby' | 'waiting' | 'playing' | 'finished'>('lobby')

  const playerCount = computed(() => players.value.length)
  const allReady = computed(() =>
    players.value.length === 4 && players.value.every(p => p.ready)
  )

  function setRoom(roomCode: string, id: string, cfg: RoomConfig, playerList: PlayerInfo[]) {
    code.value = roomCode
    roomId.value = id
    config.value = cfg
    players.value = playerList
    status.value = 'waiting'
  }

  function addPlayer(player: PlayerInfo) {
    const existing = players.value.findIndex(p => p.seat === player.seat)
    if (existing >= 0) {
      players.value[existing] = player
    } else {
      players.value.push(player)
    }
  }

  function removePlayer(seat: number) {
    players.value = players.value.filter(p => p.seat !== seat)
  }

  function setPlayerReady(seat: number) {
    const player = players.value.find(p => p.seat === seat)
    if (player) {
      player.ready = !player.ready
    }
  }

  function updateConfig(cfg: RoomConfig, playerList?: PlayerInfo[]) {
    config.value = cfg
    if (playerList) {
      players.value = playerList
    }
  }

  function updateBotDifficulty(seat: number, difficulty: BotDifficulty) {
    const player = players.value.find(p => p.seat === seat)
    if (player) {
      player.difficulty = difficulty
    }
  }

  function setPlaying() {
    status.value = 'playing'
  }

  function $reset() {
    roomId.value = ''
    code.value = ''
    players.value = []
    status.value = 'lobby'
  }

  return {
    roomId, code, players, config, status,
    playerCount, allReady,
    setRoom, addPlayer, removePlayer, setPlayerReady, updateConfig, updateBotDifficulty, setPlaying, $reset,
  }
})
