import type { User } from './user'

export type PostTopic = 'share' | 'ecology' | 'culture' | 'question' | 'other'
export type PostStatus = 'pending' | 'approved' | 'rejected'

export interface Post {
  id: string
  author: User
  title: string
  content: string
  images: string[]
  tags: string[]
  topic: PostTopic
  like_count: number
  comment_count: number
  is_liked: boolean
  status: PostStatus
  created_at: string
  updated_at: string
}

export interface Comment {
  id: string
  post_id: string
  author: User
  content: string
  parent_id: string | null
  reply_to: User | null
  created_at: string
}

export interface CreatePostRequest {
  title: string
  content: string
  images?: File[]
  tags?: string[]
  topic?: PostTopic
}
