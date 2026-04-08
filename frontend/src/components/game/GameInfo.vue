<script setup lang="ts">
import { useGameStore } from '../../stores/game'
import { usePlayerName } from '../../composables/usePlayerName'

const gameStore = useGameStore()
const { playerName } = usePlayerName()
</script>

<template>
  <div class="game-info">
    <span
      v-for="(score, seat) in gameStore.scores"
      :key="seat"
      class="score"
      :class="{ current: Number(seat) === gameStore.currentTurnSeat }"
    >
      {{ playerName(Number(seat)) }}: <strong>{{ score }}</strong>
    </span>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.game-info {
  display: flex;
  gap: $spacing-sm;
  font-size: 0.85rem;
  color: $color-text-muted;
  flex-wrap: wrap;
  justify-content: center;

  span {
    background: rgba(white, 0.05);
    padding: $spacing-xs $spacing-sm;
    border-radius: $border-radius-sm;
  }

  .current {
    color: $color-warning;
    font-weight: 600;
  }
}
</style>
