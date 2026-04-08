<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../../stores/game'
import { usePlayerName } from '../../composables/usePlayerName'

const gameStore = useGameStore()
const { playerName } = usePlayerName()

const statusText = computed(() => {
  if (gameStore.isReacting) {
    const name = playerName(gameStore.reactionFromSeat)
    return `React to ${name}'s discard`
  }
  if (gameStore.isMyTurn) {
    return 'Your turn \u2014 discard a tile'
  }
  if (gameStore.currentTurnSeat >= 0) {
    const name = playerName(gameStore.currentTurnSeat)
    return `${name}'s turn`
  }
  return ''
})

const isUrgent = computed(() => gameStore.isMyTurn || gameStore.isReacting)
</script>

<template>
  <div v-if="statusText" class="status-banner" :class="{ urgent: isUrgent }">
    {{ statusText }}
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.status-banner {
  text-align: center;
  font-size: 0.9rem;
  font-weight: 600;
  color: $color-text-muted;
  background: rgba(white, 0.05);
  padding: $spacing-xs $spacing-md;
  border-radius: $border-radius;
  transition: color 0.3s, background 0.3s;

  &.urgent {
    color: $color-warning;
    background: rgba($color-warning, 0.1);
  }
}
</style>
