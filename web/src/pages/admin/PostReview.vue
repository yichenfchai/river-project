<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElTable, ElTableColumn, ElTag, ElButton, ElPagination, ElMessage } from 'element-plus'
import { getPendingPosts, reviewPost } from '@/api/modules/admin'
import type { Post, Pagination } from '@/types'

const posts = ref<Post[]>([])
const pagination = ref<Pagination>({ page: 1, page_size: 10, total: 0, total_pages: 0 })
const loading = ref(false)

async function fetchPosts(page = 1) {
  loading.value = true
  try {
    const res = await getPendingPosts({ page, page_size: 10 })
    posts.value = res.data.posts
    pagination.value = res.data.pagination
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

async function handleReview(row: Post, action: 'approve' | 'reject') {
  try {
    await reviewPost(row.id, { action, reason: action === 'reject' ? '不符合社区规范' : undefined })
    row.status = action === 'approve' ? 'approved' : 'rejected'
    ElMessage.success(action === 'approve' ? '已通过审核' : '已驳回')
  } catch {
    // handled by interceptor
  }
}

onMounted(() => fetchPosts())
</script>

<template>
  <div class="post-review">
    <h2>📝 帖子审核</h2>

    <div class="table-wrap">
      <el-table :data="posts" v-loading="loading" stripe>
        <el-table-column prop="title" label="标题" min-width="200" />
        <el-table-column label="作者" width="120">
          <template #default="{ row }">{{ row.author?.nickname || row.author?.username }}</template>
        </el-table-column>
        <el-table-column prop="topic" label="话题" width="80">
          <template #default="{ row }">
            <el-tag size="small">{{ row.topic }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="{ row }">
            <el-tag size="small" :type="row.status === 'approved' ? 'success' : row.status === 'rejected' ? 'danger' : 'warning'">
              {{ row.status === 'approved' ? '已通过' : row.status === 'rejected' ? '已驳回' : '待审核' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="created_at" label="发布时间" width="180" />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <template v-if="(row as Post).status === 'pending'">
              <el-button size="small" type="success" @click="handleReview(row as Post, 'approve')">通过</el-button>
              <el-button size="small" type="danger" @click="handleReview(row as Post, 'reject')">驳回</el-button>
            </template>
            <span v-else class="reviewed-text">已处理</span>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <div class="pagination-wrap" v-if="pagination.total_pages > 1">
      <el-pagination
        background
        layout="prev, pager, next"
        :total="pagination.total"
        :page-size="pagination.page_size"
        :current-page="pagination.page"
        @current-change="fetchPosts"
      />
    </div>
  </div>
</template>

<style scoped>
.post-review {
  max-width: 1200px;
}

.post-review h2 {
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

.pagination-wrap {
  margin-top: 20px;
  display: flex;
  justify-content: center;
}

.reviewed-text {
  color: #909399;
  font-size: 13px;
}
</style>
