<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../../stores/game'
import MahjongTile from './MahjongTile.vue'

const emit = defineEmits<{ discard: [tile: string] }>()

const gameStore = useGameStore()

const sortedHand = computed(() => {
  return [...gameStore.hand].sort()
})

const isMyTurn = computed(() => gameStore.isMyTurn && !gameStore.isReacting)

function handleTileClick(tile: string) {
  if (isMyTurn.value) {
    emit('discard', tile)
  }
}
</script>

<template>
  <div class="player-hand">
    <div class="tiles">
      <MahjongTile
        v-for="(tile, idx) in sortedHand"
        :key="`${tile}-${idx}`"
        :code="tile"
        :is-laizi="tile === gameStore.laiziTile"
        :clickable="isMyTurn"
        @click="handleTileClick(tile)"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.player-hand {
  display: flex;
  justify-content: center;
  padding: $spacing-sm 0;
}

.tiles {
  display: flex;
  gap: 2px;
  flex-wrap: wrap;
  justify-content: center;
}
</style>
