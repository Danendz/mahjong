<script setup lang="ts">
import { TransitionGroup } from 'vue'
import { useToast } from '../composables/useToast'

const { toasts } = useToast()
</script>

<template>
  <TransitionGroup name="toast" tag="div" class="toasts">
    <div
      v-for="t in toasts"
      :key="t.id"
      class="toast"
      :class="`toast--${t.kind}`"
      role="status"
    >
      {{ t.message }}
    </div>
  </TransitionGroup>
</template>

<style lang="scss" scoped>
@use '../styles/variables' as *;

.toasts {
  position: fixed;
  top: $spacing-md;
  left: 50%;
  transform: translateX(-50%);
  display: flex;
  flex-direction: column;
  gap: $spacing-xs;
  z-index: 200;
  pointer-events: none;
  max-width: 90vw;
}

.toast {
  pointer-events: auto;
  padding: $spacing-sm $spacing-md;
  border-radius: $border-radius;
  font-size: 0.85rem;
  font-weight: 600;
  background: $color-bg-light;
  border: 1px solid $color-border;
  color: $color-text;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.4);
  text-align: center;

  &--warn {
    border-color: $color-warning;
    background: color-mix(in srgb, $color-warning 18%, $color-bg-light);
  }

  &--error {
    border-color: $color-danger;
    background: color-mix(in srgb, $color-danger 18%, $color-bg-light);
  }
}

.toast-enter-active,
.toast-leave-active {
  transition: opacity 0.2s ease, transform 0.2s ease;
}

.toast-enter-from {
  opacity: 0;
  transform: translateY(-8px);
}

.toast-leave-to {
  opacity: 0;
  transform: translateY(-4px);
}
</style>
