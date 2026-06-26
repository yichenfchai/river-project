<script setup lang="ts">
import { ref } from 'vue'
import { ElForm, ElFormItem, ElInput, ElButton, ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import { updateProfile } from '@/api/modules/auth'

const auth = useAuthStore()

const form = ref({
  nickname: auth.user?.nickname || '',
  bio: '',
})

async function handleSave() {
  try {
    await updateProfile({ nickname: form.value.nickname, bio: form.value.bio })
    await auth.fetchProfile()
    ElMessage.success('资料已更新')
  } catch {
    ElMessage.error('保存失败')
  }
}
</script>

<template>
  <div class="profile-page">
    <h2>👤 个人中心</h2>

    <div class="profile-card">
      <div class="profile-header">
        <el-avatar :size="72" :src="auth.user?.avatar_url" />
        <div class="profile-name">
          <h3>{{ auth.user?.nickname || auth.user?.username }}</h3>
          <el-tag size="small" type="warning">监测人员</el-tag>
        </div>
      </div>

      <ElForm :model="form" label-width="80px" class="profile-form">
        <ElFormItem label="用户名">
          <ElInput :model-value="auth.user?.username" disabled />
        </ElFormItem>
        <ElFormItem label="邮箱">
          <ElInput :model-value="auth.user?.email" disabled />
        </ElFormItem>
        <ElFormItem label="昵称">
          <ElInput v-model="form.nickname" placeholder="请输入昵称" />
        </ElFormItem>
        <ElFormItem label="简介">
          <ElInput v-model="form.bio" type="textarea" :rows="3" placeholder="介绍一下自己" />
        </ElFormItem>
        <ElFormItem>
          <ElButton type="primary" @click="handleSave">保存修改</ElButton>
        </ElFormItem>
      </ElForm>
    </div>
  </div>
</template>

<style scoped>
.profile-page {
  max-width: 600px;
}

.profile-page h2 {
  font-size: 22px;
  color: #303133;
  margin: 0 0 20px;
}

.profile-card {
  background: #fff;
  border-radius: 12px;
  padding: 32px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.06);
}

.profile-header {
  display: flex;
  align-items: center;
  gap: 16px;
  margin-bottom: 28px;
  padding-bottom: 20px;
  border-bottom: 1px solid #f0f2f5;
}

.profile-name h3 {
  margin: 0 0 6px;
  color: #303133;
}

.profile-form {
  max-width: 400px;
}
</style>
