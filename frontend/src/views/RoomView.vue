<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useRouter } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useRoomStore } from '../stores/room'
import { useUserStore } from '../stores/user'
import { useGameConnection } from '../composables/useGameConnection'
import type { BotDifficulty, RoomConfig } from '../types/generated'

const props = defineProps<{ code: string }>()
const router = useRouter()
const { t } = useI18n()
const roomStore = useRoomStore()
const userStore = useUserStore()
const { toggleReady, startGame, leaveRoom, addBot, removeBot, setBotDifficulty, configureRoom } = useGameConnection()

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

const seatKeys = ['east', 'south', 'west', 'north'] as const
const seatLabels = computed(() => seatKeys.map(k => t(`room.seats.${k}`)))
const difficultyLabel = (d?: string) => (d ? t(`room.difficulty.${d}`) : '')

watch(() => roomStore.status, (status) => {
  if (status === 'playing') {
    router.push({ name: 'game', params: { code: props.code } })
  }
}, { immediate: true })

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

// Settings panel
const showSettings = ref(false)
const editConfig = ref<RoomConfig>({ ...roomStore.config })

watch(() => roomStore.config, (cfg) => {
  editConfig.value = { ...cfg }
}, { deep: true })

function setConfigValue<K extends keyof RoomConfig>(key: K, value: RoomConfig[K]) {
  editConfig.value[key] = value
  configureRoom(editConfig.value)
}
</script>

