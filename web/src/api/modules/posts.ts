import { api } from '@/api/client'
import type { ApiResponse, Post, Comment, Pagination } from '@/types'

export function getPosts(params?: { page?: number; page_size?: number; topic?: string; keyword?: string; sort?: string }) {
  return api.get<ApiResponse<{ posts: Post[]; pagination: Pagination }>>('/posts', params as Record<string, unknown>)
}

export function getPost(id: string) {
  return api.get<ApiResponse<Post>>(`/posts/${id}`)
}

export function createPost(data: { title: string; content: string; topic?: string; tags?: string[] }) {
  return api.post<ApiResponse<Post>>('/posts', data)
}

export function toggleLike(postId: string) {
  return api.post<ApiResponse<{ is_liked: boolean; like_count: number }>>(`/posts/${postId}/like`)
}

export function getComments(postId: string, params?: { page?: number; page_size?: number }) {
  return api.get<ApiResponse<{ comments: Comment[]; pagination: Pagination }>>(`/posts/${postId}/comments`, params as Record<string, unknown>)
}

export function createComment(postId: string, data: { content: string; parent_id?: string }) {
  return api.post<ApiResponse<Comment>>(`/posts/${postId}/comments`, data)
}
