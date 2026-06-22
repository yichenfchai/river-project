import { api } from '@/api/client'
import type { ApiResponse, VisionClassifyResponse, GarbageReport, Pagination } from '@/types'

export function classifyGarbage(image: File, lat?: number, lng?: number) {
  const formData = new FormData()
  formData.append('image', image)
  if (lat !== undefined) formData.append('lat', String(lat))
  if (lng !== undefined) formData.append('lng', String(lng))
  return api.upload<ApiResponse<VisionClassifyResponse>>('/vision/classify', formData)
}

export function getMyReports(params?: { page?: number; page_size?: number; category?: string }) {
  return api.get<ApiResponse<{ reports: GarbageReport[]; pagination: Pagination }>>('/vision/reports', params as Record<string, unknown>)
}
