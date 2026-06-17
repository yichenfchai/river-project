<template>
  <div class="posts-page">
    <div class="page-header">
      <h2>分享广场</h2>
      <el-button type="primary" @click="$router.push('/posts/create')" v-if="auth.isLoggedIn">
        发布帖子
      </el-button>
    </div>

    <div class="filters">
      <el-radio-group v-model="topic" @change="fetchPosts">
        <el-radio-button value="all">全部</el-radio-button>
        <el-radio-button value="share">分享</el-radio-button>
        <el-radio-button value="ecology">生态</el-radio-button>
        <el-radio-button value="culture">文化</el-radio-button>
        <el-radio-button value="question">问答</el-radio-button>
      </el-radio-group>
      <el-input v-model="keyword" placeholder="搜索..." clearable @clear="fetchPosts" @keyup.enter="fetchPosts" style="width:240px">
        <template #prefix><el-icon><Search /></el-icon></template>
      </el-input>
    </div>

    <div v-loading="loading">
      <el-empty v-if="!loading && posts.length === 0" description="暂无帖子" />
      <div v-for="post in posts" :key="post.id" class="post-card" @click="$router.push(`/posts/${post.id}`)">
        <h3>{{ post.title }}</h3>
        <p class="post-preview">{{ post.content.slice(0, 200) }}{{ post.content.length > 200 ? '...' : '' }}</p>
        <div class="post-meta">
          <el-tag size="small" v-if="post.topic">{{ topicLabels[post.topic] || post.topic }}</el-tag>
          <span class="meta-item"><el-icon><Star /></el-icon> {{ post.like_count }}</span>
          <span class="meta-item"><el-icon><ChatLineSquare /></el-icon> {{ post.comment_count }}</span>
          <span class="meta-item">{{ formatTime(post.created_at) }}</span>
        </div>
      </div>
    </div>

    <el-pagination
      v-if="total > 0"
      v-model:current-page="page"
      :page-size="pageSize"
      :total="total"
      layout="prev, pager, next"
      @current-change="fetchPosts"
      class="pagination"
    />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { postsApi } from '@/api/posts'
import { useAuthStore } from '@/stores/auth'
import type { Post } from '@/types'

const auth = useAuthStore()
const posts = ref<Post[]>([])
const loading = ref(false)
const page = ref(1)
const pageSize = 20
const total = ref(0)
const topic = ref('all')
const keyword = ref('')

const topicLabels: Record<string, string> = {
  share: '分享', ecology: '生态', culture: '文化', question: '问答',
}

function formatTime(t: string) {
  return new Date(t).toLocaleDateString('zh-CN')
}

async function fetchPosts() {
  loading.value = true
  try {
    const res = await postsApi.list({
      page: page.value, page_size: pageSize,
      topic: topic.value === 'all' ? undefined : topic.value,
      keyword: keyword.value || undefined,
    })
    const data = res.data.data!
    posts.value = data.items
    total.value = data.pagination.total
  } catch { /* ignore */ }
  finally { loading.value = false }
}

onMounted(fetchPosts)
</script>

<style scoped>
.posts-page { max-width: 800px; margin: 0 auto; }
.page-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 16px; }
.filters { display: flex; justify-content: space-between; align-items: center; margin-bottom: 20px; }
.post-card { background: #fff; border-radius: 8px; padding: 20px; margin-bottom: 12px; cursor: pointer; transition: box-shadow .2s; }
.post-card:hover { box-shadow: 0 2px 12px rgba(0,0,0,.1); }
.post-card h3 { margin: 0 0 8px; font-size: 18px; }
.post-preview { color: #606266; line-height: 1.6; margin: 0 0 12px; }
.post-meta { display: flex; align-items: center; gap: 16px; color: #909399; font-size: 13px; }
.meta-item { display: flex; align-items: center; gap: 4px; }
.pagination { margin-top: 20px; justify-content: center; }
</style>
