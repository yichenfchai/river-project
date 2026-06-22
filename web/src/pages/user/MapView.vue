<script setup lang="ts">
import { ref, onMounted, onUnmounted } from 'vue'
import { ElButton, ElTag, ElEmpty, ElMessage } from 'element-plus'
import { getMapLayers, getMapLayer, getPOIs } from '@/api/modules/map'
import type { MapLayerInfo, MapPOI } from '@/api/modules/map'

declare const L: any

const layers = ref<MapLayerInfo[]>([])
const activeLayer = ref<string | null>(null)
const pois = ref<MapPOI[]>([])
const loading = ref(false)
let map: any = null
const geoLayers: Record<string, any> = {}
const poiMarkers: any[] = []
const cityMarkers: any[] = []

const canalCities = [
  { name: '北京', lat: 39.90, lng: 116.40 },
  { name: '天津', lat: 39.14, lng: 117.19 },
  { name: '沧州', lat: 38.30, lng: 116.84 },
  { name: '德州', lat: 37.43, lng: 116.36 },
  { name: '临清', lat: 36.85, lng: 115.71 },
  { name: '济宁', lat: 35.40, lng: 116.58 },
  { name: '徐州', lat: 34.30, lng: 117.20 },
  { name: '淮安', lat: 33.51, lng: 119.14 },
  { name: '扬州', lat: 32.39, lng: 119.43 },
  { name: '镇江', lat: 32.20, lng: 119.40 },
  { name: '苏州', lat: 31.32, lng: 120.62 },
  { name: '杭州', lat: 30.32, lng: 120.14 },
] as const

// 降级 GeoJSON — 后端服务不可用时前端直接渲染
const fallbackLayers: MapLayerInfo[] = [
  { id: 'fallback-sui-tang', name: '隋唐运河', era: 'sui-tang', description: '隋唐大运河以洛阳为中心', color: '#e6a23c', sort_order: 1 },
  { id: 'fallback-yuan-ming-qing', name: '元明清运河', era: 'yuan-ming-qing', description: '京杭大运河直通南北', color: '#3a7ca5', sort_order: 2 },
  { id: 'fallback-south-north', name: '现代南水北调东线', era: 'modern', description: '南水北调东线工程', color: '#67c23a', sort_order: 3 },
]

const fallbackGeoJSON: Record<string, string> = {
  'fallback-sui-tang': `{"type":"LineString","coordinates":[[116.4,39.9],[116.2,39.0],[115.9,38.0],[116.3,37.4],[116.6,35.4],[117.2,34.3],[119.1,33.5],[112.4,34.6],[114.3,32.1],[116.0,31.0],[119.4,32.4],[120.6,31.3],[120.2,30.3]]}`,
  'fallback-yuan-ming-qing': `{"type":"LineString","coordinates":[[116.4,39.9],[117.2,39.1],[116.8,38.3],[116.3,37.4],[117.0,36.7],[116.6,35.4],[117.2,34.3],[118.1,33.9],[119.1,33.5],[119.4,32.4],[119.4,32.2],[119.6,31.8],[120.6,31.3],[120.2,30.3]]}`,
  'fallback-south-north': `{"type":"LineString","coordinates":[[119.4,32.4],[119.4,32.2],[119.0,33.5],[117.2,34.3],[116.6,35.4],[116.3,37.4],[116.8,38.3],[117.2,39.1]]}`,
}

async function fetchLayers() {
  try {
    const res = await getMapLayers()
    layers.value = res.data
  } catch {
    layers.value = fallbackLayers
  }
}

function initMap() {
  map = L.map('leaflet-map', {
    center: [34.5, 115.0],
    zoom: 6,
    zoomControl: true,
  })

  L.tileLayer('https://{s}.tile.openstreetmap.org/{z}/{x}/{y}.png', {
    attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a>',
    maxZoom: 18,
  }).addTo(map)  // eslint-disable-line
}

