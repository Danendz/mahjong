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
  gameStore.availableActions.includes('gang') || gameStore.canGang.length > 0
)

function handleChi(option: [string, string]) {
  showChiOptions.value = false
  emit('chi', option)
}

function handleGang() {
  if (showReactions.value) {
    emit('gang', { type: 'open', tile: gameStore.reactionTile! })
  } else if (gameStore.canGang.length === 1) {
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
        胡
      </button>
      <button v-if="hasGang" class="btn-action gang" @click="handleGang">
        杠
      </button>
      <button v-if="hasPong" class="btn-action pong" @click="$emit('pong')">
        碰
      </button>
      <button v-if="hasChi && !showChiOptions" class="btn-action chi" @click="showChiOptions = gameStore.chiOptions.length > 1 ? true : false; if (gameStore.chiOptions.length === 1) handleChi(gameStore.chiOptions[0] as [string, string])">
        吃
      </button>
      <button class="btn-action pass" @click="$emit('pass')">
        过
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
        自摸
      </button>
      <button v-if="gameStore.canGang.length > 0" class="btn-action gang" @click="handleGang">
        杠
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
  padding: $spacing-sm $spacing-lg;
  border-radius: $border-radius;
  font-size: 1.2rem;
  font-weight: 700;
  border: none;
  cursor: pointer;
  transition: transform 0.1s;

  &:active { transform: scale(0.95); }
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
