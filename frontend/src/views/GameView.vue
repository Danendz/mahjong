<script setup lang="ts">
import { computed } from 'vue'
import { useGameStore } from '../stores/game'
import { useGameConnection } from '../composables/useGameConnection'
import PlayerHand from '../components/game/PlayerHand.vue'
import DiscardPool from '../components/game/DiscardPool.vue'
import OpponentStrip from '../components/game/OpponentStrip.vue'
import ActionBar from '../components/game/ActionBar.vue'
import LaiziIndicator from '../components/game/LaiziIndicator.vue'
import GameInfo from '../components/game/GameInfo.vue'
import ScoringOverlay from '../components/game/ScoringOverlay.vue'

defineProps<{ code: string }>()

const gameStore = useGameStore()
const conn = useGameConnection()

const opponents = computed(() => {
  const seat = gameStore.yourSeat
  return [
    (seat + 1) % 4, // right
    (seat + 2) % 4, // across
    (seat + 3) % 4, // left
  ]
})
</script>

<template>
  <div class="game">
    <div class="game-top">
      <OpponentStrip
        v-for="opp in opponents"
        :key="opp"
        :seat="opp"
        :position="opponents.indexOf(opp)"
      />
    </div>

    <div class="game-center">
      <LaiziIndicator />
      <DiscardPool />
      <GameInfo />
    </div>

    <div class="game-bottom">
      <ActionBar
        @discard="conn.discard($event)"
        @pong="conn.declarePong()"
        @chi="conn.declareChi($event)"
        @gang="conn.declareGang($event.type, $event.tile)"
        @hu="conn.declareHu()"
        @pass="conn.declarePass()"
      />
      <PlayerHand @discard="conn.discard($event)" />
    </div>

    <ScoringOverlay v-if="gameStore.isRoundEnd" />
  </div>
</template>

<style lang="scss" scoped>
@use '../styles/variables' as *;

.game {
  height: 100%;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  background: darken($color-bg, 3%);
}

.game-top {
  display: flex;
  justify-content: center;
  gap: $spacing-sm;
  padding: $spacing-sm;
  min-height: 80px;
}

.game-center {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: $spacing-md;
  position: relative;
  padding: $spacing-sm;
}

.game-bottom {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
  padding: $spacing-sm;
}
</style>
