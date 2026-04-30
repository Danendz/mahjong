<script setup lang="ts">
import { computed } from 'vue'
import { useI18n } from 'vue-i18n'
import type { TileCode, MeldInfo } from '../../types/generated'
import MahjongTile from './MahjongTile.vue'

type MeldType = MeldInfo['type']

const props = withDefaults(defineProps<{
  type: MeldType
  tiles: TileCode[] | string[]
  laiziTile?: TileCode | null
  showLabel?: boolean
  small?: boolean
}>(), {
  showLabel: true,
  small: false,
  laiziTile: null,
})

const { t } = useI18n()

const labelKey = computed(() => {
  switch (props.type) {
    case 'chi': return 'game.meld.chi'
    case 'pong': return 'game.meld.pong'
    case 'open_gang': return 'game.meld.openGang'
    case 'closed_gang': return 'game.meld.closedGang'
    case 'add_gang': return 'game.meld.addGang'
    default: return ''
  }
})

const label = computed(() => labelKey.value ? t(labelKey.value) : '')

const typeClass = computed(() => `meld-card--${props.type}`)
const isClosedGang = computed(() => props.type === 'closed_gang')
</script>

<template>
  <div class="meld-card" :class="[typeClass, { 'meld-card--small': small }]">
    <span v-if="showLabel" class="meld-card__badge">{{ label }}</span>
    <div class="meld-card__tiles">
      <MahjongTile
        v-for="(tile, idx) in tiles"
        :key="idx"
        :code="tile"
        :is-laizi="!!laiziTile && tile === laiziTile"
        :face-down="isClosedGang"
        :small="small"
      />
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.meld-card {
  --meld-color: #{$color-meld-pong};
  position: relative;
  display: flex;
  align-items: flex-end;
  padding: 6px 4px 4px;
  border: 1px solid var(--meld-color);
  background: color-mix(in srgb, var(--meld-color) 8%, transparent);
  border-radius: $border-radius-sm;
}

.meld-card--small {
  padding: 3px 2px 2px;
}

.meld-card__badge {
  position: absolute;
  top: -7px;
  left: 4px;
  padding: 1px 5px;
  font-size: 0.65rem;
  font-weight: 700;
  line-height: 1.2;
  color: white;
  background: var(--meld-color);
  border-radius: 2px;
  letter-spacing: 0.02em;
  z-index: 1;
}

.meld-card__tiles {
  display: flex;
  gap: 1px;
}

.meld-card--chi { --meld-color: #{$color-meld-chi}; }
.meld-card--pong { --meld-color: #{$color-meld-pong}; }
.meld-card--open_gang { --meld-color: #{$color-meld-gang}; }
.meld-card--add_gang { --meld-color: #{$color-meld-gang}; }
.meld-card--closed_gang { --meld-color: #{$color-meld-closed-gang}; }

@media (max-width: 480px) {
  .meld-card__badge {
    font-size: 0.6rem;
    top: -6px;
  }
}
</style>
