<template>
  <div class="profile-page" v-if="auth.user">
    <h2>个人中心</h2>
    <el-card>
      <el-form :model="form" label-width="80px">
        <el-form-item label="用户名"><el-input v-model="auth.user.username" disabled /></el-form-item>
        <el-form-item label="昵称"><el-input v-model="form.nickname" /></el-form-item>
        <el-form-item label="邮箱"><el-input v-model="auth.user.email" disabled /></el-form-item>
        <el-form-item label="简介"><el-input v-model="form.bio" type="textarea" /></el-form-item>
        <el-form-item label="角色"><el-tag>{{ auth.user.role }}</el-tag></el-form-item>
        <el-form-item label="积分">{{ stats?.total_score || 0 }}</el-form-item>
        <el-form-item label="段位"><el-tag type="warning">{{ stats?.rank_title || '青铜守护者' }}</el-tag></el-form-item>
        <el-form-item>
          <el-button type="primary" @click="save" :loading="saving">保存</el-button>
        </el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref, onMounted } from 'vue'
import { useAuthStore } from '@/stores/auth'
import { authApi } from '@/api/auth'
import { quizApi } from '@/api/quiz'
import { ElMessage } from 'element-plus'
import type { UserStats } from '@/types'

const auth = useAuthStore()
const saving = ref(false)
const stats = ref<UserStats | null>(null)

const form = reactive({ nickname: '', bio: '' })

async function save() {
  saving.value = true
  try {
    await authApi.updateProfile({ nickname: form.nickname, bio: form.bio })
    await auth.fetchProfile()
    ElMessage.success('保存成功')
  } catch { ElMessage.error('保存失败') }
  finally { saving.value = false }
}

onMounted(async () => {
  form.nickname = auth.user?.nickname || ''
  form.bio = auth.user?.bio || ''
  try {
    const res = await quizApi.getUserStats(auth.user!.id)
    stats.value = res.data.data!
  } catch { /* ignore */ }
})
</script>

<style scoped>
.profile-page { max-width: 600px; margin: 0 auto; }
</style>