async function toggleLayer(layerInfo: MapLayerInfo) {
  if (activeLayer.value === layerInfo.id) {
    if (geoLayers[layerInfo.id]) {
      map.removeLayer(geoLayers[layerInfo.id])
      delete geoLayers[layerInfo.id]
    }
    clearPOIs()
    clearCityLabels()
    activeLayer.value = null
    return
  }

  if (activeLayer.value && geoLayers[activeLayer.value]) {
    map.removeLayer(geoLayers[activeLayer.value])
    delete geoLayers[activeLayer.value]
  }

  loading.value = true
  if (layerInfo.id.startsWith('fallback-')) {
    const fb = fallbackGeoJSON[layerInfo.id]
    if (fb) {
      renderGeoJSON(layerInfo.id, fb, layerInfo.color)
      showCityLabels()
    }
    loading.value = false
    return
  }

  try {
    const res = await getMapLayer(layerInfo.id)
    const geojson = res.data.geojson
    renderGeoJSON(layerInfo.id, geojson, layerInfo.color)
    fetchPOIs()
    showCityLabels()
  } catch {
    ElMessage.error('加载图层失败')
  } finally {
    loading.value = false
  }
}

function renderGeoJSON(layerId: string, geojson: unknown, color: string) {
  if (geojson && typeof geojson === 'string') {
    const parsed = JSON.parse(geojson)
    const geo = L.geoJSON(parsed, { style: { color, weight: 3, opacity: 0.8 } }).addTo(map)
    geoLayers[layerId] = geo
    map.fitBounds(geo.getBounds(), { paddingTopLeft: [30, 320], paddingBottomRight: [30, 30], maxZoom: 7 })
    map.invalidateSize()
  }
  activeLayer.value = layerId
}

async function fetchPOIs() {
  clearPOIs()
  const center = map.getCenter()
  try {
    const res = await getPOIs(center.lat, center.lng, 500000)
    pois.value = res.data
    showPOIs()
  } catch {
    // noop
  }
}

function showCityLabels() {
  clearCityLabels()
  for (const city of canalCities) {
    const icon = L.divIcon({
      html: `<div style="font-size:11px;font-weight:500;color:#303133;background:rgba(255,255,255,0.88);padding:2px 6px;border-radius:3px;white-space:nowrap;box-shadow:0 1px 4px rgba(0,0,0,0.12)">${city.name}</div>`,
      className: 'city-label',
      iconSize: [0, 0],
      iconAnchor: [0, 0],
    })
    const m = L.marker([city.lat, city.lng], { icon, interactive: false })
    m.addTo(map)
    cityMarkers.push(m)
  }
}

function clearCityLabels() {
  for (const m of cityMarkers) map.removeLayer(m)
  cityMarkers.length = 0
}

function showPOIs() {
  clearPOIs()
  const categoryIcons: Record<string, string> = {
    cultural: '🏛',
    ecology: '🌿',
  }

  for (const poi of pois.value) {
    const icon = L.divIcon({
      html: `<div style="font-size:20px;text-align:center;line-height:1">${categoryIcons[poi.category] || '📍'}</div>`,
      className: 'poi-icon',
      iconSize: [30, 30],
      iconAnchor: [15, 15],
    })

    const marker = L.marker([poi.lat, poi.lng], { icon })
      .bindPopup(`<b>${poi.name}</b><br/>${poi.description}`)
      .addTo(map)
    poiMarkers.push(marker)
  }
}

function clearPOIs() {
  for (const m of poiMarkers) {
    map.removeLayer(m)
  }
  poiMarkers.length = 0
}

onMounted(async () => {
  initMap()
  await fetchLayers()
  if (layers.value.length > 0) {
    const defaultLayer = layers.value.find(l => l.era === 'yuan-ming-qing') || layers.value[0]
    toggleLayer(defaultLayer!)
  }
})

onUnmounted(() => {
  clearCityLabels()
  clearPOIs()
  if (map) {
    map.remove()
    map = null
  }
})
</script>

