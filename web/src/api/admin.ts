import client from './client'
import type { ApiResponse, ListResponse, User } from '@/types'

export const adminApi = {
  listUsers(page = 1, pageSize = 20) {
    return client.get<ApiResponse<ListResponse<User>>>('/admin/users', {
      params: { page, page_size: pageSize },
    })
  },
  changeRole(userId: string, role: string) {
    return client.put<ApiResponse>(`/admin/users/${userId}/role`, { role })
  },
  banUser(userId: string, ban: boolean) {
    return client.post<ApiResponse>(`/admin/users/${userId}/ban`, { ban })
  },
}
