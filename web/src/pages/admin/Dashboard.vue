<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElCard, ElRow, ElCol, ElStatistic, ElIcon, ElTable, ElTableColumn, ElTag } from 'element-plus'
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
  } catch {
    // Use defaults
  }
})
</script>

<style scoped>
.admin-dashboard {
  max-width: 1200px;
}

.admin-dashboard h2 {
  font-size: 22px;
  color: #303133;
  margin: 0 0 20px;
}

.stat-grid {
  margin-bottom: 16px;
}

.stat-card {
  cursor: pointer;
  transition: transform 0.2s;
}

.stat-card:hover {
  transform: translateY(-2px);
}

.stat-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: 14px;
  padding: 20px;
}

.stat-icon {
  width: 48px;
  height: 48px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #fff;
  flex-shrink: 0;
}

.stat-body {
  display: flex;
  flex-direction: column;
}

.stat-value {
  font-size: 24px;
  font-weight: 700;
  color: #303133;
}

.stat-label {
  font-size: 12px;
  color: #909399;
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
</style>
