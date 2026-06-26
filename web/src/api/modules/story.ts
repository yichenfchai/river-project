import { api } from '@/api/client'
import type { ApiResponse, Pagination } from '@/types'

export interface Story {
  id: string
  title: string
  content: string
  topic: string
  emoji: string
  age_group: string
  likes: number
  created_at: string
}

export function getStories(params?: { page?: number; page_size?: number }) {
  return api.get<ApiResponse<{ stories: Story[]; pagination: Pagination }>>('/stories', params as Record<string, unknown>)
}

export function getStory(id: string) {
  return api.get<ApiResponse<Story>>(`/stories/${id}`)
}
