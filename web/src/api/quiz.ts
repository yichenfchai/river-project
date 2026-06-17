import client from './client'
import type { ApiResponse, QuizSession, SubmitResponse, LeaderboardEntry, UserStats } from '@/types'

export const quizApi = {
  startSession(difficulty = 'all', count = 10) {
    return client.get<ApiResponse<QuizSession>>('/quiz/questions', {
      params: { difficulty, count },
    })
  },
  submit(sessionId: string, questionId: string, selectedOption: number) {
    return client.post<ApiResponse<SubmitResponse>>('/quiz/submit', {
      session_id: sessionId,
      question_id: questionId,
      selected_option: selectedOption,
    })
  },
  getLeaderboard(period = 'total', limit = 20) {
    return client.get<ApiResponse<LeaderboardEntry[]>>('/leaderboard', {
      params: { period, limit },
    })
  },
  getUserStats(userId: string) {
    return client.get<ApiResponse<UserStats>>(`/quiz/users/${userId}/stats`)
  },
}
