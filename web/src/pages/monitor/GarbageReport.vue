<script setup lang="ts">
import { ref } from 'vue'
import { ElMessage, ElIcon } from 'element-plus'
import { Camera, Upload, Position } from '@element-plus/icons-vue'
import { useLocation } from '@/composables/useLocation'

const { getCurrentPosition } = useLocation()

const previewUrl = ref<string | null>(null)
const uploading = ref(false)
const result = ref<{ detections: { class_name: string; category: string; confidence: number }[]; advice: string } | null>(null)
const lat = ref<number | null>(null)
const lng = ref<number | null>(null)

async function handleCapture() {
  lat.value = null
  lng.value = null
  try {
    const pos = await getCurrentPosition()
    lat.value = pos.lat
    lng.value = pos.lng
  } catch {
    // Location optional
  }
}

function onFileChange(e: Event) {
  const file = (e.target as HTMLInputElement).files?.[0]
  if (!file) return
  previewUrl.value = URL.createObjectURL(file)
}

async function submitReport() {
  result.value = null
  uploading.value = true
  try {
    await new Promise((r) => setTimeout(r, 1500))
    result.value = {
      detections: [
        { class_name: '塑料瓶', category: '可回收物', confidence: 0.96 },
      ],
      advice: '塑料瓶属于可回收物，请清洗后投入蓝色垃圾桶',
    }
    ElMessage.success('识别完成')
  } catch {
    ElMessage.error('识别失败，请重试')
  } finally {
    uploading.value = false
  }
}
</script>

<template>
  <div class="report-page">
    <h2>📷 垃圾分类上报</h2>
    <p class="page-desc">拍摄或上传垃圾图片进行 AI 识别分类</p>

    <div class="report-grid">
      <div class="capture-section">
        <div class="capture-area" @click="handleCapture">
          <div v-if="!previewUrl" class="capture-placeholder">
            <el-icon :size="48"><Camera /></el-icon>
            <p>点击拍照或选择图片</p>
            <input type="file" accept="image/*" capture="environment" class="file-input" @change="onFileChange" />
          </div>
          <img v-else :src="previewUrl" class="preview-image" alt="preview" />
        </div>

        <div class="location-info" v-if="lat !== null">
          <el-icon><Position /></el-icon>
          <span>位置已获取 ({{ lat.toFixed(4) }}, {{ lng!.toFixed(4) }})</span>
        </div>

        <el-button
          type="primary"
          size="large"
          :loading="uploading"
          :disabled="!previewUrl"
          class="submit-btn"
          @click="submitReport"
        >
          <el-icon><Upload /></el-icon>
          {{ uploading ? '识别中...' : '提交识别' }}
        </el-button>
      </div>

      <div class="result-section" v-if="result">
        <h3>识别结果</h3>
        <div v-for="d in result.detections" :key="d.class_name" class="detection-item">
          <span class="detection-name">{{ d.class_name }}</span>
          <el-tag>{{ d.category }}</el-tag>
          <span class="confidence">置信度 {{ (d.confidence * 100).toFixed(0) }}%</span>
        </div>
        <div class="advice-box">
          <span class="advice-label">分类建议</span>
          <p>{{ result.advice }}</p>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.report-page {
  max-width: 900px;
}

.report-page h2 {
  font-size: 22px;
  color: #303133;
  margin: 0 0 8px;
}

.page-desc {
  color: #909399;
  font-size: 14px;
  margin-bottom: 24px;
}

.report-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 24px;
}

.capture-section {
  display: flex;
  flex-direction: column;
}

.capture-area {
  aspect-ratio: 1;
  background: #f0ede7;
  border: 2px dashed #dcdfe6;
  border-radius: 12px;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  position: relative;
  overflow: hidden;
  transition: border-color 0.2s;
}

.capture-area:hover {
  border-color: #2c3e50;
}

.capture-placeholder {
  text-align: center;
  color: #c0c4cc;
}

.capture-placeholder p {
  margin-top: 12px;
}

.file-input {
  position: absolute;
  inset: 0;
  opacity: 0;
  cursor: pointer;
}

.preview-image {
  width: 100%;
  height: 100%;
  object-fit: cover;
}

.location-info {
  display: flex;
  align-items: center;
  gap: 6px;
  margin-top: 12px;
  color: #67c23a;
  font-size: 13px;
}

.submit-btn {
  margin-top: 16px;
}

.result-section {
  background: #fff;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.result-section h3 {
  margin: 0 0 16px;
  color: #303133;
}

.detection-item {
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 12px 0;
  border-bottom: 1px solid #f0f2f5;
}

.detection-name {
  font-weight: 500;
  color: #303133;
}

.confidence {
  color: #67c23a;
  font-size: 13px;
  margin-left: auto;
}

.advice-box {
  margin-top: 20px;
  padding: 16px;
  background: #f0f9eb;
  border-radius: 8px;
}

.advice-label {
  font-size: 12px;
  color: #67c23a;
  font-weight: 500;
}

.advice-box p {
  margin: 6px 0 0;
  color: #606266;
  font-size: 14px;
}
</style>
