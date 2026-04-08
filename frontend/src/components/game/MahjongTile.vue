<script setup lang="ts">
import { computed } from 'vue'
import type { TileCode } from '../../types/generated'

const props = defineProps<{
  code: TileCode | string
  isLaizi?: boolean
  clickable?: boolean
  small?: boolean
  faceDown?: boolean
}>()

defineEmits<{ click: [] }>()

const tileDisplay = computed(() => {
  if (props.faceDown) return { char: '', suit: '', suitClass: 'back' }

  const c = props.code as string
  const suitMap: Record<string, string> = {
    m: '萬', s: '條', p: '筒',
  }
  const windMap: Record<string, string> = {
    we: '東', ws: '南', ww: '西', wn: '北',
  }
  const dragonMap: Record<string, string> = {
    dz: '中', df: '發', db: '',
  }

  if (c.length === 2 && '123456789'.includes(c[0])) {
    return {
      char: c[0],
      suit: suitMap[c[1]] || '',
      suitClass: `suit-${c[1]}`,
    }
  }

  if (windMap[c]) {
    return { char: windMap[c], suit: '', suitClass: 'suit-wind' }
  }

  if (dragonMap[c] !== undefined) {
    const cls = c === 'dz' ? 'suit-dragon-red' : c === 'df' ? 'suit-dragon-green' : 'suit-dragon-white'
    return { char: dragonMap[c], suit: '', suitClass: cls }
  }

  return { char: '?', suit: '', suitClass: '' }
})
</script>

<template>
  <div
    class="tile"
    :class="[
      tileDisplay.suitClass,
      {
        laizi: isLaizi,
        clickable: clickable,
        small: small,
        'face-down': faceDown,
      },
    ]"
    @click="clickable && $emit('click')"
  >
    <template v-if="!faceDown">
      <span class="char">{{ tileDisplay.char }}</span>
      <span v-if="tileDisplay.suit" class="suit">{{ tileDisplay.suit }}</span>
    </template>
  </div>
</template>

<style lang="scss" scoped>
@use '../../styles/variables' as *;

.tile {
  width: 36px;
  height: 48px;
  background: $color-tile-bg;
  border: 1px solid $color-tile-border;
  border-radius: 3px;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  font-weight: 700;
  color: $color-tile-text;
  user-select: none;
  position: relative;
  transition: transform 0.15s, box-shadow 0.15s;

  &.small {
    width: 26px;
    height: 36px;
    font-size: 0.75rem;
  }

  &.clickable {
    cursor: pointer;

    &:hover {
      transform: translateY(-6px);
      box-shadow: 0 4px 12px rgba(0, 0, 0, 0.3);
    }
  }

  &.laizi {
    border-color: $color-tile-laizi;
    box-shadow: 0 0 6px rgba($color-tile-laizi, 0.4);

    &::after {
      content: '';
      position: absolute;
      top: 2px;
      right: 2px;
      width: 6px;
      height: 6px;
      background: $color-tile-laizi;
      border-radius: 50%;
    }
  }

  &.face-down {
    background: $color-surface;
    border-color: $color-border;
  }
}

.char {
  font-size: 1.1rem;
  line-height: 1;
}

.suit {
  font-size: 0.55rem;
  line-height: 1;
  opacity: 0.7;
}

.suit-m { color: #333; }
.suit-s { color: #2d7a3a; }
.suit-p { color: #c24040; }
.suit-wind { color: #333; }
.suit-dragon-red { color: #c24040; }
.suit-dragon-green { color: #2d7a3a; }
.suit-dragon-white { color: #666; }
</style>
