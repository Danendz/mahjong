<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useI18n } from 'vue-i18n'
import { useSession } from '../composables/useSession'
import { useGameConnection } from '../composables/useGameConnection'
import LanguageSelector from '../components/LanguageSelector.vue'

const router = useRouter()
const route = useRoute()
const { t } = useI18n()
const session = useSession()
const { joinRoom } = useGameConnection()

const nickname = ref(session.nickname.value || '')
const roomCode = ref('')
const mode = ref<'menu' | 'join'>('menu')
const loading = ref(false)
const error = ref('')

const API_BASE = import.meta.env.VITE_API_URL || ''

onMounted(() => {
  const code = route.query.code
  if (typeof code === 'string' && code.trim()) {
    roomCode.value = code.trim().toUpperCase()
    mode.value = 'join'
  }
})

function handleBack() {
  mode.value = 'menu'
  roomCode.value = ''
  error.value = ''
  if (route.query.code) {
    router.replace({ path: '/', query: {} })
  }
}

async function ensureSession() {
  if (!session.isAuthenticated.value || !session.sessionToken.value) {
    if (!nickname.value.trim()) {
      error.value = t('lobby.errors.nicknameRequired')
      return false
    }
    await session.createGuestSession(nickname.value.trim())
  } else if (nickname.value.trim() !== session.nickname.value) {
    session.setNickname(nickname.value.trim())
  }
  return true
}

async function handleCreate() {
  error.value = ''
  loading.value = true

  try {
    if (!await ensureSession()) return

    const res = await fetch(`${API_BASE}/api/mahjong/rooms`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        session_token: session.sessionToken.value,
        nickname: nickname.value.trim(),
      }),
    })

    if (!res.ok) throw new Error(t('lobby.errors.createFailed'))
    const data = await res.json()

    await joinRoom(data.code, nickname.value.trim(), session.sessionToken.value)
    router.push({ name: 'room', params: { code: data.code } })
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}

async function handleJoin() {
  error.value = ''
  if (!roomCode.value.trim()) {
    error.value = t('lobby.errors.roomCodeRequired')
    return
  }
  loading.value = true

  try {
    if (!await ensureSession()) return

    const code = roomCode.value.trim().toUpperCase()

    // Validate room exists before attempting WebSocket join
    const res = await fetch(`${API_BASE}/api/mahjong/rooms/${code}`)
    if (res.status === 404) {
      error.value = t('lobby.errors.roomNotFound')
      return
    }
    if (!res.ok) {
      error.value = t('lobby.errors.checkFailed')
      return
    }

    await joinRoom(code, nickname.value.trim(), session.sessionToken.value)
    router.push({ name: 'room', params: { code } })
  } catch (e: any) {
    error.value = e.message
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="lobby">
    <div class="lang-wrapper">
      <LanguageSelector />
    </div>
    <div class="lobby-card">
      <h1 class="title">{{ $t('lobby.title') }}</h1>
      <p v-if="$t('lobby.subtitle')" class="subtitle">{{ $t('lobby.subtitle') }}</p>

      <div class="form">
        <input
          v-model="nickname"
          :placeholder="$t('lobby.nicknamePlaceholder')"
          maxlength="16"
          @keyup.enter="mode === 'join' ? handleJoin() : handleCreate()"
        />

        <div v-if="mode === 'menu'" class="actions">
          <button class="btn-primary" :disabled="loading || !nickname.trim()" @click="handleCreate">
            {{ $t('lobby.createRoom') }}
          </button>
          <button class="btn-secondary" :disabled="!nickname.trim()" @click="mode = 'join'">
            {{ $t('lobby.joinRoom') }}
          </button>
        </div>

        <div v-else class="actions">
          <input
            v-model="roomCode"
            :placeholder="$t('lobby.roomCodePlaceholder')"
            maxlength="6"
            class="code-input"
            @keyup.enter="handleJoin"
          />
          <button class="btn-primary" :disabled="loading || !roomCode.trim()" @click="handleJoin">
            {{ $t('lobby.join') }}
          </button>
          <button class="btn-secondary" @click="handleBack">
            {{ $t('common.back') }}
          </button>
        </div>

        <p v-if="error" class="error">{{ error }}</p>
      </div>
    </div>
  </div>
</template>

<style lang="scss" scoped>
@use '../styles/variables' as *;

.lobby {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  padding: $spacing-md;
  position: relative;
}

.lang-wrapper {
  position: absolute;
  top: $spacing-md;
  right: $spacing-md;
}

.lobby-card {
  background: $color-bg-light;
  border: 1px solid $color-border;
  border-radius: $border-radius;
  padding: $spacing-xl;
  width: 100%;
  max-width: 400px;
  text-align: center;
}

.title {
  font-size: 2.5rem;
  margin-bottom: $spacing-xs;
}

.subtitle {
  color: $color-text-muted;
  margin-bottom: $spacing-xl;
}

.form {
  display: flex;
  flex-direction: column;
  gap: $spacing-md;

  input {
    width: 100%;
    text-align: center;
    font-size: 1.1rem;
  }

  .code-input {
    text-transform: uppercase;
    letter-spacing: 4px;
    font-weight: 700;
  }
}

.actions {
  display: flex;
  flex-direction: column;
  gap: $spacing-sm;

  button {
    width: 100%;
    padding: $spacing-md;
  }
}

.error {
  color: $color-danger;
  font-size: 0.9rem;
}
</style>
