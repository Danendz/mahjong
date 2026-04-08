<script setup lang="ts">
import { computed, TransitionGroup } from 'vue'
import { useGameStore } from '../../stores/game'
import { useRoomStore } from '../../stores/room'
import { usePlayerName } from '../../composables/usePlayerName'
import MahjongTile from './MahjongTile.vue'

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

    <div v-if="melds.length" class="melds">
      <div v-for="(meld, idx) in melds" :key="idx" class="meld">
        <MahjongTile
          v-for="(tile, tidx) in meld.tiles"
          :key="tidx"
          :code="tile"
          :is-laizi="tile === gameStore.laiziTile"
          :face-down="meld.type === 'closed_gang'"
          small
        />
      </div>
    </div>

    <TransitionGroup v-if="discardTiles.length" name="discard" tag="div" class="discards">
      <MahjongTile
        v-for="(tile, idx) in discardTiles"
        :key="`${tile}-${idx}`"
        :code="tile"
        :is-laizi="tile === gameStore.laiziTile"
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
}

.meld {
  display: flex;
  gap: 1px;
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

@keyframes turnGlow {
  0%, 100% { box-shadow: 0 0 12px rgba($color-warning, 0.3); }
  50% { box-shadow: 0 0 20px rgba($color-warning, 0.5); }
}
</style>
