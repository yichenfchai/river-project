<template>
  <div class="post-detail" v-loading="loading">
    <template v-if="post">
      <el-button text @click="$router.back()" style="margin-bottom:16px">← 返回</el-button>
      <div class="post-content">
        <h2>{{ post.title }}</h2>
        <div class="post-info">
          <el-tag size="small">{{ topicLabels[post.topic] || post.topic }}</el-tag>
          <span>{{ formatTime(post.created_at) }}</span>
        </div>
        <div class="body" v-html="post.content.replace(/\n/g, '<br>')"></div>
        <div class="tags" v-if="post.tags?.length">
          <el-tag v-for="tag in post.tags" :key="tag" size="small" type="info">{{ tag }}</el-tag>
        </div>
        <div class="actions">
          <el-button :type="liked ? 'danger' : 'default'" @click="toggleLike" :disabled="!auth.isLoggedIn">
            <el-icon><Star /></el-icon> {{ post.like_count }}
          </el-button>
          <el-button v-if="auth.isLoggedIn && auth.user?.id === post.author_id" text type="danger" @click="handleDelete">
            删除
          </el-button>
        </div>
      </div>

      <el-divider />

      <div class="comments">
        <h3>评论 ({{ post.comment_count }})</h3>
        <div v-if="auth.isLoggedIn" class="comment-input">
          <el-input v-model="commentText" placeholder="写下你的评论..." type="textarea" :rows="3" />
          <el-button type="primary" @click="sendComment" :loading="sending" style="margin-top:8px">发表</el-button>
        </div>
        <div v-for="c in comments" :key="c.id" class="comment-item">
          <div class="comment-header">
            <span class="author">{{ c.author_id?.slice(0,8) }}...</span>
            <span class="time">{{ formatTime(c.created_at) }}</span>
          </div>
          <p>{{ c.content }}</p>
        </div>
      </div>
    </template>
    <el-empty v-else description="帖子不存在" />
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { postsApi } from '@/api/posts'
import { useAuthStore } from '@/stores/auth'
import { ElMessage, ElMessageBox } from 'element-plus'
import type { Post, Comment } from '@/types'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const post = ref<Post | null>(null)
const comments = ref<Comment[]>([])
const commentText = ref('')
const liked = ref(false)
const loading = ref(true)
const sending = ref(false)

const topicLabels: Record<string, string> = {
  share: '分享', ecology: '生态', culture: '文化', question: '问答',
}

function formatTime(t: string) { return new Date(t).toLocaleString('zh-CN') }

async function fetchPost() {
  loading.value = true
  try {
    const res = await postsApi.get(route.params.id as string)
    post.value = res.data.data!
    const cRes = await postsApi.listComments(post.value!.id)
    comments.value = cRes.data.data!.items
  } catch { post.value = null }
  finally { loading.value = false }
}

async function toggleLike() {
  if (!post.value) return
  try {
    const res = await postsApi.toggleLike(post.value.id)
    liked.value = res.data.data!.liked
    post.value.like_count += liked.value ? 1 : -1
  } catch { /* ignore */ }
}

async function sendComment() {
  if (!commentText.value.trim() || !post.value) return
  sending.value = true
  try {
    const res = await postsApi.createComment(post.value.id, commentText.value)
    comments.value.unshift(res.data.data!)
    post.value.comment_count++
    commentText.value = ''
    ElMessage.success('评论成功')
  } catch { ElMessage.error('评论失败') }
  finally { sending.value = false }
}

async function handleDelete() {
  await ElMessageBox.confirm('确定删除这条帖子？', '提示', { type: 'warning' })
  await postsApi.delete(post.value!.id)
  ElMessage.success('已删除')
  router.push('/posts')
}

onMounted(fetchPost)
</script>

<style scoped>
.post-detail { max-width: 800px; margin: 0 auto; }
.post-content h2 { margin-bottom: 12px; }
.post-info { display: flex; gap: 12px; align-items: center; color: #909399; font-size: 13px; margin-bottom: 16px; }
.body { line-height: 1.8; white-space: pre-wrap; margin-bottom: 16px; }
.tags { display: flex; gap: 8px; margin-bottom: 16px; }
.actions { display: flex; gap: 8px; }
.comments { margin-top: 20px; }
.comment-input { margin-bottom: 20px; }
.comment-item { padding: 12px 0; border-bottom: 1px solid #ebeef5; }
.comment-header { display: flex; justify-content: space-between; color: #909399; font-size: 13px; margin-bottom: 4px; }
.comment-item p { margin: 0; line-height: 1.6; }
</style>
