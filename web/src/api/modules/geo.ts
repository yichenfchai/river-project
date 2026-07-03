import { api } from '@/api/client'
import type { ApiResponse } from '@/types'

export interface CanalIntro {
  is_canal_city: boolean
  city: string
  province: string
  ip: string
  intro: {
    title: string
    content: string
    era: string
    poi: string
  }
}

export function getCanalIntro() {
  return api.get<ApiResponse<CanalIntro>>('/geo/canal-intro')
}

export function getCanalIntroByCoords(lat: number, lng: number) {
  return api.get<ApiResponse<CanalIntro>>('/geo/coords', { lat, lng } as Record<string, unknown>)
}
