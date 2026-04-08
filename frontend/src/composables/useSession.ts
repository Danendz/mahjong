import { ref, readonly } from 'vue'

const STORAGE_KEY_TOKEN = 'mahjong_session_token'
const STORAGE_KEY_USER_ID = 'mahjong_user_id'
const STORAGE_KEY_NICKNAME = 'mahjong_nickname'

const API_BASE = import.meta.env.VITE_API_URL || 'http://localhost:8080'

const sessionToken = ref(localStorage.getItem(STORAGE_KEY_TOKEN) || '')
const userId = ref(localStorage.getItem(STORAGE_KEY_USER_ID) || '')
const nickname = ref(localStorage.getItem(STORAGE_KEY_NICKNAME) || '')
const isAuthenticated = ref(!!sessionToken.value)

export function useSession() {
  async function createGuestSession(name: string): Promise<void> {
    const res = await fetch(`${API_BASE}/api/auth/guest`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ nickname: name }),
    })

    if (!res.ok) {
      throw new Error('Failed to create guest session')
    }

    const data = await res.json()
    sessionToken.value = data.session_token
    userId.value = data.user_id
    nickname.value = data.nickname

    localStorage.setItem(STORAGE_KEY_TOKEN, data.session_token)
    localStorage.setItem(STORAGE_KEY_USER_ID, data.user_id)
    localStorage.setItem(STORAGE_KEY_NICKNAME, data.nickname)
    isAuthenticated.value = true
  }

  function setNickname(name: string) {
    nickname.value = name
    localStorage.setItem(STORAGE_KEY_NICKNAME, name)
  }

  function clearSession() {
    sessionToken.value = ''
    userId.value = ''
    nickname.value = ''
    isAuthenticated.value = false
    localStorage.removeItem(STORAGE_KEY_TOKEN)
    localStorage.removeItem(STORAGE_KEY_USER_ID)
    localStorage.removeItem(STORAGE_KEY_NICKNAME)
  }

  return {
    sessionToken: readonly(sessionToken),
    userId: readonly(userId),
    nickname,
    isAuthenticated: readonly(isAuthenticated),
    createGuestSession,
    setNickname,
    clearSession,
  }
}
