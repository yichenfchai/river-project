export type UserRole = 'user' | 'monitor' | 'admin' | 'guest'

export interface User {
  id: string
  username: string
  nickname: string
  email: string
  avatar_url: string
  role: UserRole
  created_at: string
}

export interface UserProfile extends User {
  bio: string
  points: number
  rank_title: string
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

export interface LoginRequest {
  username: string
  password: string
}

export interface RegisterRequest {
  username: string
  password: string
  email: string
  nickname?: string
  role?: UserRole
}
