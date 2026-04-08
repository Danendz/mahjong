<script setup lang="ts">
import { computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useUserStore } from '../stores/user'
import { useGameConnection } from '../composables/useGameConnection'

const props = defineProps<{ code: string }>()
const router = useRouter()
const roomStore = useRoomStore()
const userStore = useUserStore()
const { toggleReady, startGame, leaveRoom } = useGameConnection()

const isHost = computed(() => userStore.seat === 0)
const canStart = computed(() => roomStore.allReady && roomStore.playerCount === 4)

const seatLabels = ['East', 'South', 'West', 'North']

watch(() => roomStore.status, (status) => {
  if (status === 'playing') {
    router.push({ name: 'game', params: { code: props.code } })
  }
})

function handleLeave() {
  leaveRoom()
  roomStore.$reset()
  router.push({ name: 'lobby' })
}

function copyCode() {
  navigator.clipboard.writeText(props.code)
}
</script>

<template>
  <div class="room">
    <div class="room-card">
      <div class="room-header">
        <h2>Room</h2>
        <div class="code-display" @click="copyCode" title="Click to copy">
          {{ code }}
        </div>
      </div>

      <div class="players">
        <div
          v-for="seatIdx in 4"
          :key="seatIdx - 1"
          class="player-slot"
          :class="{
            occupied: roomStore.players.find(p => p.seat === seatIdx - 1),
            ready: roomStore.players.find(p => p.seat === seatIdx - 1)?.ready,
            you: seatIdx - 1 === userStore.seat,
          }"
        >
          <div class="seat-label">{{ seatLabels[seatIdx - 1] }}</div>
          <template v-if="roomStore.players.find(p => p.seat === seatIdx - 1)">
            <div class="player-name">
              {{ roomStore.players.find(p => p.seat === seatIdx - 1)!.nickname }}
              <span v-if="seatIdx - 1 === userStore.seat" class="you-tag">(you)</span>
              <span v-if="seatIdx - 1 === 0" class="host-tag">HOST</span>
            </div>
            <div class="ready-status" :class="{ active: roomStore.players.find(p => p.seat === seatIdx - 1)?.ready }">
              {{ roomStore.players.find(p => p.seat === seatIdx - 1)?.ready ? 'Ready' : 'Not ready' }}
            </div>
          </template>
          <template v-else>
            <div class="empty">Waiting...</div>
          </template>
        </div>
      </div>

      <div class="config-summary">
        <span>{{ roomStore.config.num_rounds }} rounds</span>
        <span>Cap: {{ roomStore.config.score_cap || 'None' }}</span>
        <span>{{ roomStore.config.open_call_mode === 'koukou' ? '口口翻' : '开口翻' }}</span>
        <span>{{ roomStore.config.turn_timer }}s turns</span>
      </div>

      <div class="actions">
        <button class="btn-primary" @click="toggleReady">
          Toggle Ready
        </button>
        <button
          v-if="isHost"
          class="btn-success"
          :disabled="!canStart"
          @click="startGame"
        >
          Start Game
        </button>
        <button class="btn-secondary" @click="handleLeave">
          Leave
        </button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '../styles/variables' as *;

.room {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: $spacing-md;
}

.room-card {
  background: $color-bg-light;
  border: 1px solid $color-border;
  border-radius: $border-radius;
  padding: $spacing-xl;
  width: 100%;
  max-width: 500px;
}

.room-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: $spacing-lg;

  h2 { font-size: 1.5rem; }
}

.code-display {
  background: $color-surface;
  padding: $spacing-sm $spacing-md;
  border-radius: $border-radius;
  font-family: monospace;
  font-size: 1.4rem;
  font-weight: 700;
  letter-spacing: 3px;
  cursor: pointer;
  user-select: all;

  &:hover {
    background: lighten($color-surface, 5%);
  }
}

.players {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
  margin-bottom: $spacing-lg;
}

.player-slot {
  display: flex;
  align-items: center;
  gap: $spacing-md;
  padding: $spacing-md;
  background: $color-bg;
  border: 1px solid $color-border;
  border-radius: $border-radius;

  &.you {
    border-color: $color-primary;
  }

  &.ready {
    border-color: $color-success;
  }
}

.seat-label {
  font-size: 0.8rem;
  color: $color-text-muted;
  width: 40px;
  text-transform: uppercase;
}

.player-name {
  flex: 1;
  font-weight: 600;
}

.you-tag {
  color: $color-primary;
  font-size: 0.8rem;
}

.host-tag {
  background: $color-warning;
  color: $color-bg;
  font-size: 0.65rem;
  padding: 2px 6px;
  border-radius: 3px;
  margin-left: $spacing-xs;
  vertical-align: middle;
}

.ready-status {
  font-size: 0.85rem;
  color: $color-text-muted;

  &.active {
    color: $color-success;
    font-weight: 600;
  }
}

.empty {
  color: $color-text-muted;
  font-style: italic;
}

.config-summary {
  display: flex;
  gap: $spacing-md;
  flex-wrap: wrap;
  margin-bottom: $spacing-lg;
  font-size: 0.85rem;
  color: $color-text-muted;

  span {
    background: $color-bg;
    padding: $spacing-xs $spacing-sm;
    border-radius: $border-radius-sm;
  }
}

.actions {
  display: flex;
  gap: $spacing-sm;

  button {
    flex: 1;
    padding: $spacing-md;
  }
}
</style>
