import { api } from '@/api/client'
import type { ApiResponse, AdminDashboardStats, User, Pagination, Post, GarbageReport, UserRole } from '@/types'
import type { ReviewAction, BanUserRequest, CreateQuestionRequest } from '@/types/admin'

export function getAdminDashboard() {
  return api.get<ApiResponse<AdminDashboardStats>>('/admin/dashboard')
}

export function getUsers(params?: { page?: number; page_size?: number; role?: UserRole; keyword?: string }) {
  return api.get<ApiResponse<{ users: User[]; pagination: Pagination }>>('/admin/users', params as Record<string, unknown>)
}

export function updateUserRole(userId: string, role: UserRole) {
  return api.put<ApiResponse<User>>(`/admin/users/${userId}/role`, { role })
}

export function banUser(userId: string, data: BanUserRequest) {
  return api.post(`/admin/users/${userId}/ban`, data)
}

export function getPendingPosts(params?: { page?: number; page_size?: number }) {
  return api.get<ApiResponse<{ posts: Post[]; pagination: Pagination }>>('/admin/posts/pending', params as Record<string, unknown>)
}

export function reviewPost(postId: string, data: ReviewAction) {
  return api.post<ApiResponse<Post>>(`/admin/posts/${postId}/review`, data)
}

export function getAllGarbageReports(params?: { page?: number; page_size?: number; status?: string }) {
  return api.get<ApiResponse<{ reports: GarbageReport[]; pagination: Pagination }>>('/admin/garbage/reports', params as Record<string, unknown>)
}

export function createQuestion(data: CreateQuestionRequest) {
  return api.post<ApiResponse<unknown>>('/admin/questions', data)
}
