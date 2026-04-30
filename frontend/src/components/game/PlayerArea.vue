<script setup lang="ts">
import { computed, TransitionGroup } from 'vue'
import { useGameStore } from '../../stores/game'
import { useRoomStore } from '../../stores/room'
import { usePlayerName } from '../../composables/usePlayerName'
import MahjongTile from './MahjongTile.vue'
import MeldCard from './MeldCard.vue'

const props = defineProps<{
  seat: number
  position: 'across' | 'left' | 'right'
}>()

const gameStore = useGameStore()
const roomStore = useRoomStore()
const { playerName } = usePlayerName()

const player = computed(() => roomStore.players.find(p => p.seat === props.seat))
const tileCount = computed(() => gameStore.tileCounts[String(props.seat)] || 0)
const melds = computed(() => gameStore.openMelds[String(props.seat)] || [])
const discardTiles = computed(() => gameStore.discards[String(props.seat)] || [])
const isCurrentTurn = computed(() => gameStore.currentTurnSeat === props.seat)
const isDisconnected = computed(() => gameStore.disconnectedSeats.has(props.seat))
const isBot = computed(() => player.value?.is_bot === true)

// Contested-tile highlight: when this seat is the one whose discard is being reacted to,
// pick a ring color based on the highest-priority available action.
const contestedRingColor = computed<string | null>(() => {
  if (!gameStore.isReacting) return null
  if (gameStore.reactionFromSeat !== props.seat) return null
  const actions = gameStore.availableActions
  if (actions.includes('hu')) return '#e94560'
  if (actions.includes('gang')) return '#f0a500'
  if (actions.includes('pong')) return '#4ecca3'
  if (actions.includes('chi')) return '#5b8def'
  return null
})

const contestedIndex = computed<number>(() => {
  if (!contestedRingColor.value || !gameStore.reactionTile) return -1
  // Last tile in discard pile — that's the one just discarded.
  const tiles = discardTiles.value
  if (tiles.length === 0) return -1
  return tiles[tiles.length - 1] === gameStore.reactionTile ? tiles.length - 1 : -1
})
</script>

<template>
  <div
    class="player-area"
    :class="[position, { active: isCurrentTurn, disconnected: isDisconnected }]"
  >
    <div class="player-header">
      <span class="name">{{ playerName(seat) }}</span>
      <span class="tile-count">{{ tileCount }}</span>
      <span v-if="isBot" class="badge bot-badge">BOT</span>
      <span v-if="isDisconnected" class="badge dc-badge">DC</span>
    </div>

    <TransitionGroup v-if="melds.length" name="meld" tag="div" class="melds">
      <MeldCard
        v-for="(meld, idx) in melds"
        :key="idx"
        :type="meld.type"
        :tiles="meld.tiles"
        :laizi-tile="gameStore.laiziTile"
        :show-label="false"
        small
      />
    </TransitionGroup>

    <TransitionGroup v-if="discardTiles.length" name="discard" tag="div" class="discards">
      <MahjongTile
        v-for="(tile, idx) in discardTiles"
        :key="`${tile}-${idx}`"
        :code="tile"
        :is-laizi="tile === gameStore.laiziTile"
        :contested="idx === contestedIndex"
        :contested-color="idx === contestedIndex ? contestedRingColor || undefined : undefined"
        small
      />
    </TransitionGroup>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.player-area {
  display: flex;
  flex-direction: column;
  gap: $spacing-xs;
  padding: $spacing-sm;
  background: $color-bg-light;
  border: 1px solid $color-border;
  border-radius: $border-radius;
  overflow: hidden;

  &.active {
    border-color: $color-warning;
    box-shadow: 0 0 12px rgba($color-warning, 0.3);
    animation: turnGlow 2s ease-in-out infinite;
  }

  &.disconnected {
    opacity: 0.6;
  }
}

.player-header {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  font-size: 0.8rem;
}

.name {
  font-weight: 600;
  font-size: 0.85rem;
}

.tile-count {
  color: $color-text-muted;
  font-size: 0.75rem;
  margin-left: auto;

  &::after {
    content: ' tiles';
  }
}

.badge {
  font-size: 0.6rem;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 2px;
}

.bot-badge {
  background: #2dd4bf;
  color: #1a1a2e;
}

.dc-badge {
  background: $color-danger;
  color: white;
}

.melds {
  display: flex;
  gap: $spacing-xs;
  flex-wrap: wrap;
  padding-top: 4px;
}

.discards {
  display: flex;
  gap: 1px;
  flex-wrap: wrap;
  padding-top: $spacing-xs;
  border-top: 1px solid rgba(white, 0.05);
}

// Position-specific styles
.across {
  align-items: center;

  .discards {
    justify-content: center;
    max-width: 300px;
  }

  .melds {
    justify-content: center;
  }
}

.left {
  .discards {
    max-width: 160px;
  }
}

.right {
  align-items: flex-end;

  .player-header {
    flex-direction: row-reverse;

    .tile-count {
      margin-left: 0;
      margin-right: auto;
    }
  }

  .melds {
    justify-content: flex-end;
  }

  .discards {
    justify-content: flex-end;
    max-width: 160px;
  }
}

// Discard tile animation
.discard-enter-active {
  transition: all 0.3s ease-out;
}

.discard-enter-from {
  opacity: 0;
  transform: translateY(-8px);
}

// Meld claim pulse
.meld-enter-active {
  transition: all 0.3s ease-out;
  animation: meldPulseSmall 0.6s ease-out 0.3s 1;
}

.meld-enter-from {
  opacity: 0;
  transform: scale(0.8);
}

@keyframes meldPulseSmall {
  0%, 100% { filter: none; }
  50% { filter: brightness(1.3) saturate(1.2); }
}

@keyframes turnGlow {
  0%, 100% { box-shadow: 0 0 12px rgba($color-warning, 0.3); }
  50% { box-shadow: 0 0 20px rgba($color-warning, 0.5); }
}
</style>
