<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import { ElCard, ElRow, ElCol, ElTable, ElTableColumn, ElTag, ElIcon, ElEmpty } from 'element-plus'
import { Camera, DataAnalysis, Check, Clock } from '@element-plus/icons-vue'
import { getMyReports } from '@/api/modules/vision'
import type { GarbageReport } from '@/types'

const router = useRouter()

interface StatValues { today: number; month: number; verified: number; pending: number }
const stats = ref<StatValues>({ today: 0, month: 0, verified: 0, pending: 0 })
const categoryStats = ref<{ category: string; count: number; percent: number }[]>([])
const recentReports = ref<GarbageReport[]>([])
const loading = ref(true)

onMounted(async () => {
  try {
    const res = await getMyReports({ page: 1, page_size: 50 })
    const reports = res.data.reports || []
    
    // 计数
    const today = new Date()
    today.setHours(0, 0, 0, 0)
    let todayCount = 0, verifiedCount = 0, pendingCount = 0
    
    reports.forEach((r: GarbageReport) => {
      if (new Date(r.reported_at) >= today) todayCount++
      if (r.status === 'verified') verifiedCount++
      if (r.status === 'pending') pendingCount++
    })
    
    stats.value = {
      today: todayCount,
      month: reports.length,
      verified: verifiedCount,
      pending: pendingCount,
    }
    
    // 分类统计
    const catMap: Record<string, number> = {}
    reports.forEach((r: GarbageReport) => {
      r.detections?.forEach(d => {
        catMap[d.category] = (catMap[d.category] || 0) + 1
      })
    })
    const total = Object.values(catMap).reduce((a, b) => a + b, 0) || 1
    categoryStats.value = Object.entries(catMap).map(([cat, cnt]) => ({
      category: cat, count: cnt, percent: Math.round((cnt / total) * 100)
    }))
    
    recentReports.value = reports.slice(0, 5)
  } catch {
    // API 未就绪时显示空状态
  } finally {
    loading.value = false
  }
})
</script>

<template>
  <div class="monitor-dashboard">
    <div class="dashboard-header">
      <h2>监测数据看板</h2>
      <el-button type="primary" @click="router.push('/monitor/report')">
        <el-icon><Camera /></el-icon> 立即上报
      </el-button>
    </div>

    <ElRow :gutter="16" class="stat-row">
      <ElCol :span="6">
        <ElCard shadow="hover">
          <el-statistic title="今日上报" :value="stats.today" />
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover">
          <el-statistic title="本月累计" :value="stats.month" />
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover">
          <el-statistic title="已验证" :value="stats.verified">
            <template #suffix><el-icon><Check /></el-icon></template>
          </el-statistic>
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover">
          <el-statistic title="待处理" :value="stats.pending">
            <template #suffix><el-icon><Clock /></el-icon></template>
          </el-statistic>
        </ElCard>
      </ElCol>
    </ElRow>

    <ElRow :gutter="16" style="margin-top: 16px">
      <ElCol :span="12">
        <ElCard>
          <template #header>
            <span><el-icon><DataAnalysis /></el-icon> 分类统计</span>
          </template>
          <el-empty v-if="categoryStats.length === 0" description="暂无数据" :image-size="60" />
          <ElTable v-else :data="categoryStats" size="small">
            <ElTableColumn prop="category" label="类别" />
            <ElTableColumn prop="count" label="数量" width="80" />
            <ElTableColumn prop="percent" label="占比" width="100">
              <template #default="{ row }">
                <el-tag size="small">{{ row.percent }}%</el-tag>
              </template>
            </ElTableColumn>
          </ElTable>
        </ElCard>
      </ElCol>
      <ElCol :span="12">
        <ElCard>
          <template #header>
            <span>最近上报</span>
          </template>
          <el-empty v-if="recentReports.length === 0" description="暂无上报记录" :image-size="60" />
          <div v-for="item in recentReports" :key="item.id" class="recent-item">
            <span class="item-type">{{ item.detections?.[0]?.class_name || '未知' }}</span>
            <span class="item-confidence" v-if="item.detections?.[0]">置信度 {{ Math.round(item.detections[0].confidence * 100) }}%</span>
            <el-tag size="small" :type="item.status === 'verified' ? 'success' : 'warning'">
              {{ item.status === 'verified' ? '已验证' : '待处理' }}
            </el-tag>
            <span class="item-time">{{ item.reported_at }}</span>
          </div>
        </ElCard>
      </ElCol>
    </ElRow>
  </div>
</template>

<style scoped>
.monitor-dashboard {
  max-width: 1200px;
}

.dashboard-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  margin-bottom: 20px;
}

.dashboard-header h2 {
  font-size: 22px;
  color: #303133;
  margin: 0;
}

.stat-row {
  margin-bottom: 0;
}

.recent-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 10px 0;
  border-bottom: 1px solid #f0f2f5;
  font-size: 13px;
}

.recent-item:last-child {
  border-bottom: none;
}

.item-type {
  font-weight: 500;
  color: #303133;
}

.item-confidence {
  color: #67c23a;
}

.item-time {
  color: #c0c4cc;
  margin-left: auto;
  font-size: 12px;
}
</style>
