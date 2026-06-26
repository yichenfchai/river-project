<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { getStory } from '@/api/modules/story'
import type { Story } from '@/api/modules/story'

const route = useRoute()
const router = useRouter()
const story = ref<Story | null>(null)
const loading = ref(true)
const notFound = ref(false)

onMounted(async () => {
  const id = route.params.id as string
  if (!id) {
    notFound.value = true
    loading.value = false
    return
  }
  try {
    const res = await getStory(id)
    story.value = res.data
  } catch {
    notFound.value = true
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="story-detail-page" v-loading="loading">
    <div class="detail-header" v-if="story && !loading">
      <button class="back-btn" @click="router.back()">&larr; 返回</button>
      <span class="detail-tag">{{ story.topic }}</span>
      <h1>{{ story.emoji }} {{ story.title }}</h1>
      <div class="detail-meta">
        <span>👤 适读：{{ story.age_group }}</span>
        <span>❤️ {{ story.likes }} 赞</span>
        <span>🕐 {{ new Date(story.created_at).toLocaleDateString('zh-CN') }}</span>
      </div>
    </div>

    <div class="detail-body" v-if="story && !loading">
      <p class="detail-content">{{ story.content }}</p>
    </div>

    <div class="not-found" v-if="notFound && !loading">
      <p>故事不存在或已被删除</p>
      <button class="back-btn" @click="router.push('/story')">返回故事列表</button>
    </div>
  </div>
</template>

<style scoped>
.story-detail-page {
  max-width: 800px;
  margin: 0 auto;
  min-height: 60vh;
}

.detail-header {
  margin-bottom: 32px;
}

.back-btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  background: none;
  border: none;
  color: #2c3e50;
  font-size: 14px;
  cursor: pointer;
  padding: 6px 0;
  margin-bottom: 16px;
  transition: color 0.2s;
}

.back-btn:hover {
  color: #c9b896;
}

.detail-tag {
  display: inline-block;
  font-size: 11px;
  padding: 2px 10px;
  background: #eae7e0;
  color: #2c3e50;
  border-radius: 4px;
  margin-bottom: 12px;
}

.detail-header h1 {
  font-size: 28px;
  color: #303133;
  margin: 0 0 16px;
  line-height: 1.4;
}

.detail-meta {
  display: flex;
  gap: 24px;
  font-size: 13px;
  color: #909399;
}

.detail-body {
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.detail-content {
  font-size: 16px;
  line-height: 2;
  color: #303133;
  white-space: pre-wrap;
  margin: 0;
}

.not-found {
  text-align: center;
  padding: 80px 20px;
  color: #909399;
}

.not-found p {
  margin: 0 0 24px;
  font-size: 16px;
}

@media (max-width: 768px) {
  .story-detail-page {
    padding: 0 12px;
  }

  .detail-header h1 {
    font-size: 22px;
  }

  .detail-body {
    padding: 20px 16px;
    border-radius: 10px;
  }

  .detail-content {
    font-size: 15px;
    line-height: 1.9;
  }

  .detail-meta {
    gap: 16px;
    flex-wrap: wrap;
  }
}
</style>
