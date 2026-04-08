<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useRoomStore } from '../stores/room'
import { useUserStore } from '../stores/user'
import { useGameConnection } from '../composables/useGameConnection'
import type { BotDifficulty } from '../types/generated'

const props = defineProps<{ code: string }>()
const router = useRouter()
const roomStore = useRoomStore()
const userStore = useUserStore()
const { toggleReady, startGame, leaveRoom, addBot, removeBot, setBotDifficulty } = useGameConnection()

const isHost = computed(() => userStore.seat === 0)
const canStart = computed(() => roomStore.allReady && roomStore.playerCount === 4)

function handleAddBot(seat: number) {
  addBot(seat)
}

function handleRemoveBot(seat: number) {
  removeBot(seat)
}

function handleDifficultyChange(seat: number, event: Event) {
  const difficulty = (event.target as HTMLSelectElement).value as BotDifficulty
  setBotDifficulty(seat, difficulty)
}

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

const copied = ref(false)

function copyCode() {
  navigator.clipboard.writeText(window.location.origin + '/room/' + props.code)
  copied.value = true
  setTimeout(() => { copied.value = false }, 1500)
}
</script>

<template>
  <div class="room">
    <div class="room-card">
      <div class="room-header">
        <h2>Room</h2>
        <div class="code-area">
          <div class="code-display" @click="copyCode" title="Click to copy invite link">
            {{ code }}
            <svg class="copy-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
          </div>
          <Transition name="fade">
            <span v-if="copied" class="copied-tooltip">Copied!</span>
          </Transition>
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
            bot: roomStore.players.find(p => p.seat === seatIdx - 1)?.is_bot,
          }"
        >
          <div class="seat-label">{{ seatLabels[seatIdx - 1] }}</div>

          <!-- Bot player -->
          <template v-if="roomStore.players.find(p => p.seat === seatIdx - 1)?.is_bot">
            <div class="player-name">
              {{ roomStore.players.find(p => p.seat === seatIdx - 1)!.nickname }}
              <span class="bot-tag">BOT</span>
            </div>
            <div class="bot-controls">
              <select
                v-if="isHost"
                class="difficulty-select"
                :value="roomStore.players.find(p => p.seat === seatIdx - 1)?.difficulty"
                @change="handleDifficultyChange(seatIdx - 1, $event)"
              >
                <option value="easy">Easy</option>
                <option value="medium">Medium</option>
                <option value="hard">Hard</option>
              </select>
              <span v-else class="difficulty-label">
                {{ roomStore.players.find(p => p.seat === seatIdx - 1)?.difficulty }}
              </span>
              <button v-if="isHost" class="btn-remove" @click="handleRemoveBot(seatIdx - 1)">Remove</button>
            </div>
          </template>

          <!-- Human player -->
          <template v-else-if="roomStore.players.find(p => p.seat === seatIdx - 1)">
            <div class="player-name">
              {{ roomStore.players.find(p => p.seat === seatIdx - 1)!.nickname }}
              <span v-if="seatIdx - 1 === userStore.seat" class="you-tag">(you)</span>
              <span v-if="seatIdx - 1 === 0" class="host-tag">HOST</span>
            </div>
            <div class="ready-status" :class="{ active: roomStore.players.find(p => p.seat === seatIdx - 1)?.ready }">
              {{ roomStore.players.find(p => p.seat === seatIdx - 1)?.ready ? 'Ready' : 'Not ready' }}
            </div>
          </template>

          <!-- Empty seat -->
          <template v-else>
            <div class="empty">Waiting...</div>
            <button
              v-if="isHost && seatIdx - 1 !== 0"
              class="btn-add-bot"
              @click="handleAddBot(seatIdx - 1)"
            >
              + Add Bot
            </button>
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

.code-area {
  position: relative;
}

.code-display {
  display: inline-flex;
  align-items: center;
  gap: $spacing-sm;
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

.copy-icon {
  opacity: 0.6;
  vertical-align: middle;
  flex-shrink: 0;

  .code-display:hover & {
    opacity: 1;
  }
}

.copied-tooltip {
  position: absolute;
  top: 100%;
  right: 0;
  margin-top: $spacing-xs;
  font-size: 0.8rem;
  color: $color-success;
  font-weight: 600;
  white-space: nowrap;
}

.fade-leave-active {
  transition: opacity 0.3s ease;
}

.fade-leave-to {
  opacity: 0;
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

.bot-tag {
  background: #2dd4bf;
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
  flex: 1;
}

.bot-controls {
  display: flex;
  align-items: center;
  gap: $spacing-sm;
  margin-left: auto;
}

.difficulty-select {
  background: $color-surface;
  color: $color-text;
  border: 1px solid $color-border;
  border-radius: $border-radius-sm;
  padding: 2px $spacing-sm;
  font-size: 0.8rem;
  cursor: pointer;

  &:focus {
    outline: none;
    border-color: $color-primary;
  }
}

.difficulty-label {
  font-size: 0.8rem;
  color: $color-text-muted;
  text-transform: capitalize;
}

.btn-remove {
  background: transparent;
  color: $color-danger;
  border: 1px solid $color-danger;
  border-radius: $border-radius-sm;
  padding: 2px $spacing-sm;
  font-size: 0.75rem;
  cursor: pointer;

  &:hover {
    background: $color-danger;
    color: $color-text;
  }
}

.btn-add-bot {
  background: transparent;
  color: #2dd4bf;
  border: 1px solid #2dd4bf;
  border-radius: $border-radius-sm;
  padding: 2px $spacing-md;
  font-size: 0.8rem;
  cursor: pointer;
  margin-left: auto;

  &:hover {
    background: #2dd4bf;
    color: $color-bg;
  }
}

.player-slot.bot {
  border-color: #2dd4bf;
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
