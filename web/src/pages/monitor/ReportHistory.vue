<script setup lang="ts">
import { ElTable, ElTableColumn, ElTag, ElImage, ElPagination } from 'element-plus'
</script>

<template>
  <div class="history-page">
    <h2>📋 上报记录</h2>
    <p class="page-desc">查看所有垃圾分类上报历史记录</p>

    <div class="filter-bar">
      <span v-for="f in filters" :key="f.value" class="filter-tag" :class="{ active: f.value === 'all' }">{{ f.label }}</span>
    </div>

    <div class="table-wrap">
      <el-table :data="reports" stripe style="width: 100%">
        <el-table-column label="图片" width="80">
          <template #default="{ row }">
            <div class="thumb" :style="{ background: row.bgColor }">{{ row.emoji }}</div>
          </template>
        </el-table-column>
        <el-table-column prop="type" label="类型" width="100" />
        <el-table-column prop="category" label="分类" width="100">
          <template #default="{ row }">
            <el-tag size="small">{{ row.category }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="location" label="位置" />
        <el-table-column prop="status" label="状态" width="100">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 'verified' ? 'success' : row.status === 'dismissed' ? 'info' : 'warning'">
              {{ row.status === 'verified' ? '已验证' : row.status === 'dismissed' ? '已驳回' : '待处理' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="time" label="时间" width="160" />
      </el-table>
    </div>

    <div class="pagination-wrap">
      <el-pagination background layout="prev, pager, next" :total="50" />
    </div>
  </div>
</template>

<script lang="ts">
const filters = [
  { label: '全部', value: 'all' },
  { label: '已验证', value: 'verified' },
  { label: '待处理', value: 'pending' },
  { label: '已驳回', value: 'dismissed' },
]
const reports = [
  { emoji: '🧴', bgColor: '#e6f7ff', type: '塑料瓶', category: '可回收物', location: '扬州市广陵区运河西路', status: 'verified', time: '2026-06-20 14:30' },
  { emoji: '🛍', bgColor: '#fff7e6', type: '塑料袋', category: '其他垃圾', location: '扬州市邗江区文昌中路', status: 'pending', time: '2026-06-20 13:15' },
  { emoji: '🔋', bgColor: '#fff1f0', type: '废电池', category: '有害垃圾', location: '扬州市广陵区泰州路', status: 'verified', time: '2026-06-20 11:00' },
  { emoji: '🍾', bgColor: '#e6f7ff', type: '玻璃瓶', category: '可回收物', location: '扬州市邗江区大学北路', status: 'pending', time: '2026-06-20 10:20' },
  { emoji: '🥬', bgColor: '#f6ffed', type: '厨余垃圾', category: '厨余垃圾', location: '扬州市广陵区东关街', status: 'verified', time: '2026-06-19 16:45' },
]
</script>

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
}

.filter-tag.active,
.filter-tag:hover {
  background: #2c3e50;
  color: #fff;
}

.thumb {
  width: 44px;
  height: 44px;
  border-radius: 6px;
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 22px;
}

.table-wrap {
  background: #fff;
  border-radius: 10px;
  overflow: hidden;
  box-shadow: 0 1px 4px rgba(0, 0, 0, 0.04);
}

.pagination-wrap {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}
</style>
