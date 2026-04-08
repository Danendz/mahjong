<script setup lang="ts">
import { ref, watch, onUnmounted } from 'vue'
import { useGameStore } from '../../stores/game'

const gameStore = useGameStore()
const remaining = ref(0)
let interval: ReturnType<typeof setInterval> | null = null

function startTimer(seconds: number) {
  stopTimer()
  remaining.value = seconds
  interval = setInterval(() => {
    remaining.value = Math.max(0, remaining.value - 1)
    if (remaining.value <= 0) stopTimer()
  }, 1000)
}

function stopTimer() {
  if (interval) {
    clearInterval(interval)
    interval = null
  }
}

watch(() => gameStore.timeLimit, (limit) => {
  if (limit > 0 && gameStore.isMyTurn) startTimer(limit)
})

watch(() => gameStore.reactionTimeLimit, (limit) => {
  if (limit > 0 && gameStore.isReacting) startTimer(limit)
})

watch([() => gameStore.isMyTurn, () => gameStore.isReacting], ([myTurn, reacting]) => {
  if (!myTurn && !reacting) stopTimer()
})

onUnmounted(stopTimer)
</script>

<template>
  <div v-if="remaining > 0" class="timer" :class="{ urgent: remaining <= 5 }">
    {{ remaining }}s
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.timer {
  font-size: 1.5rem;
  font-weight: 700;
  color: $color-text;
  background: rgba(white, 0.08);
  padding: $spacing-xs $spacing-md;
  border-radius: $border-radius;
  font-variant-numeric: tabular-nums;

  &.urgent {
    color: $color-danger;
    animation: urgentPulse 1s ease-in-out infinite;
  }
}

@keyframes urgentPulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.5; }
}
</style>
