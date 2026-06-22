import { api } from '@/api/client'
import type { ApiResponse, LoginRequest, LoginResponse, RegisterRequest, UserProfile } from '@/types'

export function login(data: LoginRequest) {
  return api.post<ApiResponse<LoginResponse>>('/auth/login', data)
}

export function register(data: RegisterRequest) {
  return api.post<ApiResponse<LoginResponse>>('/auth/register', data)
}

export function logout() {
  return api.post('/auth/logout')
}

export function getMyProfile() {
  return api.get<ApiResponse<UserProfile>>('/users/me')
}

export function updateProfile(data: Partial<Pick<UserProfile, 'nickname' | 'bio' | 'avatar_url'>>) {
  return api.put<ApiResponse<UserProfile>>('/users/me', data)
}
