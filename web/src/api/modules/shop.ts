import { api } from '@/api/client'
import type { ApiResponse } from '@/types'

export interface ShopItem {
  id: string
  name: string
  description: string
  image_url: string
  points_cost: number
  stock: number
  is_active: boolean
  created_at: string
  updated_at: string
}

export interface Redemption {
  id: string
  user_id: string
  item_id: string
  item_name: string
  points_spent: number
  status: string
  created_at: string
}

export interface RedeemResult {
  redemption: Redemption
  user_points: number
}

export function getShopItems() {
  return api.get<ApiResponse<ShopItem[]>>('/shop/items')
}

export function redeemItem(itemId: string) {
  return api.post<ApiResponse<RedeemResult>>('/shop/redeem', { item_id: itemId })
}

export function getHistory(page = 1, pageSize = 20) {
  return api.get<ApiResponse<{ items: Redemption[]; pagination: { page: number; total: number } }>>('/shop/history', {
    page,
    page_size: pageSize,
  })
}

export function createShopItem(data: Omit<ShopItem, 'id' | 'created_at' | 'updated_at'> & { stock?: number }) {
  return api.post<ApiResponse<ShopItem>>('/admin/shop/items', data)
}

export function updateShopItem(id: string, data: Omit<ShopItem, 'id' | 'created_at' | 'updated_at'> & { is_active: boolean; stock?: number }) {
  return api.put<ApiResponse<ShopItem>>(`/admin/shop/items/${id}`, data)
}

export function deleteShopItem(id: string) {
  return api.delete(`/admin/shop/items/${id}`)
}
