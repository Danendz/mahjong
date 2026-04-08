import { defineStore } from 'pinia'
import { ref } from 'vue'

export const useUserStore = defineStore('user', () => {
  const nickname = ref('')
  const sessionToken = ref('')
  const userId = ref('')
  const seat = ref(-1)

  function setUser(nick: string, token: string, id: string) {
    nickname.value = nick
    sessionToken.value = token
    userId.value = id
  }

  function setSeat(s: number) {
    seat.value = s
  }

  function $reset() {
    nickname.value = ''
    sessionToken.value = ''
    userId.value = ''
    seat.value = -1
  }

  return { nickname, sessionToken, userId, seat, setUser, setSeat, $reset }
})