<template>
  <div class="room">
    <div class="room-card">
      <div class="room-header">
        <h2>{{ $t('room.title') }}</h2>
        <div class="code-area">
          <div class="code-display" @click="copyCode" :title="$t('room.copyInvite')">
            {{ code }}
            <svg class="copy-icon" xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="9" y="9" width="13" height="13" rx="2" ry="2"/><path d="M5 15H4a2 2 0 0 1-2-2V4a2 2 0 0 1 2-2h9a2 2 0 0 1 2 2v1"/></svg>
          </div>
          <Transition name="fade">
            <span v-if="copied" class="copied-tooltip">{{ $t('room.copied') }}</span>
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
              <span class="bot-tag">{{ $t('room.botTag') }}</span>
            </div>
            <div class="bot-controls">
              <select
                v-if="isHost"
                class="difficulty-select"
                :value="roomStore.players.find(p => p.seat === seatIdx - 1)?.difficulty"
                @change="handleDifficultyChange(seatIdx - 1, $event)"
              >
                <option value="easy">{{ $t('room.difficulty.easy') }}</option>
                <option value="medium">{{ $t('room.difficulty.medium') }}</option>
                <option value="hard">{{ $t('room.difficulty.hard') }}</option>
              </select>
              <span v-else class="difficulty-label">
                {{ difficultyLabel(roomStore.players.find(p => p.seat === seatIdx - 1)?.difficulty) }}
              </span>
              <button v-if="isHost" class="btn-remove" @click="handleRemoveBot(seatIdx - 1)">{{ $t('room.removeBot') }}</button>
            </div>
          </template>

          <!-- Human player -->
          <template v-else-if="roomStore.players.find(p => p.seat === seatIdx - 1)">
            <div class="player-name">
              {{ roomStore.players.find(p => p.seat === seatIdx - 1)!.nickname }}
              <span v-if="seatIdx - 1 === userStore.seat" class="you-tag">({{ $t('common.you') }})</span>
              <span v-if="seatIdx - 1 === 0" class="host-tag">{{ $t('room.hostTag') }}</span>
            </div>
            <div class="ready-status" :class="{ active: roomStore.players.find(p => p.seat === seatIdx - 1)?.ready }">
              {{ roomStore.players.find(p => p.seat === seatIdx - 1)?.ready ? $t('room.ready') : $t('room.notReady') }}
            </div>
          </template>

          <!-- Empty seat -->
          <template v-else>
            <div class="empty">{{ $t('room.waiting') }}</div>
            <button
              v-if="isHost && seatIdx - 1 !== 0"
              class="btn-add-bot"
              @click="handleAddBot(seatIdx - 1)"
            >
              {{ $t('room.addBot') }}
            </button>
          </template>
        </div>
      </div>

      <!-- Non-host: read-only summary -->
      <div v-if="!isHost" class="config-summary">
        <span>{{ $t('room.summary.rounds', { count: roomStore.config.num_rounds }) }}</span>
        <span>{{ $t('room.summary.cap', { value: roomStore.config.score_cap || $t('room.settings.scoreCapNone') }) }}</span>
        <span>{{ roomStore.config.open_call_mode === 'koukou' ? $t('room.settings.callModeKoukou') : $t('room.settings.callModeKaikou') }}</span>
        <span>{{ $t('room.summary.turnTimer', { seconds: roomStore.config.turn_timer }) }}</span>
        <span v-if="roomStore.config.zimo_only">{{ $t('room.summary.zimoOnly') }}</span>
        <span v-if="roomStore.config.dealer_continuation">{{ $t('room.summary.dealerCont') }}</span>
      </div>

      <!-- Host: expandable settings panel -->
      <div v-else class="settings-section">
        <div class="settings-header" @click="showSettings = !showSettings">
          <span class="settings-label">{{ $t('room.settings.title') }}</span>
          <svg class="gear-icon" :class="{ open: showSettings }" xmlns="http://www.w3.org/2000/svg" width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
            <circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 0 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 0 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 0 1-2.83-2.83l.06-.06A1.65 1.65 0 0 0 4.68 15a1.65 1.65 0 0 0-1.51-1H3a2 2 0 0 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 0 1 2.83-2.83l.06.06A1.65 1.65 0 0 0 9 4.68a1.65 1.65 0 0 0 1-1.51V3a2 2 0 0 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 0 1 2.83 2.83l-.06.06A1.65 1.65 0 0 0 19.4 9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 0 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>
          </svg>
        </div>

        <div v-if="showSettings" class="settings-panel">
          <div class="setting-row">
            <label>{{ $t('room.settings.rounds') }}</label>
            <div class="segmented">
              <button v-for="v in [4, 8, 16]" :key="v"
                :class="{ active: editConfig.num_rounds === v }"
                @click="setConfigValue('num_rounds', v as 4 | 8 | 16)">{{ v }}</button>
            </div>
          </div>

          <div class="setting-row">
            <label>{{ $t('room.settings.scoreCap') }}</label>
            <div class="segmented">
              <button v-for="v in [200, 500, 1000, 0]" :key="v"
                :class="{ active: editConfig.score_cap === v }"
                @click="setConfigValue('score_cap', v as 200 | 500 | 1000 | 0)">{{ v === 0 ? $t('room.settings.scoreCapNone') : v }}</button>
            </div>
          </div>

          <div class="setting-row">
            <label>{{ $t('room.settings.callMode') }}</label>
            <div class="segmented">
              <button :class="{ active: editConfig.open_call_mode === 'koukou' }"
                :title="$t('room.settings.callModeKoukouTooltip')"
                @click="setConfigValue('open_call_mode', 'koukou')">{{ $t('room.settings.callModeKoukou') }}</button>
              <button :class="{ active: editConfig.open_call_mode === 'kaikou' }"
                :title="$t('room.settings.callModeKaikouTooltip')"
                @click="setConfigValue('open_call_mode', 'kaikou')">{{ $t('room.settings.callModeKaikou') }}</button>
            </div>
          </div>

          <div class="setting-row">
            <label>{{ $t('room.settings.turnTimer') }}</label>
            <div class="segmented">
              <button v-for="v in [10, 15, 20, 30]" :key="v"
                :class="{ active: editConfig.turn_timer === v }"
                @click="setConfigValue('turn_timer', v as 10 | 15 | 20 | 30)">{{ v }}{{ $t('room.settings.secondsSuffix') }}</button>
            </div>
          </div>

          <div class="setting-row">
            <label>{{ $t('room.settings.reactionTimer') }}</label>
            <div class="segmented">
              <button v-for="v in [5, 8, 10, 15]" :key="v"
                :class="{ active: editConfig.reaction_timer === v }"
                @click="setConfigValue('reaction_timer', v as 5 | 8 | 10 | 15)">{{ v }}{{ $t('room.settings.secondsSuffix') }}</button>
            </div>
          </div>

          <div class="setting-row">
            <label>{{ $t('room.settings.zimoOnly') }}</label>
            <button class="toggle" :class="{ on: editConfig.zimo_only }"
              @click="setConfigValue('zimo_only', !editConfig.zimo_only)">
              <span class="toggle-knob" />
            </button>
          </div>

          <div class="setting-row">
            <label>{{ $t('room.settings.dealerContinuation') }}</label>
            <button class="toggle" :class="{ on: editConfig.dealer_continuation }"
              @click="setConfigValue('dealer_continuation', !editConfig.dealer_continuation)">
              <span class="toggle-knob" />
            </button>
          </div>
        </div>
      </div>

      <div class="actions">
        <button class="btn-primary" @click="toggleReady">
          {{ $t('room.toggleReady') }}
        </button>
        <button
          v-if="isHost"
          class="btn-success"
          :disabled="!canStart"
          @click="startGame"
        >
          {{ $t('room.startGame') }}
        </button>
        <button class="btn-secondary" @click="handleLeave">
          {{ $t('room.leave') }}
        </button>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use 'sass:color';
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
    background: color.adjust($color-surface, $lightness: 5%);
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

