<script setup lang="ts">
import { useGameStore } from '../../stores/game'
import MahjongTile from './MahjongTile.vue'

const gameStore = useGameStore()
</script>

<template>
  <div v-if="gameStore.laiziIndicator" class="laizi-indicator">
    <div class="label">Laizi</div>
    <div class="tiles">
      <div class="indicator">
        <MahjongTile :code="gameStore.laiziIndicator" small />
        <span class="arrow">&rarr;</span>
        <MahjongTile :code="gameStore.laiziTile!" is-laizi small />
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.laizi-indicator {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  background: rgba(white, 0.05);
  padding: $spacing-xs $spacing-md;
  border-radius: $border-radius;
}

.label {
  font-size: 0.75rem;
  color: $color-text-muted;
  text-transform: uppercase;
}

.indicator {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
}

.arrow {
  color: $color-text-muted;
  font-size: 1.2rem;
}
</style>
