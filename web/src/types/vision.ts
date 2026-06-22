export type GarbageCategory = '可回收物' | '有害垃圾' | '厨余垃圾' | '其他垃圾'
export type ReportStatus = 'pending' | 'verified' | 'dismissed'

export interface GarbageDetection {
  class_name: string
  category: GarbageCategory
  confidence: number
  bbox: {
    x: number
    y: number
    w: number
    h: number
  }
}

export interface VisionClassifyResponse {
  image_id: string
  detections: GarbageDetection[]
  processing_time_ms: number
  advice: string
}

export interface GarbageReport {
  id: string
  image_url: string
  detections: GarbageDetection[]
  lat: number | null
  lng: number | null
  status: ReportStatus
  reported_at: string
}