.settings-section {
  margin-bottom: $spacing-lg;
}

.settings-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: $spacing-sm $spacing-md;
  background: $color-bg;
  border: 1px solid $color-border;
  border-radius: $border-radius;
  cursor: pointer;
  user-select: none;

  &:hover {
    border-color: $color-primary;
  }
}

.settings-label {
  font-size: 0.9rem;
  font-weight: 600;
}

.gear-icon {
  opacity: 0.6;
  transition: transform 0.3s ease, opacity 0.2s;

  &.open {
    transform: rotate(90deg);
    opacity: 1;
  }

  .settings-header:hover & {
    opacity: 1;
  }
}

.settings-panel {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;
  padding: $spacing-md;
  margin-top: $spacing-xs;
  background: $color-bg;
  border: 1px solid $color-border;
  border-radius: $border-radius;
}

.setting-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: $spacing-md;

  label {
    font-size: 0.8rem;
    color: $color-text-muted;
    white-space: nowrap;
    min-width: 100px;
  }
}

.segmented {
  display: flex;

  button {
    background: $color-surface;
    color: $color-text-muted;
    border: 1px solid $color-border;
    padding: $spacing-xs $spacing-sm;
    font-size: 0.75rem;
    cursor: pointer;
    transition: background 0.15s, color 0.15s;

    &:first-child {
      border-radius: $border-radius-sm 0 0 $border-radius-sm;
    }

    &:last-child {
      border-radius: 0 $border-radius-sm $border-radius-sm 0;
    }

    &:not(:first-child) {
      border-left: none;
    }

    &.active {
      background: $color-primary;
      color: $color-text;
      border-color: $color-primary;

      + button {
        border-left-color: $color-primary;
      }
    }

    &:hover:not(.active) {
      background: color.adjust($color-surface, $lightness: 5%);
      color: $color-text;
    }
  }
}

.toggle {
  position: relative;
  width: 40px;
  height: 22px;
  background: $color-surface;
  border: 1px solid $color-border;
  border-radius: 11px;
  cursor: pointer;
  transition: background 0.2s, border-color 0.2s;
  padding: 0;
  flex-shrink: 0;

  &.on {
    background: $color-success;
    border-color: $color-success;

    .toggle-knob {
      transform: translateX(18px);
    }
  }
}

.toggle-knob {
  display: block;
  width: 16px;
  height: 16px;
  background: $color-text;
  border-radius: 50%;
  position: absolute;
  top: 2px;
  left: 2px;
  transition: transform 0.2s ease;
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
