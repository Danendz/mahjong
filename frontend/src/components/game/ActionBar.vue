<script setup lang="ts">
import { computed, ref } from 'vue'
import { useGameStore } from '../../stores/game'

const emit = defineEmits<{
  discard: [tile: string]
  pong: []
  chi: [tiles: [string, string]]
  gang: [data: { type: 'open' | 'closed' | 'add'; tile: string }]
  hu: []
  pass: []
}>()

const gameStore = useGameStore()
const showChiOptions = ref(false)

const showReactions = computed(() => gameStore.isReacting)
const showTurnActions = computed(() => gameStore.isMyTurn && !gameStore.isReacting)

const hasHu = computed(() => gameStore.availableActions.includes('hu') || gameStore.canHu)
const hasPong = computed(() => gameStore.availableActions.includes('pong'))
const hasChi = computed(() => gameStore.availableActions.includes('chi'))
const hasGang = computed(() =>
  gameStore.availableActions.includes('gang') || (gameStore.canGang?.length ?? 0) > 0
)

function handleChi(option: [string, string]) {
  showChiOptions.value = false
  emit('chi', option)
}

function handleGang() {
  if (showReactions.value) {
    emit('gang', { type: 'open', tile: gameStore.reactionTile! })
  } else if (gameStore.canGang?.length === 1) {
    // Auto-select the only option
    const tile = gameStore.canGang[0]
    // Determine type: check if it's closed or add
    emit('gang', { type: 'closed', tile })
  }
}
</script>

<template>
  <div v-if="showReactions || showTurnActions" class="action-bar">
    <!-- Reaction buttons -->
    <template v-if="showReactions">
      <button v-if="hasHu" class="btn-action hu" @click="$emit('hu')">
        <span class="action-cn">胡</span>
        <span class="action-en">Win{{ gameStore.huScorePreview ? ` (${gameStore.huScorePreview}pts)` : '' }}</span>
      </button>
      <button v-if="hasGang" class="btn-action gang" @click="handleGang">
        <span class="action-cn">杠</span>
        <span class="action-en">Kong</span>
      </button>
      <button v-if="hasPong" class="btn-action pong" @click="$emit('pong')">
        <span class="action-cn">碰</span>
        <span class="action-en">Pong</span>
      </button>
      <button v-if="hasChi && !showChiOptions" class="btn-action chi" @click="showChiOptions = gameStore.chiOptions.length > 1 ? true : false; if (gameStore.chiOptions.length === 1) handleChi(gameStore.chiOptions[0] as [string, string])">
        <span class="action-cn">吃</span>
        <span class="action-en">Chi</span>
      </button>
      <button class="btn-action pass" @click="$emit('pass')">
        <span class="action-cn">过</span>
        <span class="action-en">Pass</span>
      </button>
    </template>

    <!-- Chi options sub-menu -->
    <div v-if="showChiOptions" class="chi-options">
      <button
        v-for="(opt, idx) in gameStore.chiOptions"
        :key="idx"
        class="btn-action chi-option"
        @click="handleChi(opt as [string, string])"
      >
        {{ opt[0] }} + {{ opt[1] }}
      </button>
      <button class="btn-action pass" @click="showChiOptions = false">
        Cancel
      </button>
    </div>

    <!-- Turn actions (kong, hu on self-draw) -->
    <template v-if="showTurnActions && !showReactions">
      <button v-if="gameStore.canHu" class="btn-action hu" @click="$emit('hu')">
        <span class="action-cn">自摸</span>
        <span class="action-en">Zimo{{ gameStore.huScorePreview ? ` (${gameStore.huScorePreview}pts)` : '' }}</span>
      </button>
      <button v-if="(gameStore.canGang?.length ?? 0) > 0" class="btn-action gang" @click="handleGang">
        <span class="action-cn">杠</span>
        <span class="action-en">Kong</span>
      </button>
    </template>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.action-bar {
  display: flex;
  justify-content: center;
  gap: $spacing-sm;
  padding: $spacing-sm;
}

.btn-action {
  display: flex;
  flex-direction: column;
  align-items: center;
  padding: $spacing-sm $spacing-lg;
  border-radius: $border-radius;
  font-size: 1.2rem;
  font-weight: 700;
  border: none;
  cursor: pointer;
  transition: transform 0.1s;

  &:active { transform: scale(0.95); }
}

.action-cn {
  font-size: 1.2rem;
  line-height: 1;
}

.action-en {
  font-size: 0.6rem;
  opacity: 0.8;
  line-height: 1;
  margin-top: 2px;
}

.hu {
  background: $color-danger;
  color: white;
}

.gang {
  background: $color-warning;
  color: $color-bg;
}

.pong {
  background: $color-success;
  color: $color-bg;
}

.chi {
  background: #5b8def;
  color: white;
}

.pass {
  background: $color-surface;
  color: $color-text;
  border: 1px solid $color-border;
}

.chi-options {
  display: flex;
  gap: $spacing-xs;
}

.chi-option {
  background: #5b8def;
  color: white;
  font-size: 0.9rem;
  padding: $spacing-sm $spacing-md;
}
</style>
