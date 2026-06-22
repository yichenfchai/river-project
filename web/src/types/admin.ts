export interface AdminDashboardStats {
  total_users: number
  active_today: number
  total_posts: number
  pending_reviews: number
  quiz_players: number
  garbage_reports: number
}

export interface ReviewAction {
  action: 'approve' | 'reject'
  reason?: string
}

export interface BanUserRequest {
  banned: boolean
  reason?: string
}

export interface CreateQuestionRequest {
  question: string
  options: string[]
  answer: string
  explanation?: string
  difficulty: 'easy' | 'medium' | 'hard'
  category: 'history' | 'ecology' | 'culture' | 'geography' | 'water_conservancy'
}
