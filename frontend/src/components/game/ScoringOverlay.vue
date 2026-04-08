<script setup lang="ts">
import { useGameStore } from '../../stores/game'
import { useRoomStore } from '../../stores/room'
import MahjongTile from './MahjongTile.vue'

const gameStore = useGameStore()
const roomStore = useRoomStore()

function playerName(seat: number): string {
  return roomStore.players.find(p => p.seat === seat)?.nickname || `Player ${seat}`
}
</script>

<template>
  <div v-if="gameStore.roundResult" class="overlay">
    <div class="overlay-card">
      <h2 v-if="gameStore.roundResult.result === 'hu'">
        {{ playerName(gameStore.roundResult.winner_seat!) }} Wins!
      </h2>
      <h2 v-else>Draw</h2>

      <div v-if="gameStore.roundResult.winning_hand" class="winning-hand">
        <MahjongTile
          v-for="(tile, idx) in gameStore.roundResult.winning_hand"
          :key="idx"
          :code="tile"
          :is-laizi="tile === gameStore.laiziTile"
          small
        />
      </div>

      <div v-if="gameStore.roundResult.scoring" class="scoring">
        <div class="score-line">
          Base: {{ gameStore.roundResult.scoring.base_points }}
        </div>
        <div
          v-for="(mult, idx) in gameStore.roundResult.scoring.multipliers"
          :key="idx"
          class="score-line"
        >
          {{ mult.reason }}: x{{ mult.value }}
        </div>
        <div class="score-total">
          Per loser: {{ gameStore.roundResult.scoring.total_per_loser }}
          <span v-if="gameStore.roundResult.scoring.capped">(capped)</span>
        </div>
      </div>

      <div class="score-deltas">
        <div
          v-for="(delta, seat) in gameStore.roundResult.score_deltas"
          :key="seat"
          class="delta"
          :class="{ positive: delta > 0, negative: delta < 0 }"
        >
          {{ playerName(Number(seat)) }}: {{ delta > 0 ? '+' : '' }}{{ delta }}
        </div>
      </div>

      <div class="total-scores">
        <div v-for="(score, seat) in gameStore.roundResult.total_scores" :key="seat">
          {{ playerName(Number(seat)) }}: {{ score }}
        </div>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.overlay {
  position: fixed;
  inset: 0;
  background: rgba(0, 0, 0, 0.7);
  display: flex;
  align-items: center;
  justify-content: center;
  z-index: 100;
}

.overlay-card {
  background: $color-bg-light;
  border: 1px solid $color-border;
  border-radius: $border-radius;
  padding: $spacing-xl;
  max-width: 450px;
  width: 90%;
  text-align: center;

  h2 {
    margin-bottom: $spacing-lg;
    font-size: 1.5rem;
  }
}

.winning-hand {
  display: flex;
  justify-content: center;
  flex-wrap: wrap;
  gap: 2px;
  margin-bottom: $spacing-lg;
}

.scoring {
  margin-bottom: $spacing-lg;
  font-size: 0.9rem;
}

.score-line {
  color: $color-text-muted;
  padding: $spacing-xs 0;
}

.score-total {
  font-weight: 700;
  font-size: 1.1rem;
  margin-top: $spacing-sm;
}

.score-deltas {
  display: flex;
  flex-direction: column;
  gap: $spacing-xs;
  margin-bottom: $spacing-lg;
}

.delta {
  font-weight: 600;

  &.positive { color: $color-success; }
  &.negative { color: $color-danger; }
}

.total-scores {
  font-size: 0.85rem;
  color: $color-text-muted;
  display: flex;
  justify-content: center;
  gap: $spacing-md;
}
</style>
