import client from './client'
import type { ApiResponse, ListResponse, Post, Comment } from '@/types'

export const postsApi = {
  list(params: { page?: number; page_size?: number; topic?: string; keyword?: string; sort?: string }) {
    return client.get<ApiResponse<ListResponse<Post>>>('/posts', { params })
  },
  get(id: string) {
    return client.get<ApiResponse<Post>>(`/posts/${id}`)
  },
  create(data: { title: string; content: string; images?: string[]; tags?: string[]; topic?: string }) {
    return client.post<ApiResponse<Post>>('/posts', data)
  },
  update(id: string, data: { title?: string; content?: string; images?: string[]; tags?: string[]; topic?: string }) {
    return client.put<ApiResponse<Post>>(`/posts/${id}`, data)
  },
  delete(id: string) {
    return client.delete<ApiResponse>(`/posts/${id}`)
  },
  toggleLike(id: string) {
    return client.post<ApiResponse<{ liked: boolean }>>(`/posts/${id}/like`)
  },
  listComments(postId: string, page = 1, pageSize = 20) {
    return client.get<ApiResponse<ListResponse<Comment>>>(`/posts/${postId}/comments`, {
      params: { page, page_size: pageSize },
    })
  },
  createComment(postId: string, content: string) {
    return client.post<ApiResponse<Comment>>(`/posts/${postId}/comments`, { content })
  },
  deleteComment(commentId: string) {
    return client.delete<ApiResponse>(`/comments/${commentId}`)
  },
}
