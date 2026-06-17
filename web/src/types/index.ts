// 统一响应体
export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data?: T
}

// 分页
export interface PageData {
  page: number
  page_size: number
  total: number
  total_pages: number
}

export interface ListResponse<T> {
  items: T[]
  pagination: PageData
}

// 用户
export interface User {
  id: string
  username: string
  nickname: string
  email: string
  avatar_url: string
  role: 'user' | 'monitor' | 'admin'
  bio: string
  points: number
  rank_title: string
  created_at: string
}

export interface TokenPair {
  access_token: string
  refresh_token: string
  token_type: string
  expires_in: number
}

export interface LoginResponse extends TokenPair {
  user: User
}

// 帖子
export interface Post {
  id: string
  author_id: string
  title: string
  content: string
  images: string[]
  tags: string[]
  topic: string
  like_count: number
  comment_count: number
  status: string
  created_at: string
}

export interface Comment {
  id: string
  post_id: string
  author_id: string
  content: string
  created_at: string
}

// 问答
export interface Question {
  question_id: string
  content: string
  options: string[]
  difficulty: string
  category: string
}

export interface QuizSession {
  session_id: string
  questions: Question[]
}

export interface SubmitResponse {
  correct: boolean
  score: number
  total_score: number
  streak: number
  total_index: number
  total_questions: number
  has_more: boolean
}

export interface LeaderboardEntry {
  user_id: string
  nickname: string
  score: number
  rank: number
}

export interface UserStats {
  user_id: string
  total_score: number
  total_questions: number
  correct_count: number
  accuracy: number
  streak: number
  max_streak: number
  rank_title: string
}
