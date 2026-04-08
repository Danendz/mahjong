<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameConnection } from './composables/useGameConnection'

const { status, init } = useGameConnection()

onMounted(() => {
  init()
})
</script>

<template>
  <div class="app">
    <div v-if="status === 'reconnecting'" class="reconnecting-banner">
      Reconnecting...
    </div>
    <router-view />
  </div>
</template>

<style lang="scss">
@use './styles/global';

.app {
  height: 100%;
  position: relative;
}

.reconnecting-banner {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: #f0a500;
  color: #1a1a2e;
  text-align: center;
  padding: 8px;
  font-weight: 600;
  z-index: 1000;
}
</style>
