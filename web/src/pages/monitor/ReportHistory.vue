<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElTable, ElTableColumn, ElTag, ElImage, ElPagination, ElEmpty } from 'element-plus'
import { getMyReports } from '@/api/modules/vision'
import type { GarbageReport } from '@/types'

const reports = ref<GarbageReport[]>([])
const loading = ref(true)
const total = ref(0)
const page = ref(1)
const pageSize = ref(10)
const statusFilter = ref('')

const statusMap: Record<string, { text: string; type: 'success' | 'warning' | 'info' | 'danger' | 'primary' }> = {
  pending: { text: '待处理', type: 'warning' },
  verified: { text: '已验证', type: 'success' },
  dismissed: { text: '已驳回', type: 'info' },
}

async function loadReports() {
  loading.value = true
  try {
    const params: Record<string, unknown> = { page: page.value, page_size: pageSize.value }
    if (statusFilter.value) params.category = statusFilter.value
    const res = await getMyReports(params as any)
    reports.value = res.data.reports || []
    total.value = res.data.pagination?.total || 0
  } catch {
    reports.value = []
  } finally {
    loading.value = false
  }
}

function handlePageChange(p: number) {
  page.value = p
  loadReports()
}

function handleFilter(status: string) {
  statusFilter.value = status === 'all' ? '' : status
  page.value = 1
  loadReports()
}

function formatTime(t: string) {
  if (!t) return ''
  return new Date(t).toLocaleString('zh-CN')
}

onMounted(loadReports)
</script>

<template>
  <div class="history-page">
    <h2>上报记录</h2>
    <p class="page-desc">查看所有垃圾分类上报历史记录</p>

    <div class="filter-bar">
      <span
        v-for="f in [{ label: '全部', value: '' }, { label: '已验证', value: 'verified' }, { label: '待处理', value: 'pending' }, { label: '已驳回', value: 'dismissed' }]"
        :key="f.value"
        class="filter-tag"
        :class="{ active: statusFilter === f.value }"
        @click="handleFilter(f.value)"
      >{{ f.label }}</span>
    </div>

    <el-empty v-if="!loading && reports.length === 0" description="暂无上报记录" />

    <div v-else class="table-wrap">
      <el-table :data="reports" stripe style="width: 100%" v-loading="loading">
        <el-table-column label="检测类型" width="120">
          <template #default="{ row }">
            {{ row.detections?.[0]?.class_name || '未知' }}
          </template>
        </el-table-column>
        <el-table-column label="分类" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.detections?.[0]?.category || '未知' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="位置" min-width="160">
          <template #default="{ row }">
            <span v-if="row.lat && row.lng">{{ row.lat.toFixed(4) }}, {{ row.lng.toFixed(4) }}</span>
            <span v-else class="no-loc">—</span>
          </template>
        </el-table-column>
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="statusMap[row.status]?.type || 'info'">
              {{ statusMap[row.status]?.text || row.status }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column label="时间" width="170">
          <template #default="{ row }">
            {{ formatTime(row.reported_at) }}
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div v-if="total > pageSize" class="pagination-wrap">
      <el-pagination
        background layout="prev, pager, next"
        :total="total"
        :page-size="pageSize"
        :current-page="page"
        @current-change="handlePageChange"
      />
    </div>
  </div>
</template>

<style scoped>
.history-page {
  max-width: 1100px;
}

.history-page h2 {
  font-size: 22px;
  color: #303133;
  margin: 0 0 8px;
}

.page-desc {
  color: #909399;
  font-size: 14px;
  margin-bottom: 20px;
}

.filter-bar {
  display: flex;
  gap: 8px;
  margin-bottom: 16px;
}

.filter-tag {
  padding: 4px 16px;
  border-radius: 16px;
  background: #f0f2f5;
  color: #606266;
  cursor: pointer;
  font-size: 13px;
  transition: all 0.2s;
  user-select: none;
}

.filter-tag.active,
.filter-tag:hover {
  background: #2c3e50;
  color: #fff;
}

.table-wrap {
  background: #fff;
  border-radius: 8px;
  overflow: hidden;
}

.pagination-wrap {
  display: flex;
  justify-content: center;
  margin-top: 20px;
}

.no-loc {
  color: #c0c4cc;
}
</style>
