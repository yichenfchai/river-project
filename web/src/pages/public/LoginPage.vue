<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useMessage } from '@/composables/useMessage'
import { User, Monitor, Setting, UserFilled } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'
import type { UserRole } from '@/types'


const msg = useMessage()

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const activeTab = ref<UserRole>('user')
const form = ref({ username: '', password: '' })
const loggingIn = ref(false)
const showForgotDialog = ref(false)
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
    msg.success(`欢迎回来，${user.nickname || user.username}`)
    if (role === 'admin') {
      router.push('/admin/dashboard')
    } else if (role === 'monitor') {
      router.push('/monitor/dashboard')
    } else {
      router.push('/home')
    }
  } catch {
    // 错误消息已由 API 拦截器处理，此处无需重复提示
  } finally {
    loggingIn.value = false
  }
}

function handleGuestLogin() {
  auth.loginAsGuest()
  msg.success('已进入游客模式，登录后可解锁全部功能')
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
              autocomplete="username"
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
              autocomplete="current-password"
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

        <div class="forgot-line">
          <span class="forgot-hint" @click="showForgotDialog = true">忘记密码？</span>
        </div>
        <div class="login-footer">
          <span>没有账号？</span>
          <router-link to="/register" class="register-link">立即注册</router-link>
        </div>
      </el-card>
    </div>
  </div>

  <el-dialog v-model="showForgotDialog" title="找回密码" width="400px" :close-on-click-modal="false">
    <p style="color: #606266; line-height: 1.8;">
      请联系<span style="font-weight: 600; color: #303133;">管理员</span>重置密码。
    </p>
    <p style="color: #909399; font-size: 13px; margin-top: 8px;">
      管理员可在后台 → 用户管理 → 重置密码
    </p>
    <template #footer>
      <el-button type="primary" @click="showForgotDialog = false">知道了</el-button>
    </template>
  </el-dialog>
</template>

<style scoped>
.login-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(160deg, #0f0f1e 0%, #1a1a2e 25%, #1d2b3a 50%, #1a2a28 75%, #1a1a2e 100%);
}
/* 水墨背景层 */
.login-page::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 20% 80%, rgba(201,184,150,0.06) 0%, transparent 50%),
    radial-gradient(ellipse at 80% 20%, rgba(201,184,150,0.04) 0%, transparent 50%),
    radial-gradient(ellipse at 60% 60%, rgba(44,62,80,0.3) 0%, transparent 70%);
  pointer-events: none;
}
/* 远山装饰 */
.login-page::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 35vh;
  background:
    /* 远山层 */
    radial-gradient(ellipse 180% 100% at 15% 100%, rgba(44,62,80,0.5) 0%, transparent 70%),
    radial-gradient(ellipse 160% 100% at 50% 100%, rgba(26,26,46,0.6) 0%, transparent 65%),
    radial-gradient(ellipse 140% 100% at 85% 100%, rgba(44,62,80,0.4) 0%, transparent 60%);
  pointer-events: none;
}

.login-bg {
  position: absolute;
  inset: 0;
  background:
    url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1440 200'%3E%3Cpath fill='rgba(201,184,150,0.04)' d='M0,64L48,69.3C96,75,192,85,288,80C384,75,480,53,576,48C672,43,768,53,864,69.3C960,85,1056,107,1152,101.3C1248,96,1344,64,1392,48L1440,32L1440,0L1392,0C1344,0,1248,0,1152,0C1056,0,960,0,864,0C768,0,672,0,576,0C480,0,384,0,288,0C192,0,96,0,48,0L0,0Z'/%3E%3C/svg%3E") no-repeat bottom;
  background-size: cover;
  pointer-events: none;
  opacity: 0.6;
}

.login-container {
  position: relative;
  z-index: 1;
  width: 460px;
  max-width: 92vw;
  animation: fadeInUp 0.7s var(--ease-ink);
}

.login-header {
  text-align: center;
  margin-bottom: 32px;
}

