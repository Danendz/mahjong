<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useGameStore } from '../../stores/game'
import { usePlayerName } from '../../composables/usePlayerName'
import MahjongTile from './MahjongTile.vue'

const gameStore = useGameStore()
const { playerName } = usePlayerName()

// Next round countdown
const countdown = ref(30)
let countdownInterval: ReturnType<typeof setInterval> | null = null

onMounted(() => {
  countdownInterval = setInterval(() => {
    if (countdown.value > 0) {
      countdown.value--
    } else if (countdownInterval) {
      clearInterval(countdownInterval)
      countdownInterval = null
    }
  }, 1000)
})

onUnmounted(() => {
  if (countdownInterval) {
    clearInterval(countdownInterval)
    countdownInterval = null
  }
})

// Player rankings sorted by score descending
const rankColors = ['#f0a500', '#999', '#cd7f32', '#666']

const rankedScores = computed(() => {
  const scores = gameStore.roundResult?.total_scores
  if (!scores) return []
  return Object.entries(scores)
    .map(([seat, score]) => ({ seat: Number(seat), score: score as number }))
    .sort((a, b) => b.score - a.score)
    .map((entry, idx) => ({
      ...entry,
      rank: idx + 1,
      color: rankColors[idx] ?? '#666',
    }))
})
</script>

<template>
  <div v-if="gameStore.roundResult" class="overlay">
    <div class="overlay-card">
      <h2 v-if="gameStore.roundResult.result === 'hu'">
        {{ $t('scoring.wins', { name: playerName(gameStore.roundResult.winner_seat!) }) }}
      </h2>
      <h2 v-else>{{ $t('scoring.draw') }}</h2>

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
          {{ $t('scoring.base') }} {{ gameStore.roundResult.scoring.base_points }}
        </div>
        <div
          v-for="(mult, idx) in gameStore.roundResult.scoring.multipliers"
          :key="idx"
          class="score-line"
        >
          {{ mult.reason }}: x{{ mult.value }}
        </div>
        <div class="score-total">
          {{ $t('scoring.perLoser', { amount: gameStore.roundResult.scoring.total_per_loser }) }}
          <span v-if="gameStore.roundResult.scoring.capped">{{ $t('scoring.capped') }}</span>
        </div>
      </div>

      <div class="score-deltas">
        <div
          v-for="(delta, seat) in gameStore.roundResult.score_deltas"
          :key="seat"
          class="delta"
          :class="{ positive: delta > 0, negative: delta < 0 }"
        >
          {{ $t('scoring.delta', { name: playerName(Number(seat)), delta: (delta > 0 ? '+' : '') + delta }) }}
        </div>
      </div>

      <div class="total-scores">
        <div v-for="entry in rankedScores" :key="entry.seat" class="ranked-player">
          <span class="rank-badge" :style="{ backgroundColor: entry.color }">
            {{ entry.rank }}
          </span>
          {{ $t('scoring.rankedScore', { name: playerName(entry.seat), score: entry.score }) }}
        </div>
      </div>

      <div v-if="countdown > 0" class="countdown">
        {{ $t('scoring.nextRoundIn', { countdown }) }}
      </div>
      <div v-else class="countdown">
        {{ $t('scoring.starting') }}
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
  flex-direction: column;
  align-items: center;
  gap: $spacing-xs;
}

.ranked-player {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

.rank-badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  color: #fff;
  font-size: 0.7rem;
  font-weight: 700;
  flex-shrink: 0;
}

.countdown {
  margin-top: $spacing-lg;
  font-size: 0.85rem;
  color: $color-text-muted;
}
</style>
