import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import { authApi } from '@/api/auth'
import type { User } from '@/types'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(loadUser())
  const loading = ref(false)

  const isLoggedIn = computed(() => !!user.value && !!localStorage.getItem('access_token'))
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isMonitor = computed(() => user.value?.role === 'monitor' || user.value?.role === 'admin')

  function loadUser(): User | null {
    const raw = localStorage.getItem('user')
    return raw ? JSON.parse(raw) : null
  }

  async function login(username: string, password: string) {
    loading.value = true
    try {
      const res = await authApi.login({ username, password })
      const data = res.data.data!
      localStorage.setItem('access_token', data.access_token)
      localStorage.setItem('refresh_token', data.refresh_token)
      localStorage.setItem('user', JSON.stringify(data.user))
      user.value = data.user
      return true
    } catch {
      return false
    } finally {
      loading.value = false
    }
  }

  async function register(data: { username: string; password: string; email: string; nickname?: string }) {
    loading.value = true
    try {
      await authApi.register(data)
      return true
    } catch {
      return false
    } finally {
      loading.value = false
    }
  }

  async function fetchProfile() {
    try {
      const res = await authApi.getProfile()
      user.value = res.data.data!
      localStorage.setItem('user', JSON.stringify(user.value))
    } catch {
      logout()
    }
  }

  function logout() {
    authApi.logout().catch(() => {})
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('user')
    user.value = null
  }

  return { user, loading, isLoggedIn, isAdmin, isMonitor, login, register, fetchProfile, logout }
})
