import client from './client'
import type { ApiResponse, TokenPair, LoginResponse, User } from '@/types'

export const authApi = {
  register(data: { username: string; password: string; email: string; nickname?: string }) {
    return client.post<ApiResponse<User>>('/auth/register', data)
  },
  login(data: { username: string; password: string; device_id?: string }) {
    return client.post<ApiResponse<LoginResponse>>('/auth/login', data)
  },
  refresh(refreshToken: string) {
    return client.post<ApiResponse<TokenPair>>('/auth/refresh', { refresh_token: refreshToken })
  },
  logout() {
    return client.post<ApiResponse>('/users/logout')
  },
  getProfile() {
    return client.get<ApiResponse<User>>('/users/me')
  },
  updateProfile(data: { nickname?: string; avatar_url?: string; bio?: string }) {
    return client.put<ApiResponse<User>>('/users/me', data)
  },
  getUser(id: string) {
    return client.get<ApiResponse<User>>(`/users/${id}`)
  },
}
