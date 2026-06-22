export function useLocation() {
  async function getCurrentPosition(options?: {
    enableHighAccuracy?: boolean
    timeout?: number
  }): Promise<{ lat: number; lng: number; accuracy: number }> {
    return new Promise((resolve, reject) => {
      if (!navigator.geolocation) {
        reject(new Error('浏览器不支持定位'))
        return
      }
      navigator.geolocation.getCurrentPosition(
        (pos) =>
          resolve({
            lat: pos.coords.latitude,
            lng: pos.coords.longitude,
            accuracy: pos.coords.accuracy,
          }),
        (err) => {
          const messages: Record<number, string> = {
            1: '定位权限被拒绝',
            2: '无法获取位置信息',
            3: '定位超时',
          }
          reject(new Error(messages[err.code] || '定位失败'))
        },
        {
          enableHighAccuracy: options?.enableHighAccuracy ?? false,
          timeout: options?.timeout ?? 10000,
        },
      )
    })
  }

  return { getCurrentPosition }
}
