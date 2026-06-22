<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMessage } from 'element-plus'
import { User, Monitor, Setting, UserFilled } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { UserRole } from '@/types'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const activeTab = ref<UserRole>('user')
const form = ref({ username: '', password: '' })
const loggingIn = ref(false)
const rules = {
  username: [{ required: true, message: '请输入用户名', trigger: 'blur' }],
  password: [{ required: true, message: '请输入密码', trigger: 'blur', min: 6 }],
}

const roleTabs = [
  { label: '普通用户', value: 'user' as UserRole, icon: User, desc: '浏览、互动、问答' },
  { label: '监测人员', value: 'monitor' as UserRole, icon: Monitor, desc: '垃圾监测与上报' },
  { label: '管理员', value: 'admin' as UserRole, icon: Setting, desc: '审核、管理、运维' },
]

const currentTab = computed(() => roleTabs.find((t) => t.value === activeTab.value))

async function handleLogin() {
  loggingIn.value = true
  try {
    const user = await auth.login(form.value.username, form.value.password)
    const role = user.role
    ElMessage.success(`欢迎回来，${user.nickname || user.username}`)
    if (role === 'admin') {
      router.push('/admin/dashboard')
    } else if (role === 'monitor') {
      router.push('/monitor/dashboard')
    } else {
      router.push('/home')
    }
  } catch {
    ElMessage.error('用户名或密码错误')
  } finally {
    loggingIn.value = false
  }
}

function handleGuestLogin() {
  auth.loginAsGuest()
  ElMessage.success('已进入游客模式，登录后可解锁全部功能')
  router.push('/home')
}
</script>

<template>
  <div class="login-page">
    <div class="login-bg"></div>
    <div class="login-container">
      <div class="login-header">
        <h1 class="login-title">大运河生态与文化保护平台</h1>
        <p class="login-subtitle">Grand Canal Guardian</p>
      </div>

      <el-card class="login-card" shadow="always">
        <el-tabs v-model="activeTab" class="role-tabs">
          <el-tab-pane v-for="t in roleTabs" :key="t.value" :label="t.label" :name="t.value" />
        </el-tabs>

        <div class="role-info" v-if="currentTab">
          <el-icon :size="20"><component :is="currentTab.icon" /></el-icon>
          <span>{{ currentTab.desc }}</span>
        </div>

        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          @submit.prevent="handleLogin"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="form.username"
              placeholder="请输入用户名"
              prefix-icon="User"
              size="large"
            />
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="请输入密码"
              prefix-icon="Lock"
              show-password
              size="large"
              @keyup.enter="handleLogin"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="loggingIn"
              class="login-btn"
              @click="handleLogin"
            >
              {{ loggingIn ? '登录中...' : '登  录' }}
            </el-button>
          </el-form-item>
          <el-form-item>
            <el-button
              size="large"
              class="guest-btn"
              @click="handleGuestLogin"
            >
              <el-icon><UserFilled /></el-icon>
              游客登录
            </el-button>
          </el-form-item>
        </el-form>

        <div class="login-footer">
          <span>没有账号？</span>
          <el-link type="primary" :underline="false">立即注册</el-link>
        </div>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(135deg, #1a2232 0%, #2c3e50 30%, #3d4f3e 70%, #2c3e50 100%);
}

.login-bg {
  position: absolute;
  inset: 0;
  background: url('data:image/svg+xml,<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1440 320"><path fill="rgba(255,255,255,0.05)" d="M0,96L48,112C96,128,192,160,288,186.7C384,213,480,235,576,224C672,213,768,171,864,149.3C960,128,1056,128,1152,149.3C1248,171,1344,213,1392,234.7L1440,256L1440,320L1392,320C1344,320,1248,320,1152,320C1056,320,960,320,864,320C768,320,672,320,576,320C480,320,384,320,288,320C192,320,96,320,48,320L0,320Z"/></svg>') no-repeat bottom;
  background-size: cover;
  pointer-events: none;
}

.login-container {
  position: relative;
  z-index: 1;
  width: 440px;
  max-width: 90vw;
}

.login-header {
  text-align: center;
  margin-bottom: 28px;
}

.login-title {
  font-size: 26px;
  color: #fff;
  margin: 0 0 6px;
  letter-spacing: 2px;
}

.login-subtitle {
  color: rgba(255, 255, 255, 0.7);
  font-size: 14px;
  margin: 0;
  letter-spacing: 1px;
}

.login-card {
  border-radius: 12px;
}

.login-card :deep(.el-card__body) {
  padding: 24px 32px 32px;
}

.role-tabs :deep(.el-tabs__header) {
  margin-bottom: 12px;
}

.role-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
}

.role-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #909399;
  font-size: 13px;
  margin-bottom: 20px;
  padding: 8px 12px;
  background: #f2efe8;
  border-radius: 6px;
}

.login-btn {
  width: 100%;
  letter-spacing: 4px;
  font-size: 16px;
}

.login-footer {
  text-align: center;
  color: #909399;
  font-size: 13px;
}

.guest-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px dashed #c0c4cc;
  color: #606266;
}

.guest-btn:hover {
  border-color: #8b7355;
  color: #8b7355;
}
</style>
