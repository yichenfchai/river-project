import axios, { type AxiosInstance, type AxiosError, type InternalAxiosRequestConfig } from 'axios'
import { ElMessage } from 'element-plus'
import type { ApiResponse } from '@/types'

const TOKEN_REFRESH_PATH = '/api/v1/auth/refresh'

class ApiClient {
  private instance: AxiosInstance
  private isRefreshing = false
  private refreshSubscribers: ((token: string) => void)[] = []

  constructor() {
    const baseURL = import.meta.env.VITE_API_BASE_URL || '/api/v1'

    this.instance = axios.create({
      baseURL,
      timeout: 30000,
      headers: { 'Content-Type': 'application/json' },
    })

    this.setupInterceptors()
  }

  private getAccessToken(): string | null {
    return localStorage.getItem('access_token')
  }

  private getRefreshToken(): string | null {
    return localStorage.getItem('refresh_token')
  }

  private setTokens(access: string, refresh: string) {
    localStorage.setItem('access_token', access)
    localStorage.setItem('refresh_token', refresh)
  }

  private clearTokens() {
    localStorage.removeItem('access_token')
    localStorage.removeItem('refresh_token')
  }

  private onRefreshed(token: string) {
    this.refreshSubscribers.forEach((cb) => cb(token))
    this.refreshSubscribers = []
  }

  private addRefreshSubscriber(cb: (token: string) => void) {
    this.refreshSubscribers.push(cb)
  }

  private async refreshAccessToken(): Promise<string | null> {
    const refreshToken = this.getRefreshToken()
    if (!refreshToken) return null

    try {
      const { data } = await axios.post<ApiResponse<{ access_token: string; refresh_token: string }>>(
        `${this.instance.defaults.baseURL}${TOKEN_REFRESH_PATH}`,
        { refresh_token: refreshToken },
        { headers: { 'Content-Type': 'application/json' } },
      )

      const { access_token, refresh_token } = data.data
      this.setTokens(access_token, refresh_token)
      return access_token
    } catch {
      this.clearTokens()
      return null
    }
  }

  private setupInterceptors() {
    this.instance.interceptors.request.use(
      (config: InternalAxiosRequestConfig) => {
        const token = this.getAccessToken()
        if (token && config.headers) {
          config.headers.Authorization = `Bearer ${token}`
        }
        return config
      },
      (error) => Promise.reject(error),
    )

    this.instance.interceptors.response.use(
      (response) => response,
      async (error: AxiosError<ApiResponse>) => {
        const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean }
        if (!originalRequest) return Promise.reject(error)

        if (error.response?.status !== 401 || originalRequest._retry) {
          this.handleError(error)
          return Promise.reject(error)
        }

        if (this.isRefreshing) {
          return new Promise((resolve) => {
            this.addRefreshSubscriber((token: string) => {
              originalRequest.headers.Authorization = `Bearer ${token}`
              resolve(this.instance(originalRequest))
            })
          })
        }

        originalRequest._retry = true
        this.isRefreshing = true

        const newToken = await this.refreshAccessToken()

        if (!newToken) {
          this.isRefreshing = false
          this.clearTokens()
          window.location.href = '/login'
          return Promise.reject(error)
        }

        this.isRefreshing = false
        this.onRefreshed(newToken)
        originalRequest.headers.Authorization = `Bearer ${newToken}`
        return this.instance(originalRequest)
      },
    )
  }

  private handleError(error: AxiosError<ApiResponse>) {
    const status = error.response?.status
    const message = error.response?.data?.message

    switch (status) {
      case 403:
        ElMessage.error('无权执行此操作')
        break
      case 404:
        ElMessage.error(message || '资源不存在')
        break
      case 429:
        ElMessage.warning('请求过于频繁，请稍后再试')
        break
      case 500:
        ElMessage.error('服务器内部错误，请稍后重试')
        break
      default:
        if (message) ElMessage.error(message)
    }
  }

  async get<T>(url: string, params?: Record<string, unknown>): Promise<T> {
    const { data } = await this.instance.get<T>(url, { params })
    return data
  }

  async post<T>(url: string, body?: unknown): Promise<T> {
    const { data } = await this.instance.post<T>(url, body)
    return data
  }

  async put<T>(url: string, body?: unknown): Promise<T> {
    const { data } = await this.instance.put<T>(url, body)
    return data
  }

  async delete<T = void>(url: string): Promise<T> {
    const { data } = await this.instance.delete<T>(url)
    return data
  }

  async upload<T>(url: string, formData: FormData): Promise<T> {
    const { data } = await this.instance.post<T>(url, formData, {
      headers: { 'Content-Type': 'multipart/form-data' },
    })
    return data
  }
}

export const api = new ApiClient()
