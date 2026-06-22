import { ref, computed } from 'vue'
import { defineStore } from 'pinia'
import type { User, UserRole } from '@/types'
import { login as loginApi, logout as logoutApi, getMyProfile } from '@/api/modules/auth'

export const useAuthStore = defineStore('auth', () => {
  const user = ref<User | null>(null)
  const token = ref<string | null>(localStorage.getItem('access_token'))
  const loading = ref(false)

  const isLoggedIn = computed(() => !!token.value)
  const isAdmin = computed(() => user.value?.role === 'admin')
  const isMonitor = computed(() => user.value?.role === 'monitor')
  const isGuest = computed(() => user.value?.role === 'guest')
  const isUser = computed(() => user.value?.role === 'user')

  function parseJwtPayload(tokenStr: string): { sub?: string; role?: UserRole; exp?: number; iat?: number } | null {
    try {
      const parts = tokenStr.split('.')
      if (parts.length !== 3) return null
      const payload = JSON.parse(atob(parts[1]!.replace(/-/g, '+').replace(/_/g, '/')))
      return payload
    } catch {
      return null
    }
  }

  async function login(username: string, password: string) {
    loading.value = true
    try {
      const res = await loginApi({ username, password })
      localStorage.setItem('access_token', res.data.access_token)
      localStorage.setItem('refresh_token', res.data.refresh_token)
      user.value = res.data.user
      token.value = res.data.access_token
      return res.data.user
    } finally {
      loading.value = false
    }
  }

  async function fetchProfile() {
    try {
      const res = await getMyProfile()
      user.value = res.data as unknown as User
    } catch {
      // Silently fail, user might still be valid
    }
  }

  function loginAsGuest() {
    const guestUser: User = {
      id: 'guest',
      username: '游客',
      nickname: '游客',
      email: '',
      avatar_url: '',
      role: 'guest',
      created_at: '',
    }
    user.value = guestUser
    token.value = 'guest_token'
    localStorage.setItem('access_token', 'guest_token')
    localStorage.setItem('guest_user', JSON.stringify(guestUser))
  }

  function logout() {
    const refreshToken = localStorage.getItem('refresh_token')
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
    localStorage.removeItem('guest_user')
    user.value = null
    token.value = null
    if (refreshToken && refreshToken !== 'guest_token') {
      logoutApi().catch(() => {})
    }
  }

  function restoreSession() {
    const storedToken = localStorage.getItem('access_token')
    if (!storedToken) return false

    if (storedToken === 'guest_token') {
      const guestData = localStorage.getItem('guest_user')
      if (guestData) {
        try {
          user.value = JSON.parse(guestData) as User
          token.value = 'guest_token'
          return true
        } catch {
          logout()
          return false
        }
      }
      logout()
      return false
    }

    const payload = parseJwtPayload(storedToken)
    if (!payload) {
      logout()
      return false
    }

    const now = Math.floor(Date.now() / 1000)
    if (payload.exp && payload.exp < now) {
      logout()
      return false
    }

    token.value = storedToken
    fetchProfile()
    return true
  }

  return {
    user,
    token,
    loading,
    isLoggedIn,
    isAdmin,
    isMonitor,
    isUser,
    isGuest,
    login,
    loginAsGuest,
    logout,
    fetchProfile,
    restoreSession,
  }
})
