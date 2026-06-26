import { api } from '@/api/client'
import type { ApiResponse, Pagination } from '@/types'
import type { SessionResponse, BatchSubmitResult, UserStats, LeaderboardResponse, SubmitAnswerResponse } from '@/types/quiz'

export function getQuestions(params?: { count?: number; difficulty?: string; category?: string }) {
  return api.get<ApiResponse<SessionResponse>>('/quiz/questions', params as Record<string, unknown>)
}

export function submitAnswer(data: { session_id: string; question_id: string; answer: string; answer_time_ms?: number }) {
  return api.post<ApiResponse<SubmitAnswerResponse>>('/quiz/submit', data)
}

export function submitBatch(data: { session_id: string; answers: { question_id: string; answer: string; answer_time_ms?: number }[] }) {
  return api.post<ApiResponse<BatchSubmitResult>>('/quiz/submit-batch', data)
}

export function getUserStats(userId: string) {
  return api.get<ApiResponse<UserStats>>(`/quiz/users/${userId}/stats`)
}

export function getLeaderboard(params?: { period?: string; page?: number; page_size?: number }) {
  return api.get<ApiResponse<LeaderboardResponse>>('/leaderboard', params as Record<string, unknown>)
}
