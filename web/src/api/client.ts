import axios from 'axios'
import type { ApiResponse, TokenPair } from '@/types'

const client = axios.create({
  baseURL: '/api/v1',
  timeout: 15000,
})

// 请求拦截器: 自动带 Token
client.interceptors.request.use((config) => {
  const token = localStorage.getItem('access_token')
  if (token) {
    config.headers.Authorization = `Bearer ${token}`
  }
  return config
})

// 响应拦截器: 401 自动刷新 Token
let isRefreshing = false
let refreshQueue: Array<{ resolve: (token: string) => void; reject: (err: Error) => void }> = []

client.interceptors.response.use(
  (res) => res,
  async (error) => {
    const { config, response } = error
    if (response?.status !== 401 || config._retry) {
      return Promise.reject(error)
    }

    const refreshToken = localStorage.getItem('refresh_token')
    if (!refreshToken) {
      clearAuth()
      window.location.href = '/login'
      return Promise.reject(error)
    }

    if (isRefreshing) {
      return new Promise((resolve, reject) => {
        refreshQueue.push({ resolve, reject })
      }).then((token) => {
        config.headers.Authorization = `Bearer ${token}`
        return client(config)
      })
    }

    isRefreshing = true
    config._retry = true

    try {
      const res = await axios.post<ApiResponse<TokenPair>>('/api/v1/auth/refresh', {
        refresh_token: refreshToken,
      })
      const pair = res.data.data!
      localStorage.setItem('access_token', pair.access_token)
      localStorage.setItem('refresh_token', pair.refresh_token)

      refreshQueue.forEach(({ resolve }) => resolve(pair.access_token))
      refreshQueue = []

      config.headers.Authorization = `Bearer ${pair.access_token}`
      return client(config)
    } catch {
      refreshQueue.forEach(({ reject }) => reject(new Error('refresh failed')))
      refreshQueue = []
      clearAuth()
      window.location.href = '/login'
      return Promise.reject(error)
    } finally {
      isRefreshing = false
    }
  },
)

function clearAuth() {
  localStorage.removeItem('access_token')
  localStorage.removeItem('refresh_token')
  localStorage.removeItem('user')
}

export default client
