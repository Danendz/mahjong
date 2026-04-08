<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../../stores/game'

const gameStore = useGameStore()
const TOTAL_TILES = 136 - 14 * 4 // 136 tiles minus 4 hands of ~14 = ~80 drawable tiles

const percentage = computed(() =>
  Math.max(0, Math.min(100, (gameStore.wallRemaining / TOTAL_TILES) * 100))
)

const isLow = computed(() => gameStore.wallRemaining <= 15)
</script>

<template>
  <div class="wall-bar">
    <div class="bar-track">
      <div
        class="bar-fill"
        :class="{ low: isLow }"
        :style="{ width: percentage + '%' }"
      />
      <span class="bar-label">{{ gameStore.wallRemaining }}</span>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.wall-bar {
  width: 100%;
  max-width: 200px;
}

.bar-track {
  position: relative;
  height: 24px;
  background: rgba(white, 0.08);
  border-radius: 12px;
  overflow: hidden;
}

.bar-fill {
  height: 100%;
  background: $color-surface;
  border-radius: 12px;
  transition: width 0.5s ease;

  &.low {
    background: rgba($color-danger, 0.4);
  }
}

.bar-label {
  position: absolute;
  inset: 0;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 0.75rem;
  font-weight: 700;
  color: $color-text;
  font-variant-numeric: tabular-nums;
}
</style>
