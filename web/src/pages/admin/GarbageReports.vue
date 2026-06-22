<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElTable, ElTableColumn, ElTag, ElImage, ElPagination } from 'element-plus'
import { getAllGarbageReports } from '@/api/modules/admin'
import type { GarbageReport, Pagination } from '@/types'

const reports = ref<GarbageReport[]>([])
const pagination = ref<Pagination>({ page: 1, page_size: 10, total: 0, total_pages: 0 })
const loading = ref(false)

async function fetchReports(page = 1) {
  loading.value = true
  try {
    const res = await getAllGarbageReports({ page, page_size: 10 })
    reports.value = res.data.reports
    pagination.value = res.data.pagination
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

onMounted(() => fetchReports())
</script>

<template>
  <div class="garbage-reports">
    <h2>🗑 垃圾上报汇总</h2>

    <div class="table-wrap">
      <el-table :data="reports" v-loading="loading" stripe>
        <el-table-column label="分类" min-width="160">
          <template #default="{ row }">
            <div v-if="row.detections?.[0]" class="detection-info">
              <span class="detection-name">{{ row.detections[0].class_name }}</span>
              <el-tag size="small">{{ row.detections[0].category }}</el-tag>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="位置" min-width="160">
          <template #default="{ row }">
            <span v-if="row.lat">{{ row.lat.toFixed(4) }}, {{ row.lng?.toFixed(4) }}</span>
            <span v-else class="no-data">未记录</span>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 'verified' ? 'success' : row.status === 'dismissed' ? 'info' : 'warning'">
              {{ row.status === 'verified' ? '已验证' : row.status === 'dismissed' ? '已驳回' : '待处理' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="reported_at" label="上报时间" width="180" />
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="pagination.total_pages > 1">
      <el-pagination
        background
        layout="prev, pager, next"
        :total="pagination.total"
        :page-size="pagination.page_size"
        :current-page="pagination.page"
        @current-change="fetchReports"
      />
    </div>
  </div>
</template>

<style scoped>
.garbage-reports {
  max-width: 1000px;
}

.garbage-reports h2 {
  font-size: 22px;
  color: #303133;
  margin: 0 0 20px;
}

.table-wrap {
  background: #fff;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.detection-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.detection-name {
  font-weight: 500;
  color: #303133;
}

.no-data {
  color: #c0c4cc;
}

.pagination-wrap {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>
