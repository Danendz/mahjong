<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import { useGameStore } from '../../stores/game'
import { usePlayerName } from '../../composables/usePlayerName'

const { t } = useI18n()
const gameStore = useGameStore()
const { playerName } = usePlayerName()

const statusText = computed(() => {
  if (gameStore.isReacting) {
    return t('game.status.reactTo', { name: playerName(gameStore.reactionFromSeat) })
  }
  if (gameStore.isMyTurn) {
    return t('game.status.yourTurn')
  }
  if (gameStore.currentTurnSeat >= 0) {
    return t('game.status.otherTurn', { name: playerName(gameStore.currentTurnSeat) })
  }
  return ''
})

const isUrgent = computed(() => gameStore.isMyTurn || gameStore.isReacting)
</script>

<template>
  <div v-if="statusText" class="status-banner" :class="{ urgent: isUrgent }">
    {{ statusText }}
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.status-banner {
  text-align: center;
  font-size: 0.9rem;
  font-weight: 600;
  color: $color-text-muted;
  background: rgba(white, 0.05);
  padding: $spacing-xs $spacing-md;
  border-radius: $border-radius;
  transition: color 0.3s, background 0.3s;

  &.urgent {
    color: $color-warning;
    background: rgba($color-warning, 0.1);
  }
}
</style>
