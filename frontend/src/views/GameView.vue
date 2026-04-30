<script setup lang="ts">
import { computed, watch, TransitionGroup } from 'vue'
import { useGameStore } from '../stores/game'
import { useGameConnection } from '../composables/useGameConnection'
import { usePlayerName } from '../composables/usePlayerName'
import { useTurnSound } from '../composables/useTurnSound'
import PlayerHand from '../components/game/PlayerHand.vue'
import PlayerArea from '../components/game/PlayerArea.vue'
import ActionBar from '../components/game/ActionBar.vue'
import LaiziIndicator from '../components/game/LaiziIndicator.vue'
import GameInfo from '../components/game/GameInfo.vue'
import TurnTimer from '../components/game/TurnTimer.vue'
import StatusBanner from '../components/game/StatusBanner.vue'
import WallDisplay from '../components/game/WallDisplay.vue'
import ScoringOverlay from '../components/game/ScoringOverlay.vue'
import MahjongTile from '../components/game/MahjongTile.vue'
import MeldCard from '../components/game/MeldCard.vue'

defineProps<{ code: string }>()

const gameStore = useGameStore()
const conn = useGameConnection()
const { playerName } = usePlayerName()

// Seat positions relative to current player (table orientation)
const acrossSeat = computed(() => (gameStore.yourSeat + 2) % 4)
const rightSeat = computed(() => (gameStore.yourSeat + 1) % 4)
const leftSeat = computed(() => (gameStore.yourSeat + 3) % 4)

const myDiscards = computed(() => gameStore.discards[String(gameStore.yourSeat)] || [])
const myMelds = computed(() => gameStore.openMelds[String(gameStore.yourSeat)] || [])
const myName = computed(() => playerName(gameStore.yourSeat))

// Turn notification sound
const { playTurnSound } = useTurnSound()
watch(() => gameStore.turnVersion, () => {
  if (gameStore.isMyTurn || gameStore.isReacting) playTurnSound()
})
</script>

<template>
  <div class="game-table">
    <!-- Across player (top) -->
    <div class="area-across">
      <PlayerArea :seat="acrossSeat" position="across" />
    </div>

    <!-- Left player (middle-left) -->
    <div class="area-left">
      <PlayerArea :seat="leftSeat" position="left" />
    </div>

    <!-- Center area -->
    <div class="area-center">
      <StatusBanner />
      <div class="info-row">
        <LaiziIndicator />
        <WallDisplay />
        <TurnTimer />
      </div>
      <GameInfo />
    </div>

    <!-- Right player (middle-right) -->
    <div class="area-right">
      <PlayerArea :seat="rightSeat" position="right" />
    </div>

    <!-- Self area (bottom) — order: discards → actions → tenpai → melds dock → hand -->
    <div class="area-self" :class="{ 'active-turn': gameStore.isMyTurn }">
      <div v-if="myDiscards.length" class="self-discards">
        <span class="discard-label">{{ myName }}</span>
        <div class="discard-tiles">
          <MahjongTile
            v-for="(tile, idx) in myDiscards"
            :key="idx"
            :code="tile"
            :is-laizi="tile === gameStore.laiziTile"
            small
          />
        </div>
      </div>
      <ActionBar
        @discard="conn.discard($event)"
        @pong="conn.declarePong()"
        @chi="conn.declareChi($event)"
        @gang="conn.declareGang($event.type, $event.tile)"
        @hu="conn.declareHu()"
        @pass="conn.declarePass()"
      />
      <div v-if="gameStore.waitingTiles.length > 0" class="tenpai-hint">
        <span class="tenpai-label">{{ $t('game.tenpaiLabel') }}</span>
        <MahjongTile
          v-for="(tile, idx) in gameStore.waitingTiles"
          :key="idx"
          :code="tile"
          :is-laizi="tile === gameStore.laiziTile"
          small
        />
      </div>
      <TransitionGroup v-if="myMelds.length" name="meld" tag="div" class="self-melds">
        <MeldCard
          v-for="(meld, idx) in myMelds"
          :key="idx"
          :type="meld.type"
          :tiles="meld.tiles"
          :laizi-tile="gameStore.laiziTile"
        />
      </TransitionGroup>
      <PlayerHand @discard="conn.discard($event)" />
    </div>

    <ScoringOverlay v-if="gameStore.isRoundEnd" />
  </div>
</template>

<style lang="scss" scoped>
@use 'sass:color';
@use '../styles/variables' as *;

.game-table {
  height: 100%;
  display: grid;
  grid-template-areas:
    ".      across   ."
    "left   center   right"
    "self   self     self";
  grid-template-columns: minmax(100px, 1fr) 2fr minmax(100px, 1fr);
  grid-template-rows: auto 1fr auto;
  gap: $spacing-xs;
  padding: $spacing-xs;
  overflow: hidden;
  background: color.adjust($color-bg, $lightness: -3%);
}

.area-across {
  grid-area: across;
}

.area-left {
  grid-area: left;
  display: flex;
  align-items: flex-start;
}

.area-center {
  grid-area: center;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;
}

.area-right {
  grid-area: right;
  display: flex;
  align-items: flex-start;
  justify-content: flex-end;
}

.area-self {
  grid-area: self;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: $spacing-xs;
  border-top: 2px solid transparent;

  &.active-turn {
    background: rgba($color-warning, 0.06);
    border-radius: $border-radius;
    box-shadow: 0 -2px 16px rgba($color-warning, 0.15);
    border-top-color: rgba($color-warning, 0.4);
  }
}

.info-row {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
}

// Self melds dock — sits directly above the hand
.self-melds {
  display: flex;
  flex-wrap: wrap;
  gap: $spacing-md;
  justify-content: center;
  padding-top: 6px;  // room for the corner badge
}

// Self discards
.self-discards {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 2px;
}

.discard-label {
  font-size: 0.7rem;
  color: $color-text-muted;
}

.discard-tiles {
  display: flex;
  gap: 1px;
  flex-wrap: wrap;
  justify-content: center;
  max-width: 350px;
}

// Meld claim animation — fade + scale + brief pulse
.meld-enter-active {
  transition: all 0.3s ease-out;
  animation: meldPulse 0.6s ease-out 0.3s 1;
}

.meld-enter-from {
  opacity: 0;
  transform: scale(0.8);
}

@keyframes meldPulse {
  0%, 100% { filter: none; }
  50% { filter: brightness(1.25) saturate(1.2); }
}

// Tenpai hint
.tenpai-hint {
  display: flex;
  align-items: center;
  gap: $spacing-xs;
  padding: $spacing-xs $spacing-sm;
  background: rgba($color-success, 0.1);
  border: 1px solid rgba($color-success, 0.3);
  border-radius: $border-radius;
}

.tenpai-label {
  font-size: 0.85rem;
  font-weight: 700;
  color: $color-success;
}

// Mobile responsive
@media (max-width: $breakpoint-mobile) {
  .game-table {
    grid-template-areas:
      "across  across"
      "left    right"
      "center  center"
      "self    self";
    grid-template-columns: 1fr 1fr;
    grid-template-rows: auto auto auto auto;
    gap: $spacing-xs;
    padding: $spacing-xs;
  }

  .area-center {
    gap: $spacing-xs;
  }

  .self-melds {
    gap: $spacing-sm;
  }
}

@media (max-width: $breakpoint-tablet) and (min-width: $breakpoint-mobile + 1) {
  .game-table {
    grid-template-columns: minmax(80px, 1fr) 2fr minmax(80px, 1fr);
  }
}
</style>
