import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(localStorage.getItem('token'))
  const username = ref<string | null>(localStorage.getItem('username'))

  const isAuthenticated = computed(() => !!token.value)

  async function login(usernameInput: string, password: string) {
    const res = await authApi.login(usernameInput, password)
    token.value = res.data.token
    username.value = usernameInput
    localStorage.setItem('token', res.data.token)
    localStorage.setItem('username', usernameInput)
  }

  function logout() {
    token.value = null
    username.value = null
    localStorage.removeItem('token')
    localStorage.removeItem('username')
  }

  async function checkAuthStatus() {
    if (!token.value) {
      return false
    }
    try {
      const res = await authApi.status()
      if (!res.data.authenticated) {
        logout()
        return false
      }
      return true
    } catch {
      logout()
      return false
    }
  }

  return {
    token,
    username,
    isAuthenticated,
    login,
    logout,
    checkAuthStatus
  }
})
