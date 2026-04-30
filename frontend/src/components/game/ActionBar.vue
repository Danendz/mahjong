<script setup lang="ts">
import { computed, ref } from 'vue'
import { useGameStore } from '../../stores/game'
import type { TileCode } from '../../types/generated'
import MahjongTile from './MahjongTile.vue'

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
const showGangOptions = ref(false)

const showReactions = computed(() => gameStore.isReacting)
const showTurnActions = computed(() => gameStore.isMyTurn && !gameStore.isReacting)

const hasHu = computed(() => gameStore.availableActions.includes('hu') || gameStore.canHu)
const hasPong = computed(() => gameStore.availableActions.includes('pong'))
const hasChi = computed(() => gameStore.availableActions.includes('chi'))
const hasReactionGang = computed(() => gameStore.availableActions.includes('gang'))
const hasTurnGang = computed(() => (gameStore.canGang?.length ?? 0) > 0)

// Sort chi tiles ascending by numeric rank, returning the full 3-tile sequence
// (the two from hand + the contested tile). All three are guaranteed same suit.
function chiSequence(opt: [TileCode, TileCode] | [string, string]): string[] {
  const contested = gameStore.reactionTile ?? ''
  const all: string[] = [opt[0], opt[1], contested]
  return all.sort((a, b) => parseInt(a[0], 10) - parseInt(b[0], 10))
}

function isContestedInChi(tile: string): boolean {
  return tile === gameStore.reactionTile
}

// For each tile in canGang, decide whether it's closed (4 in hand) or add (you already
// have a pong of that tile and just drew the 4th to upgrade).
function gangTypeFor(tile: TileCode): 'closed' | 'add' {
  const myMelds = gameStore.openMelds[String(gameStore.yourSeat)] || []
  const hasPongOfTile = myMelds.some(m => m.type === 'pong' && m.tiles.includes(tile))
  return hasPongOfTile ? 'add' : 'closed'
}

function handleChiClick() {
  if (gameStore.chiOptions.length === 1) {
    handleChi(gameStore.chiOptions[0] as [string, string])
  } else {
    showChiOptions.value = true
  }
}

function handleChi(option: [string, string]) {
  showChiOptions.value = false
  emit('chi', option)
}

function handleReactionGang() {
  emit('gang', { type: 'open', tile: gameStore.reactionTile! })
}

function handleTurnGangClick() {
  const opts = gameStore.canGang
  if (!opts || opts.length === 0) return
  if (opts.length === 1) {
    const tile = opts[0]
    emit('gang', { type: gangTypeFor(tile), tile })
  } else {
    showGangOptions.value = true
  }
}

function handleGangOption(tile: TileCode) {
  showGangOptions.value = false
  emit('gang', { type: gangTypeFor(tile), tile })
}
</script>

