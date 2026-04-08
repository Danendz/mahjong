<script setup lang="ts">
import { onMounted } from 'vue'
import { useGameConnection } from './composables/useGameConnection'
import ConnectionStatus from './components/ConnectionStatus.vue'

const { init } = useGameConnection()

onMounted(() => {
  init()
})
</script>

<template>
  <div class="app">
    <ConnectionStatus />
    <router-view />
  </div>
</template>

<style lang="scss">
@use './styles/global';

.app {
  height: 100%;
  position: relative;
}
</style>
