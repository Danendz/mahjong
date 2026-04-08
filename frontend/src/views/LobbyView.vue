<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { useSession } from '../composables/useSession'
import { useGameConnection } from '../composables/useGameConnection'

const router = useRouter()
const session = useSession()
const { joinRoom } = useGameConnection()

const nickname = ref(session.nickname.value || '')
const roomCode = ref('')
const mode = ref<'menu' | 'join'>('menu')
const loading = ref(false)
const error = ref('')

const API_BASE = import.meta.env.VITE_API_URL || ''

async function ensureSession() {
  if (!session.isAuthenticated.value || !session.sessionToken.value) {
    if (!nickname.value.trim()) {
      error.value = 'Please enter a nickname'
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

    const res = await fetch(`${API_BASE}/api/rooms`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        session_token: session.sessionToken.value,
        nickname: nickname.value.trim(),
      }),
    })

    if (!res.ok) throw new Error('Failed to create room')
    const data = await res.json()

    joinRoom(data.code, nickname.value.trim(), session.sessionToken.value)
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
    error.value = 'Please enter a room code'
    return
  }
  loading.value = true

  try {
    if (!await ensureSession()) return

    const code = roomCode.value.trim().toUpperCase()
    joinRoom(code, nickname.value.trim(), session.sessionToken.value)
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
    <div class="lobby-card">
      <h1 class="title">武汉麻将</h1>
      <p class="subtitle">Wuhan Mahjong</p>

      <div class="form">
        <input
          v-model="nickname"
          placeholder="Your nickname"
          maxlength="16"
          @keyup.enter="mode === 'join' ? handleJoin() : handleCreate()"
        />

        <div v-if="mode === 'menu'" class="actions">
          <button class="btn-primary" :disabled="loading || !nickname.trim()" @click="handleCreate">
            Create Room
          </button>
          <button class="btn-secondary" :disabled="!nickname.trim()" @click="mode = 'join'">
            Join Room
          </button>
        </div>

        <div v-else class="actions">
          <input
            v-model="roomCode"
            placeholder="Room code (e.g. H7KM3P)"
            maxlength="6"
            class="code-input"
            @keyup.enter="handleJoin"
          />
          <button class="btn-primary" :disabled="loading || !roomCode.trim()" @click="handleJoin">
            Join
          </button>
          <button class="btn-secondary" @click="mode = 'menu'">
            Back
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
