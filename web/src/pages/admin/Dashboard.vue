<template>
  <div class="admin-dashboard" v-loading="loading">
    <h2>数据看板</h2>

    <el-alert
      v-if="error"
      :title="error"
      type="error"
      show-icon
      closable
      @close="error = ''"
      style="margin-bottom: 16px"
    />

    <el-row :gutter="16" class="stat-grid">
      <el-col v-for="card in statCards" :key="card.key" :xs="12" :sm="8" :md="4">
        <el-card class="stat-card" shadow="hover">
          <div class="stat-icon" :style="{ background: card.bg }">
            <el-icon :size="22"><component :is="card.icon" /></el-icon>
          </div>
          <div class="stat-body">
            <span class="stat-value">{{ stats[card.key] }}</span>
            <span class="stat-label">{{ card.label }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElCard, ElRow, ElCol, ElIcon, ElAlert } from 'element-plus'
import { User, ChatLineSquare, Document, DataBoard, Check, DeleteFilled } from '@element-plus/icons-vue'
import { getAdminDashboard } from '@/api/modules/admin'

const statCards = [
  { key: 'total_users', label: '注册用户', icon: User, bg: '#e6f7ff' },
  { key: 'active_today', label: '今日活跃', icon: DataBoard, bg: '#f6ffed' },
  { key: 'total_posts', label: '帖子总数', icon: ChatLineSquare, bg: '#fff7e6' },
  { key: 'pending_reviews', label: '待审核', icon: Document, bg: '#fff1f0' },
  { key: 'quiz_players', label: '答题用户', icon: Check, bg: '#f0f5ff' },
  { key: 'garbage_reports', label: '垃圾上报', icon: DeleteFilled, bg: '#f9f0ff' },
] as const

const loading = ref(true)
const error = ref('')

const stats = ref({
  total_users: 0,
  active_today: 0,
  total_posts: 0,
  pending_reviews: 0,
  quiz_players: 0,
  garbage_reports: 0,
})

onMounted(async () => {
  try {
    const res = await getAdminDashboard()
    stats.value = res.data
  } catch (e: unknown) {
    error.value = (e as { message?: string })?.message || '加载数据失败，请稍后重试'
  } finally {
    loading.value = false
  }
})
</script>

<style scoped>
.admin-dashboard {
  max-width: 1200px;
}

.admin-dashboard h2 {
  font-family: 'Noto Serif SC', 'STSong', serif;
  font-size: 22px;
  color: var(--text-primary);
  margin: 0 0 20px;
}

.stat-grid { margin-bottom: 16px; }

.stat-card {
  cursor: pointer;
  transition: transform 0.2s, box-shadow 0.2s;
  border: 1px solid var(--border-hairline);
  border-radius: var(--radius-lg);
  overflow: hidden;
}
.stat-card:hover { transform: translateY(-3px); box-shadow: var(--shadow-md); }

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 20px;
  background: var(--surface-card);
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: var(--radius-md);
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
  box-shadow: inset 0 1px 0 rgba(255,255,255,0.2);
}

.stat-body { display: flex; flex-direction: column; }

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: var(--text-primary);
  font-family: 'Georgia', 'Noto Serif SC', serif;
}

.stat-label {
  font-size: 12px;
  color: var(--text-secondary);
  margin-top: 2px;
}

.cards-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 16px;
}

.info-card {
  min-height: 200px;
}

.status-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.status-item {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.status-label {
  font-size: 14px;
  color: #606266;
}

.quick-actions {
  display: flex;
  flex-wrap: wrap;
  gap: 10px;
}

@media (max-width: 768px) {
  .cards-grid {
    grid-template-columns: 1fr;
  }

  .stat-card :deep(.el-card__body) {
    padding: 14px;
    gap: 10px;
  }

  .stat-icon {
    width: 40px;
    height: 40px;
  }

  .stat-value {
    font-size: 20px;
  }
}
</style>
