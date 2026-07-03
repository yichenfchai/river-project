<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { getStories } from '@/api/modules/story'
import type { Story } from '@/api/modules/story'

const router = useRouter()
const stories = ref<Story[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getStories({ page_size: 20 })
    stories.value = res.data.stories || []
  } catch {
    stories.value = []
  } finally {
    loading.value = false
  }
})

function openStory(id: string) {
  router.push(`/story/${id}`)
}
</script>

<template>
  <div class="story-page">
    <h2>📖 科普故事会</h2>
    <p class="page-desc">AI 驱动的运河文化科普内容 —— 跨越 2500 年的运河故事</p>

    <div class="story-grid" v-loading="loading">
      <div v-if="!loading && stories.length === 0" class="empty-hint">暂无科普故事，敬请期待</div>
      <div v-for="story in stories" :key="story.id" class="story-card" @click="openStory(story.id)">
        <div class="story-cover">{{ story.emoji || '📖' }}</div>
        <div class="story-info">
          <span class="story-tag">{{ story.topic }}</span>
          <h3>{{ story.title }}</h3>
          <p>{{ story.content.slice(0, 80) }}...</p>
          <div class="story-meta">
            <span>{{ story.age_group }}</span>
            <span>{{ story.likes }} 赞</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>


<style scoped>
.story-page {
  max-width: 1200px;
  margin: 0 auto;
}

.story-page h2 {
  font-size: 22px;
  color: #303133;
}

.page-desc {
  color: #909399;
  font-size: 14px;
  margin-bottom: 24px;
}

.story-grid {
  display: grid;
  grid-template-columns: repeat(2, 1fr);
  gap: 20px;
}

.story-card {
  display: flex;
  background: #fff;
  border-radius: 12px;
  overflow: hidden;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.06);
  cursor: pointer;
  transition: box-shadow 0.2s;
}

.story-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
}

.story-cover {
  width: 120px;
  min-height: 140px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 48px;
  background: #fff;
  flex-shrink: 0;
}

.story-info {
  padding: 16px 20px;
  flex: 1;
}

.story-tag {
  font-size: 11px;
  padding: 2px 8px;
  background: #eae7e0;
  color: #2c3e50;
  border-radius: 4px;
}

.story-info h3 {
  margin: 8px 0;
  font-size: 16px;
  color: #303133;
}

.story-info p {
  color: #909399;
  font-size: 13px;
  line-height: 1.6;
  margin: 0;
}

.story-meta {
  margin-top: 12px;
  display: flex;
  gap: 16px;
  font-size: 12px;
  color: #c0c4cc;
}

.empty-hint {
  text-align: center;
  color: #c0c4cc;
  padding: 40px 0;
  grid-column: 1 / -1;
}

@media (max-width: 768px) {
  .story-grid {
    grid-template-columns: 1fr;
  }

  .story-page {
    padding: 0 8px;
  }
}

@media (max-width: 480px) {
  .story-card {
    flex-direction: column;
  }

  .story-cover {
    width: 100%;
    min-height: 100px;
    font-size: 36px;
  }

  .story-info {
    padding: 12px 14px;
  }
}
</style>
