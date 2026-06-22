<script setup lang="ts">
import { ElCard, ElRow, ElCol, ElTable, ElTableColumn, ElTag, ElIcon, ElStatistic } from 'element-plus'
import { Camera, DataAnalysis, Check, Clock } from '@element-plus/icons-vue'
import { useRouter } from 'vue-router'

const router = useRouter()
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
          <el-statistic title="今日上报" :value="12" />
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover">
          <el-statistic title="本月累计" :value="86" />
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover">
          <el-statistic title="已验证" :value="72" />
          <template #suffix><el-icon><Check /></el-icon></template>
        </ElCard>
      </ElCol>
      <ElCol :span="6">
        <ElCard shadow="hover">
          <el-statistic title="待处理" :value="14" />
          <template #suffix><el-icon><Clock /></el-icon></template>
        </ElCard>
      </ElCol>
    </ElRow>

    <ElRow :gutter="16" style="margin-top: 16px">
      <ElCol :span="12">
        <ElCard>
          <template #header>
            <span><el-icon><DataAnalysis /></el-icon> 分类统计</span>
          </template>
          <ElTable :data="categoryStats" size="small">
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
          <div v-for="item in recentReports" :key="item.id" class="recent-item">
            <span class="item-type">{{ item.type }}</span>
            <span class="item-confidence">置信度 {{ item.confidence }}%</span>
            <el-tag size="small" :type="item.status === 'verified' ? 'success' : 'warning'">
              {{ item.status === 'verified' ? '已验证' : '待处理' }}
            </el-tag>
            <span class="item-time">{{ item.time }}</span>
          </div>
        </ElCard>
      </ElCol>
    </ElRow>
  </div>
</template>

<script lang="ts">
const categoryStats = [
  { category: '可回收物', count: 35, percent: 41 },
  { category: '其他垃圾', count: 28, percent: 33 },
  { category: '厨余垃圾', count: 12, percent: 14 },
  { category: '有害垃圾', count: 11, percent: 12 },
]
const recentReports = [
  { id: 1, type: '塑料瓶', confidence: 96, status: 'verified', time: '10 分钟前' },
  { id: 2, type: '塑料袋', confidence: 88, status: 'pending', time: '30 分钟前' },
  { id: 3, type: '废电池', confidence: 94, status: 'verified', time: '1 小时前' },
  { id: 4, type: '玻璃瓶', confidence: 91, status: 'pending', time: '2 小时前' },
]
</script>

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
}
</style>
