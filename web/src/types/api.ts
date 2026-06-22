export interface ApiResponse<T = unknown> {
  code: number
  message: string
  data: T
}

export interface Pagination {
  page: number
  page_size: number
  total: number
  total_pages: number
}
