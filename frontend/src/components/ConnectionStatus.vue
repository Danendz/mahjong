<script setup lang="ts">
import { useGameConnection } from '../composables/useGameConnection'

const { status } = useGameConnection()
</script>

<template>
  <div v-if="status !== 'connected'" class="connection-status" :class="status">
    <template v-if="status === 'connecting'">
      <span class="dot"></span> {{ $t('connection.connecting') }}
    </template>
    <template v-else-if="status === 'reconnecting'">
      <span class="dot"></span> {{ $t('connection.reconnecting') }}
    </template>
    <template v-else>
      <span class="dot"></span> {{ $t('connection.disconnected') }}
    </template>
  </div>
</template>

<style lang="scss" scoped>
@use '../styles/variables' as *;

.connection-status {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  text-align: center;
  padding: 6px;
  font-size: 0.85rem;
  font-weight: 600;
  z-index: 1000;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: $spacing-sm;

  &.connecting, &.reconnecting {
    background: $color-warning;
    color: $color-bg;
  }

  &.disconnected {
    background: $color-danger;
    color: white;
  }
}

.dot {
  width: 8px;
  height: 8px;
  border-radius: 50%;
  background: currentColor;
  animation: pulse 1.5s ease-in-out infinite;
}

@keyframes pulse {
  0%, 100% { opacity: 1; }
  50% { opacity: 0.3; }
}
</style>
