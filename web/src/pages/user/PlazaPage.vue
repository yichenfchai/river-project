<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { ElDialog, ElForm, ElFormItem, ElInput, ElButton, ElPagination, ElSelect, ElOption } from 'element-plus'
import { useMessage } from '@/composables/useMessage'
import { Edit } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import { getPosts, createPost, toggleLike } from '@/api/modules/posts'
import type { Post, Pagination } from '@/types'


const msg = useMessage()

const auth = useAuthStore()

const posts = ref<Post[]>([])
const pagination = ref<Pagination>({ page: 1, page_size: 10, total: 0, total_pages: 0 })
const loading = ref(false)
const activeTopic = ref('')

const topics = [
  { label: '全部', value: '' },
  { label: '生态', value: 'ecology' },
  { label: '文化', value: 'culture' },
  { label: '历史', value: 'share' },
  { label: '问答', value: 'question' },
]

async function fetchPosts(page = 1) {
  loading.value = true
  try {
    const res = await getPosts({ page, page_size: 10, topic: activeTopic.value || undefined })
    posts.value = res.data.posts || []
    pagination.value = res.data.pagination
  } catch {
    // handled by interceptor
  } finally {
    loading.value = false
  }
}

function switchTopic(topic: string) {
  activeTopic.value = topic === activeTopic.value ? '' : topic
  fetchPosts(1)
}

async function handleLike(post: Post) {
  if (!auth.isLoggedIn) {
    msg.warning('请先登录')
    return
  }
  try {
    const res = await toggleLike(post.id)
    post.is_liked = res.data.is_liked
    post.like_count = res.data.like_count
  } catch {
    // handled
  }
}

function formatTime(ts: string) {
  const d = new Date(ts)
  const now = new Date()
  const diff = now.getTime() - d.getTime()
  if (diff < 3600000) return `${Math.floor(diff / 60000)} 分钟前`
  if (diff < 86400000) return `${Math.floor(diff / 3600000)} 小时前`
  return ts.slice(0, 10)
}

// ─── Create Dialog ───

const dialogVisible = ref(false)
const createForm = ref({ title: '', content: '', topic: 'share' })
const submitting = ref(false)

async function submitPost() {
  if (!createForm.value.title.trim() || !createForm.value.content.trim()) {
    msg.warning('标题和内容不能为空')
    return
  }
  submitting.value = true
  try {
    await createPost(createForm.value)
    msg.success('发布成功，等待审核')
    dialogVisible.value = false
    createForm.value = { title: '', content: '', topic: 'share' }
    fetchPosts(1)
  } catch {
    // handled
  } finally {
    submitting.value = false
  }
}

onMounted(() => fetchPosts())
</script>

