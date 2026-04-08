import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import type {
  TileCode, MeldInfo, ReactionAction,
  GameStateMsg, GameStartedMsg, YourTurnMsg, RoundEndMsg,
} from '../types/generated'

export const useGameStore = defineStore('game', () => {
  // Player's hand
  const hand = ref<TileCode[]>([])
  const drawnTile = ref<TileCode | null>(null)

  // Game info
  const laiziIndicator = ref<TileCode | null>(null)
  const laiziTile = ref<TileCode | null>(null)
  const dealerSeat = ref(0)
  const currentTurnSeat = ref(0)
  const wallRemaining = ref(0)
  const yourSeat = ref(0)

  // All players' visible state
  const openMelds = ref<Record<string, MeldInfo[]>>({})
  const discards = ref<Record<string, TileCode[]>>({})
  const tileCounts = ref<Record<string, number>>({})
  const scores = ref<Record<string, number>>({})

  // Turn actions
  const canGang = ref<TileCode[]>([])
  const canHu = ref(false)
  const timeLimit = ref(0)
  const huScorePreview = ref<number | null>(null)
  const waitingTiles = ref<TileCode[]>([])

  // Reaction state
  const reactionTile = ref<TileCode | null>(null)
  const reactionFromSeat = ref(-1)
  const availableActions = ref<ReactionAction[]>([])
  const chiOptions = ref<[TileCode, TileCode][]>([])
  const reactionTimeLimit = ref(0)

  // Round end
  const roundResult = ref<RoundEndMsg | null>(null)
  const isRoundEnd = ref(false)

  // Timer trigger — incremented each turn/reaction so watchers always fire
  const turnVersion = ref(0)

  // Connection status per player
  const disconnectedSeats = ref<Set<number>>(new Set())

  const isMyTurn = computed(() => currentTurnSeat.value === yourSeat.value)
  const isReacting = computed(() => availableActions.value.length > 0)

  function handleGameStarted(msg: GameStartedMsg) {
    hand.value = msg.your_hand || []
    laiziIndicator.value = msg.laizi_indicator
    laiziTile.value = msg.laizi_tile
    dealerSeat.value = msg.dealer_seat
    wallRemaining.value = msg.wall_remaining
    drawnTile.value = null
    isRoundEnd.value = false
    roundResult.value = null
    availableActions.value = []

    // Reset discards, melds, and tile counts for new round
    openMelds.value = { '0': [], '1': [], '2': [], '3': [] }
    discards.value = { '0': [], '1': [], '2': [], '3': [] }
    tileCounts.value = {}
    for (let i = 0; i < 4; i++) {
      tileCounts.value[String(i)] = i === msg.dealer_seat ? 14 : 13
    }
    canGang.value = []
    canHu.value = false
    turnVersion.value = 0
  }

  function handleYourTurn(msg: YourTurnMsg) {
    drawnTile.value = msg.drawn_tile ?? null
    if (msg.drawn_tile) {
      hand.value.push(msg.drawn_tile)
      // Increment own tile count for the draw
      const key = String(yourSeat.value)
      tileCounts.value[key] = (tileCounts.value[key] || 13) + 1
    }
    canGang.value = msg.can_gang || []
    canHu.value = msg.can_hu ?? false
    huScorePreview.value = msg.hu_score_preview ?? null
    waitingTiles.value = msg.waiting_tiles || []
    timeLimit.value = msg.time_limit
    wallRemaining.value = msg.wall_remaining
    currentTurnSeat.value = yourSeat.value
    availableActions.value = []
    turnVersion.value++
  }

  function handleTileDiscarded(seat: number, tile: TileCode, remaining: number) {
    wallRemaining.value = remaining
    const key = String(seat)
    if (!discards.value[key]) discards.value[key] = []
    discards.value[key].push(tile)

    // Update tile count for the discarding player
    if (tileCounts.value[key] !== undefined) {
      tileCounts.value[key]--
    }

    currentTurnSeat.value = -1 // Waiting for reactions or next turn
  }

  function handleReactionPrompt(
    tile: TileCode,
    fromSeat: number,
    actions: ReactionAction[],
    options: [TileCode, TileCode][] | undefined,
    limit: number,
    scorePreview?: number,
  ) {
    reactionTile.value = tile
    reactionFromSeat.value = fromSeat
    availableActions.value = actions
    chiOptions.value = options || []
    reactionTimeLimit.value = limit
    huScorePreview.value = scorePreview ?? null
    turnVersion.value++
  }

  function handleActionResolved(seat: number, action: string, tilesRevealed: TileCode[], nextTurnSeat: number) {
    const key = String(seat)
    if (!openMelds.value[key]) openMelds.value[key] = []

    if (action === 'pong' || action === 'chi' || action === 'gang') {
      openMelds.value[key].push({
        type: action === 'gang' ? 'open_gang' : action as MeldInfo['type'],
        tiles: tilesRevealed,
      })
    }

    currentTurnSeat.value = nextTurnSeat
    availableActions.value = []
    reactionTile.value = null
  }

  function handleRoundEnd(msg: RoundEndMsg) {
    roundResult.value = msg
    isRoundEnd.value = true
    scores.value = msg.total_scores
    availableActions.value = []
  }

  function handleGameState(msg: GameStateMsg) {
    yourSeat.value = msg.your_seat
    hand.value = msg.your_hand
    openMelds.value = msg.open_melds
    discards.value = msg.discards
    tileCounts.value = msg.tile_counts
    currentTurnSeat.value = msg.current_turn_seat
    laiziIndicator.value = msg.laizi_indicator
    laiziTile.value = msg.laizi_tile
    wallRemaining.value = msg.wall_remaining
    scores.value = msg.total_scores
    dealerSeat.value = msg.dealer_seat
  }

  function discardTile(tile: TileCode) {
    const idx = hand.value.indexOf(tile)
    if (idx >= 0) {
      hand.value.splice(idx, 1)
    }
    drawnTile.value = null
    canGang.value = []
    canHu.value = false
  }

  function clearReaction() {
    availableActions.value = []
    reactionTile.value = null
    chiOptions.value = []
  }

  function setDisconnected(seat: number) {
    disconnectedSeats.value.add(seat)
  }

  function setReconnected(seat: number) {
    disconnectedSeats.value.delete(seat)
  }

  function $reset() {
    hand.value = []
    drawnTile.value = null
    laiziIndicator.value = null
    laiziTile.value = null
    openMelds.value = {}
    discards.value = {}
    tileCounts.value = {}
    scores.value = {}
    availableActions.value = []
    roundResult.value = null
    isRoundEnd.value = false
    disconnectedSeats.value.clear()
    turnVersion.value = 0
  }

  return {
    hand, drawnTile, laiziIndicator, laiziTile, dealerSeat,
    currentTurnSeat, wallRemaining, yourSeat, turnVersion,
    openMelds, discards, tileCounts, scores,
    canGang, canHu, timeLimit, huScorePreview, waitingTiles,
    reactionTile, reactionFromSeat, availableActions, chiOptions, reactionTimeLimit,
    roundResult, isRoundEnd, disconnectedSeats,
    isMyTurn, isReacting,
    handleGameStarted, handleYourTurn, handleTileDiscarded,
    handleReactionPrompt, handleActionResolved, handleRoundEnd,
    handleGameState, discardTile, clearReaction,
    setDisconnected, setReconnected, $reset,
  }
})