<template>
  <div class="map-page">
    <div class="map-header">
      <h2>🗺 大运河时空地图</h2>
      <p class="page-desc">探索大运河从隋唐至今的千年变迁 — 点击图层查看完整路线</p>
    </div>

    <div class="map-layout">
      <div class="map-sidebar">
        <div class="layer-list">
          <div
            v-for="layer in layers"
            :key="layer.id"
            class="layer-item"
            :class="{ active: activeLayer === layer.id }"
            :style="{ '--layer-color': layer.color }"
            v-loading="loading && activeLayer !== layer.id"
            @click="toggleLayer(layer)"
          >
            <span class="layer-dot" :style="{ background: layer.color }"></span>
            <div class="layer-info">
              <span class="layer-name">{{ layer.name }}</span>
              <span class="layer-era">{{ layer.era }}</span>
            </div>
            <el-tag v-if="activeLayer === layer.id" size="small" type="success">选中</el-tag>
          </div>
        </div>

        <div v-if="activeLayer" class="poi-section">
          <h4>沿线遗产与站点</h4>
          <div v-for="poi in pois" :key="poi.id" class="poi-item" @click="map?.panTo([poi.lat, poi.lng])">
            <span class="poi-icon">{{ poi.category === 'cultural' ? '🏛' : '🌿' }}</span>
            <div class="poi-info">
              <span class="poi-name">{{ poi.name }}</span>
              <span class="poi-desc">{{ poi.description }}</span>
            </div>
          </div>
          <el-empty v-if="pois.length === 0" description="暂无POI数据" :image-size="40" />
        </div>
      </div>

      <div class="map-container">
        <div id="leaflet-map" style="width: 100%; height: 100%"></div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.map-page {
  flex: 1;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.map-header {
  padding: 16px 24px 8px;
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
}

.map-header h2 {
  margin: 0 0 4px;
  font-size: 20px;
}

.page-desc {
  margin: 0;
  color: #909399;
  font-size: 13px;
}

.map-layout {
  flex: 1;
  display: flex;
  overflow: hidden;
}

.map-sidebar {
  width: 300px;
  overflow-y: auto;
  background: #fff;
  border-right: 1px solid #e4e7ed;
  display: flex;
  flex-direction: column;
}

.layer-list {
  padding: 12px;
}

.layer-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px;
  border-radius: 8px;
  cursor: pointer;
  transition: background 0.2s;
  margin-bottom: 6px;
  border: 1px solid transparent;
}

.layer-item:hover {
  background: #fff;
}

.layer-item.active {
  background: #eae7e0;
  border-color: var(--layer-color);
}

.layer-dot {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  flex-shrink: 0;
}

.layer-info {
  flex: 1;
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.layer-name {
  font-size: 14px;
  color: #303133;
  font-weight: 500;
}

.layer-era {
  font-size: 12px;
  color: #909399;
}

.poi-section {
  padding: 12px;
  border-top: 1px solid #e4e7ed;
  flex: 1;
  overflow-y: auto;
}

.poi-section h4 {
  margin: 0 0 8px;
  font-size: 14px;
  color: #303133;
}

.poi-item {
  display: flex;
  align-items: flex-start;
  gap: 8px;
  padding: 8px;
  border-radius: 6px;
  cursor: pointer;
  transition: background 0.15s;
  margin-bottom: 4px;
}

.poi-item:hover {
  background: #fff;
}

.poi-info {
  display: flex;
  flex-direction: column;
  gap: 2px;
}

.poi-name {
  font-size: 13px;
  color: #303133;
  font-weight: 500;
}

.poi-desc {
  font-size: 11px;
  color: #909399;
  line-height: 1.4;
}

.map-container {
  flex: 1;
  position: relative;
}

:deep(.leaflet-popup-content) {
  font-size: 13px;
}

@media (max-width: 768px) {
  .map-layout {
    flex-direction: column;
  }

  .map-sidebar {
    width: 100%;
    max-height: 200px;
  }

  .map-container {
    flex: 1;
  }
}
</style>
