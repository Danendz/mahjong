<script setup lang="ts">
import { useGameStore } from '../../stores/game'

const gameStore = useGameStore()
</script>

<template>
  <div class="game-info">
    <span>Wall: {{ gameStore.wallRemaining }}</span>
    <span v-for="(score, seat) in gameStore.scores" :key="seat">
      P{{ seat }}: {{ score }}
    </span>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.game-info {
  display: flex;
  gap: $spacing-md;
  font-size: 0.8rem;
  color: $color-text-muted;

  span {
    background: rgba(white, 0.05);
    padding: $spacing-xs $spacing-sm;
    border-radius: $border-radius-sm;
  }
}
</style>
