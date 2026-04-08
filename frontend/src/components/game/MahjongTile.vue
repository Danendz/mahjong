<script setup lang="ts">
import type { TileCode } from '../../types/generated'
import { tileSvgUrl, tileBackUrl } from '../../utils/tileAssets'

const props = defineProps<{
  code: TileCode | string
  isLaizi?: boolean
  clickable?: boolean
  small?: boolean
  faceDown?: boolean
}>()

defineEmits<{ click: [] }>()
</script>

<template>
  <div
    class="tile"
    :class="{
      laizi: isLaizi,
      clickable: clickable,
      small: small,
      'face-down': faceDown,
    }"
    @click="clickable && $emit('click')"
  >
    <img
      v-if="faceDown"
      :src="tileBackUrl()"
      class="tile-img"
      draggable="false"
      alt="tile back"
    />
    <img
      v-else
      :src="tileSvgUrl(code as string)"
      class="tile-img"
      draggable="false"
      :alt="code as string"
    />
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
  align-items: center;
  justify-content: center;
  user-select: none;
  position: relative;
  transition: transform 0.15s, box-shadow 0.15s;
  overflow: hidden;

  &.small {
    width: 26px;
    height: 36px;
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
      z-index: 1;
    }
  }

  &.face-down {
    background: $color-surface;
    border-color: $color-border;
  }
}

.tile-img {
  width: 100%;
  height: 100%;
  object-fit: contain;
  pointer-events: none;
}

@media (max-width: 480px) {
  .tile {
    width: 32px;
    height: 44px;

    &.small {
      width: 24px;
      height: 32px;
    }
  }
}
</style>
