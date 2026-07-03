import type { User } from './user'

export interface Question {
  id: string
  type: string
  difficulty: string
  category: string
  question: string
  options: string[]
}

export interface SessionResponse {
  session_id: string
  questions: Question[]
}

export interface SubmitAnswerResponse {
  is_correct: boolean
  correct_answer: string
  explanation: string
  points_earned: number
  streak_bonus: number
  total_points: number
}

export interface BatchSubmitResult {
  correct_count: number
  total_count: number
  total_points_earned: number
  new_total_points: number
  rank_title: string
  results: SubmitAnswerResponse[]
}

export interface UserStats {
  total_points: number
  rank_title: string
  total_answers: number
  correct_answers: number
  accuracy: number
  current_streak: number
  max_streak: number
  category_breakdown: Record<string, { total: number; correct: number }>
}

export interface LeaderboardEntry {
  rank: number
  user: User
  total_points: number
  rank_title: string
  answer_count: number
  accuracy: number
}

export interface MyRank {
  rank: number
  total_points: number
}

export interface LeaderboardResponse {
  period: string
  leaderboard: LeaderboardEntry[]
  my_rank: MyRank | null
}
