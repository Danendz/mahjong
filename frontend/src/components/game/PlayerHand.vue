<script setup lang="ts">
import { computed, TransitionGroup } from 'vue'
import { useGameStore } from '../../stores/game'
import MahjongTile from './MahjongTile.vue'

const emit = defineEmits<{ discard: [tile: string] }>()

const gameStore = useGameStore()

function suitOrder(tile: string): number {
  if (tile.endsWith('m')) return 0
  if (tile.endsWith('p')) return 1
  if (tile.endsWith('s')) return 2
  if (tile.startsWith('w')) return 3
  if (tile.startsWith('d')) return 4
  return 5
}

function tileSortComparator(a: string, b: string): number {
  const suitDiff = suitOrder(a) - suitOrder(b)
  if (suitDiff !== 0) return suitDiff
  return a.localeCompare(b)
}

const sortedHand = computed(() => {
  const tiles = [...gameStore.hand]
  // Remove the drawn tile from the sorted hand — it's displayed separately
  if (gameStore.drawnTile) {
    const idx = tiles.lastIndexOf(gameStore.drawnTile)
    if (idx >= 0) tiles.splice(idx, 1)
  }
  return tiles.sort(tileSortComparator)
})

const groupedHand = computed(() => {
  const sorted = sortedHand.value
  if (sorted.length === 0) return []
  const groups: string[][] = []
  let currentGroup: string[] = [sorted[0]]
  let currentSuit = suitOrder(sorted[0])
  for (let i = 1; i < sorted.length; i++) {
    const suit = suitOrder(sorted[i])
    if (suit !== currentSuit) {
      groups.push(currentGroup)
      currentGroup = []
      currentSuit = suit
    }
    currentGroup.push(sorted[i])
  }
  groups.push(currentGroup)
  return groups
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
      <div v-for="(group, gIdx) in groupedHand" :key="gIdx" class="tile-group">
        <MahjongTile
          v-for="(tile, idx) in group"
          :key="`${tile}-${idx}`"
          :code="tile"
          :is-laizi="tile === gameStore.laiziTile"
          :clickable="isMyTurn"
          @click="handleTileClick(tile)"
        />
      </div>
    </div>
    <TransitionGroup name="drawn" tag="div" class="drawn-tile-area" v-if="gameStore.drawnTile && isMyTurn">
      <span class="drawn-label" key="label">{{ $t('game.drawnLabel') }}</span>
      <MahjongTile
        :key="gameStore.drawnTile"
        :code="gameStore.drawnTile"
        :is-laizi="gameStore.drawnTile === gameStore.laiziTile"
        :clickable="isMyTurn"
        @click="handleTileClick(gameStore.drawnTile!)"
      />
    </TransitionGroup>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.player-hand {
  display: flex;
  justify-content: center;
  align-items: flex-end;
  padding: $spacing-sm 0;
  gap: $spacing-md;
}

.tiles {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  justify-content: center;
}

.tile-group {
  display: flex;
  gap: 2px;
}

.drawn-tile-area {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.drawn-label {
  font-size: 0.65rem;
  color: $color-warning;
  font-weight: 600;
  text-transform: uppercase;
}

// Drawn tile animation — slide in from right
.drawn-enter-active {
  transition: all 0.3s ease-out;
}

.drawn-enter-from {
  opacity: 0;
  transform: translateX(20px);
}
</style>
