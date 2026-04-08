<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../../stores/game'
import MahjongTile from './MahjongTile.vue'

const gameStore = useGameStore()

const allDiscards = computed(() => {
  const result: { seat: number; tiles: string[] }[] = []
  for (let i = 0; i < 4; i++) {
    const key = String(i)
    result.push({
      seat: i,
      tiles: gameStore.discards[key] || [],
    })
  }
  return result
})
</script>

<template>
  <div class="discard-pool">
    <div
      v-for="pd in allDiscards"
      :key="pd.seat"
      class="player-discards"
    >
      <MahjongTile
        v-for="(tile, idx) in pd.tiles"
        :key="`${pd.seat}-${idx}`"
        :code="tile"
        :is-laizi="tile === gameStore.laiziTile"
        small
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.discard-pool {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: $spacing-sm;
  max-width: 400px;
  width: 100%;
}

.player-discards {
  display: flex;
  flex-wrap: wrap;
  gap: 2px;
  padding: $spacing-xs;
  min-height: 40px;
  background: rgba(white, 0.03);
  border-radius: $border-radius-sm;
}
</style>