<template>
  <div v-if="showReactions || showTurnActions" class="action-bar">
    <!-- Chi options sub-menu (mini-tile sequences) -->
    <div v-if="showChiOptions" class="chi-options">
      <button
        v-for="(opt, idx) in gameStore.chiOptions"
        :key="idx"
        class="btn-action chi-option"
        @click="handleChi(opt as [string, string])"
      >
        <MahjongTile
          v-for="(tile, tidx) in chiSequence(opt as [string, string])"
          :key="tidx"
          :code="tile"
          :contested="isContestedInChi(tile)"
          :contested-color="isContestedInChi(tile) ? '#5b8def' : undefined"
          small
        />
      </button>
      <button class="btn-action pass" @click="showChiOptions = false">
        {{ $t('common.cancel') }}
      </button>
    </div>

    <!-- Gang options sub-menu (mini-tile quads, badge per type) -->
    <div v-else-if="showGangOptions" class="gang-options">
      <button
        v-for="(tile, idx) in gameStore.canGang"
        :key="idx"
        class="btn-action gang-option"
        :class="`gang-option--${gangTypeFor(tile)}`"
        @click="handleGangOption(tile)"
      >
        <span class="gang-option__badge">
          {{ gangTypeFor(tile) === 'closed'
              ? $t('game.meld.closedGang')
              : $t('game.meld.addGang') }}
        </span>
        <MahjongTile v-for="i in 4" :key="i" :code="tile" small />
      </button>
      <button class="btn-action pass" @click="showGangOptions = false">
        {{ $t('common.cancel') }}
      </button>
    </div>

    <!-- Reaction buttons -->
    <template v-else-if="showReactions">
      <button v-if="hasHu" class="btn-action hu" @click="$emit('hu')">
        <span class="action-cn">{{ $t('game.action.hu') }}</span>
        <span v-if="$t('game.action.huSub')" class="action-en">
          {{ $t('game.action.huSub') }}<template v-if="gameStore.huScorePreview"> ({{ $t('game.action.huScorePreview', { points: gameStore.huScorePreview }) }})</template>
        </span>
        <span v-else-if="gameStore.huScorePreview" class="action-en">
          {{ $t('game.action.huScorePreview', { points: gameStore.huScorePreview }) }}
        </span>
      </button>
      <button v-if="hasReactionGang" class="btn-action gang" @click="handleReactionGang">
        <span class="action-cn">{{ $t('game.action.gang') }}</span>
        <span v-if="$t('game.action.gangSub')" class="action-en">{{ $t('game.action.gangSub') }}</span>
      </button>
      <button v-if="hasPong" class="btn-action pong" @click="$emit('pong')">
        <span class="action-cn">{{ $t('game.action.pong') }}</span>
        <span v-if="$t('game.action.pongSub')" class="action-en">{{ $t('game.action.pongSub') }}</span>
      </button>
      <button v-if="hasChi" class="btn-action chi" @click="handleChiClick">
        <span class="action-cn">{{ $t('game.action.chi') }}</span>
        <span v-if="$t('game.action.chiSub')" class="action-en">{{ $t('game.action.chiSub') }}</span>
      </button>
      <button class="btn-action pass" @click="$emit('pass')">
        <span class="action-cn">{{ $t('game.action.pass') }}</span>
        <span v-if="$t('game.action.passSub')" class="action-en">{{ $t('game.action.passSub') }}</span>
      </button>
    </template>

    <!-- Turn actions (kong, hu on self-draw) -->
    <template v-else-if="showTurnActions">
      <button v-if="gameStore.canHu" class="btn-action hu" @click="$emit('hu')">
        <span class="action-cn">{{ $t('game.action.zimo') }}</span>
        <span v-if="$t('game.action.zimoSub')" class="action-en">
          {{ $t('game.action.zimoSub') }}<template v-if="gameStore.huScorePreview"> ({{ $t('game.action.huScorePreview', { points: gameStore.huScorePreview }) }})</template>
        </span>
        <span v-else-if="gameStore.huScorePreview" class="action-en">
          {{ $t('game.action.huScorePreview', { points: gameStore.huScorePreview }) }}
        </span>
      </button>
      <button v-if="hasTurnGang" class="btn-action gang" @click="handleTurnGangClick">
        <span class="action-cn">{{ $t('game.action.gang') }}</span>
        <span v-if="$t('game.action.gangSub')" class="action-en">{{ $t('game.action.gangSub') }}</span>
      </button>
    </template>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.action-bar {
  display: flex;
  justify-content: center;
  align-items: center;
  gap: $spacing-sm;
  padding: $spacing-sm;
  flex-wrap: wrap;
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
  background: $color-meld-chi;
  color: white;
}

.pass {
  background: $color-surface;
  color: $color-text;
  border: 1px solid $color-border;
}

// Chi / Gang option submenus — buttons render mini-tile previews instead of text
.chi-options,
.gang-options {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  flex-wrap: wrap;
  justify-content: center;
}

.chi-option {
  position: relative;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 2px;
  padding: 6px 10px;
  background: rgba($color-meld-chi, 0.18);
  border: 1px solid $color-meld-chi;
  color: white;
  font-size: 0.85rem;
}

.gang-option {
  position: relative;
  display: flex;
  flex-direction: row;
  align-items: center;
  gap: 2px;
  padding: 10px 10px 6px;
  color: white;
  font-size: 0.85rem;

  &--closed {
    background: rgba($color-meld-closed-gang, 0.18);
    border: 1px solid $color-meld-closed-gang;
  }

  &--add {
    background: rgba($color-meld-gang, 0.18);
    border: 1px solid $color-meld-gang;
  }
}

.gang-option__badge {
  position: absolute;
  top: -8px;
  left: 4px;
  padding: 1px 5px;
  font-size: 0.6rem;
  font-weight: 700;
  line-height: 1.2;
  color: white;
  border-radius: 2px;

  .gang-option--closed & {
    background: $color-meld-closed-gang;
  }

  .gang-option--add & {
    background: $color-meld-gang;
  }
}
</style>