<template>
  <div class="plaza-page">
    <div class="plaza-header">
      <h2>💬 分享广场</h2>
      <el-button type="primary" :icon="Edit" @click="dialogVisible = true" :disabled="!auth.isLoggedIn">
        {{ auth.isLoggedIn ? '发布帖子' : '请先登录' }}
      </el-button>
    </div>

    <div class="topic-filter">
      <span
        v-for="t in topics"
        :key="t.value"
        class="topic-tag"
        :class="{ active: activeTopic === t.value }"
        @click="switchTopic(t.value)"
      >{{ t.label }}</span>
    </div>

    <div class="post-list" v-loading="loading">
      <div v-if="posts.length === 0 && !loading" class="empty-hint">暂无帖子，快来第一个分享吧</div>
      <div v-for="post in posts" :key="post.id" class="post-card">
        <div class="post-author">
          <el-avatar :size="40" :src="post.author.avatar_url">{{ post.author.nickname[0] }}</el-avatar>
          <div class="author-info">
            <span class="author-name">{{ post.author.nickname || post.author.username }}</span>
            <span class="post-time">{{ formatTime(post.created_at) }}</span>
          </div>
        </div>
        <h3 class="post-title">{{ post.title }}</h3>
        <p class="post-content">{{ post.content }}</p>
        <div class="post-footer">
          <span class="footer-action" @click="handleLike(post)" :class="{ liked: post.is_liked }">
            {{ post.is_liked ? '❤️' : '🤍' }} {{ post.like_count }}
          </span>
          <span>💬 {{ post.comment_count }}</span>
        </div>
      </div>
    </div>

    <div class="pagination-wrap" v-if="pagination.total_pages > 1">
      <el-pagination background layout="prev, pager, next" :total="pagination.total" :page-size="10" @current-change="fetchPosts" />
    </div>

    <!-- Create Dialog -->
    <el-dialog v-model="dialogVisible" title="发布帖子" width="500px">
      <el-form :model="createForm" label-width="60px">
        <el-form-item label="标题">
          <el-input v-model="createForm.title" maxlength="200" placeholder="帖子标题" />
        </el-form-item>
        <el-form-item label="话题">
          <el-select v-model="createForm.topic" style="width:100%">
            <el-option label="分享" value="share" />
            <el-option label="生态" value="ecology" />
            <el-option label="文化" value="culture" />
            <el-option label="问答" value="question" />
            <el-option label="其他" value="other" />
          </el-select>
        </el-form-item>
        <el-form-item label="内容">
          <el-input v-model="createForm.content" type="textarea" :rows="6" maxlength="10000" show-word-limit placeholder="分享你的运河故事..." />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitPost" :loading="submitting">发布</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.plaza-page { max-width:800px;margin:0 auto }
.plaza-header { display:flex;align-items:center;justify-content:space-between }
.plaza-header h2 { font-family:'Noto Serif SC','STSong',serif;font-size:22px;color:var(--text-primary);margin:0 }
.topic-filter { display:flex;gap:8px;margin:16px 0 20px;flex-wrap:wrap }
.topic-tag { padding:10px 18px;border-radius:16px;background:var(--paper-light);color:var(--text-secondary);font-size:13px;cursor:pointer;transition:.2s;user-select:none;min-height:44px;display:inline-flex;align-items:center;border:1px solid transparent }
.topic-tag.active,.topic-tag:hover { background:var(--ink-primary);color:#fff;border-color:var(--ink-primary) }
.post-list { display:flex;flex-direction:column;gap:16px }
.post-card { background:var(--surface-card);border-radius:var(--radius-lg);padding:20px;box-shadow:var(--shadow-sm);cursor:pointer;transition:all .25s var(--ease-ink);border:1px solid var(--border-hairline) }
.post-card:hover { box-shadow:var(--shadow-md);transform:translateY(-2px);border-color:var(--gold) }
.post-author { display:flex;align-items:center;gap:10px;margin-bottom:12px }
.author-info { display:flex;flex-direction:column }
.author-name { font-size:14px;color:var(--text-primary);font-weight:500 }
.post-time { font-size:12px;color:var(--text-muted) }
.post-title { font-size:18px;margin:0 0 8px;color:var(--text-primary) }
.post-content { color:var(--text-body);font-size:14px;line-height:1.6;margin:0 0 16px }
.post-footer { display:flex;gap:20px;color:var(--text-muted);font-size:13px }
.footer-action { cursor:pointer;user-select:none;transition:color .2s }
.footer-action:hover { color:var(--gold-dark) }
.footer-action.liked { color:var(--cinnabar) }
.empty-hint { text-align:center;color:var(--text-muted);padding:40px 0 }
.pagination-wrap { margin-top:20px;display:flex;justify-content:center }

@media (max-width: 768px) {
  .plaza-page { padding: 0 8px }
  .post-card { padding: 16px }
  .post-title { font-size: 16px }
}

@media (max-width: 480px) {
  .plaza-header { flex-direction: column; gap: 12px; align-items: flex-start }
  .plaza-header h2 { font-size: 20px }
  .post-card { padding: 14px }
}
</style>