.login-title {
  font-family: 'Ma Shan Zheng', 'ZCOOL XiaoWei', 'STKaiti', 'KaiTi', cursive;
  font-size: 30px;
  color: var(--gold-light);
  margin: 0 0 8px;
  letter-spacing: 4px;
  font-weight: 400;
  text-shadow: 0 2px 16px rgba(0,0,0,0.3);
}

.login-subtitle {
  color: rgba(255,255,255,0.5);
  font-size: 13px;
  margin: 0;
  letter-spacing: 2px;
  text-transform: uppercase;
}

.login-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-hairline);
  box-shadow: var(--shadow-lg);
  overflow: visible;
  position: relative;
}
.login-card::before {
  content: '';
  position: absolute;
  top: 0;
  left: 24px;
  right: 24px;
  height: 2px;
  background: linear-gradient(90deg, transparent, var(--gold), transparent);
  opacity: 0.6;
  z-index: 1;
}

.login-card :deep(.el-card__body) {
  padding: 28px 36px 36px;
}

.role-tabs :deep(.el-tabs__header) {
  margin-bottom: 14px;
}
.role-tabs :deep(.el-tabs__nav-wrap::after) {
  height: 1px;
  background: var(--border-light);
}
.role-tabs :deep(.el-tabs__item) {
  font-size: 14px;
  color: var(--text-secondary);
  transition: color 0.25s;
}
.role-tabs :deep(.el-tabs__item.is-active) {
  color: var(--ink-primary);
  font-weight: 600;
}
.role-tabs :deep(.el-tabs__active-bar) {
  background: var(--gold-dark);
  height: 2px;
}

.role-info {
  display: flex;
  align-items: center;
  gap: 8px;
  color: var(--text-secondary);
  font-size: 13px;
  margin-bottom: 24px;
  padding: 10px 14px;
  background: var(--paper-light);
  border-radius: var(--radius-md);
  border-left: 3px solid var(--gold);
}

.login-btn {
  width: 100%;
  letter-spacing: 6px;
  font-size: 16px;
  font-weight: 600;
  height: 46px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--ink-primary) 0%, #243445 100%);
  border: none;
  transition: all 0.3s var(--ease-ink);
}
.login-btn:hover {
  background: linear-gradient(135deg, var(--ink-primary-light) 0%, #2c3e50 100%);
  box-shadow: 0 6px 24px rgba(44,62,80,0.35);
  transform: translateY(-2px);
}

.login-footer {
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
}

.guest-btn {
  width: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 6px;
  border: 1px dashed var(--border-medium);
  color: var(--text-secondary);
  background: var(--paper-light);
  height: 46px;
  border-radius: var(--radius-md);
  transition: all 0.25s;
}
.guest-btn:hover {
  border-color: var(--gold);
  color: var(--gold-dark);
  background: #faf6ed;
}

.forgot-line {
  text-align: center;
  margin-bottom: 14px;
}

.forgot-hint {
  color: var(--text-muted);
  font-size: 13px;
  cursor: pointer;
  transition: color 0.2s;
}
.forgot-hint:hover {
  color: var(--ink-primary);
  text-decoration: underline;
}

.register-link {
  color: var(--ink-primary);
  text-decoration: none;
  font-weight: 600;
  margin-left: 4px;
  transition: color 0.2s;
}
.register-link:hover {
  color: var(--gold-dark);
  text-decoration: underline;
}

/* Tablet */
@media (max-width: 768px) {
  .login-container { width: 400px; }
  .login-card :deep(.el-card__body) { padding: 20px 24px 28px; }
}

@media (max-width: 480px) {
  .login-title { font-size: 24px; letter-spacing: 2px; }
  .login-subtitle { font-size: 11px; }
  .login-card :deep(.el-card__body) { padding: 16px 16px 22px; }
  .login-card::before { left: 16px; right: 16px; }
  .role-tabs :deep(.el-tabs__item) { font-size: 12px; padding: 0 8px; }
  .login-btn { font-size: 14px; letter-spacing: 3px; }
  .guest-btn { font-size: 13px; }
}
</style>
