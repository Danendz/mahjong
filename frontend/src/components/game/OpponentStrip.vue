<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../../stores/game'
import { useRoomStore } from '../../stores/room'
import MahjongTile from './MahjongTile.vue'

const props = defineProps<{
  seat: number
  position: number // 0=right, 1=across, 2=left
}>()

const gameStore = useGameStore()
const roomStore = useRoomStore()

const player = computed(() => roomStore.players.find(p => p.seat === props.seat))
const tileCount = computed(() => gameStore.tileCounts[String(props.seat)] || 0)
const melds = computed(() => gameStore.openMelds[String(props.seat)] || [])
const isCurrentTurn = computed(() => gameStore.currentTurnSeat === props.seat)
const isDisconnected = computed(() => gameStore.disconnectedSeats.has(props.seat))
const isBot = computed(() => player.value?.is_bot === true)
const positionLabel = computed(() => ['Right', 'Across', 'Left'][props.position])
</script>

<template>
  <div class="opponent" :class="{ active: isCurrentTurn, disconnected: isDisconnected }">
    <div class="opponent-info">
      <span class="name">{{ player?.nickname || 'Player' }}</span>
      <span class="position">{{ positionLabel }}</span>
      <span class="tile-count">{{ tileCount }} tiles</span>
      <span v-if="isBot" class="bot-badge">BOT</span>
      <span v-if="isDisconnected" class="dc-badge">DC</span>
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
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.opponent {
  flex: 1;
  background: $color-bg-light;
  border: 1px solid $color-border;
  border-radius: $border-radius;
  padding: $spacing-sm;
  min-width: 120px;
  max-width: 250px;

  &.active {
    border-color: $color-warning;
  }

  &.disconnected {
    opacity: 0.6;
  }
}

.opponent-info {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  margin-bottom: $spacing-xs;
  font-size: 0.8rem;
}

.name {
  font-weight: 600;
  font-size: 0.9rem;
}

.position {
  color: $color-text-muted;
}

.tile-count {
  color: $color-text-muted;
  margin-left: auto;
}

.bot-badge {
  background: #2dd4bf;
  color: #1a1a2e;
  font-size: 0.6rem;
  font-weight: 700;
  padding: 1px 4px;
  border-radius: 2px;
}

.dc-badge {
  background: $color-danger;
  color: white;
  font-size: 0.6rem;
  padding: 1px 4px;
  border-radius: 2px;
}

.melds {
  display: flex;
  gap: $spacing-sm;
  flex-wrap: wrap;
}

.meld {
  display: flex;
  gap: 1px;
}
</style>
