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
        <div class="tile-with-label">
          <MahjongTile :code="gameStore.laiziIndicator" small />
          <span class="sub-label">翻</span>
        </div>
        <span class="arrow">&rarr;</span>
        <div class="tile-with-label">
          <MahjongTile :code="gameStore.laiziTile!" is-laizi small />
          <span class="sub-label wild">癞子</span>
        </div>
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

.tile-with-label {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.sub-label {
  font-size: 0.6rem;
  color: $color-text-muted;

  &.wild {
    color: $color-tile-laizi;
    font-weight: 600;
  }
}
</style>
