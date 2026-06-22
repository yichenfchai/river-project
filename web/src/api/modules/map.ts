import { api } from '@/api/client'
import type { ApiResponse } from '@/types'

export interface MapLayerInfo {
  id: string
  name: string
  era: string
  description: string
  color: string
  sort_order: number
}

export interface MapLayerDetail extends MapLayerInfo {
  geojson: unknown
}

export interface MapPOI {
  id: string
  name: string
  description: string
  category: string
  lat: number
  lng: number
  layer_id: string
}

export function getMapLayers() {
  return api.get<ApiResponse<MapLayerInfo[]>>('/map/layers')
}

export function getMapLayer(id: string) {
  return api.get<ApiResponse<MapLayerDetail>>(`/map/layers/${id}`)
}

export function getPOIs(lat: number, lng: number, radius = 50000, category?: string) {
  return api.get<ApiResponse<MapPOI[]>>('/map/pois', {
    lat,
    lng,
    radius,
    category: category || undefined,
  })
}
