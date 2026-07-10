<script setup lang="ts">
import { ref, reactive } from 'vue'
import { useRouter } from 'vue-router'
import { useMessage } from '@/composables/useMessage'
import { User, Lock, Message, UserFilled } from '@element-plus/icons-vue'
import type { FormInstance, FormRules } from 'element-plus'
import { register as registerApi } from '@/api/modules/auth'
import { useAuthStore } from '@/stores/auth'


const msg = useMessage()

const router = useRouter()
const auth = useAuthStore()
const formRef = ref<FormInstance>()
const registering = ref(false)

const form = reactive({
  username: '',
  password: '',
  confirmPassword: '',
  email: '',
  nickname: '',
  role: 'user' as 'user' | 'monitor',
})

const validateConfirm = (_rule: unknown, value: string, callback: (err?: Error) => void) => {
  if (value !== form.password) {
    callback(new Error('两次输入的密码不一致'))
  } else {
    callback()
  }
}

const rules: FormRules = {
  username: [
    { required: true, message: '请输入用户名', trigger: 'blur' },
    { min: 3, max: 32, message: '用户名长度为 3-32 个字符', trigger: 'blur' },
    { pattern: /^[a-zA-Z0-9_\u4e00-\u9fa5]+$/, message: '用户名只能包含字母、数字、下划线和中文', trigger: 'blur' },
  ],
  password: [
    { required: true, message: '请输入密码', trigger: 'blur' },
    { min: 8, max: 128, message: '密码长度为 8-128 个字符', trigger: 'blur' },
  ],
  confirmPassword: [
    { required: true, message: '请确认密码', trigger: 'blur' },
    { validator: validateConfirm, trigger: 'blur' },
  ],
  email: [
    { required: true, message: '请输入邮箱', trigger: 'blur' },
    { type: 'email', message: '请输入正确的邮箱格式', trigger: 'blur' },
  ],
  nickname: [
    { max: 64, message: '昵称最多 64 个字符', trigger: 'blur' },
  ],
}

async function handleRegister() {
  if (!formRef.value) return
  const valid = await formRef.value.validate().catch(() => false)
  if (!valid) return

  registering.value = true
  try {
    const res = await registerApi({
      username: form.username,
      password: form.password,
      email: form.email,
      nickname: form.nickname || undefined,
      role: form.role,
    })
    localStorage.setItem('access_token', res.data.access_token)
    localStorage.setItem('refresh_token', res.data.refresh_token)
    auth.user = res.data.user
    auth.token = res.data.access_token
    msg.success(`注册成功，欢迎 ${res.data.user.nickname || res.data.user.username}！`)
    const role = res.data.user.role
    if (role === 'admin') router.push('/admin/dashboard')
    else if (role === 'monitor') router.push('/monitor/dashboard')
    else router.push('/home')
  } catch (err: any) {
    const msg = err?.response?.data?.message
    msg.error(msg || '注册失败，请稍后重试')
  } finally {
    registering.value = false
  }
}
</script>

<template>
  <div class="register-page">
    <div class="register-bg"></div>
    <div class="register-container">
      <div class="register-header">
        <h1 class="register-title">创建账号</h1>
        <p class="register-subtitle">加入大运河生态与文化保护平台</p>
      </div>

      <el-card class="register-card" shadow="always">
        <el-form
          ref="formRef"
          :model="form"
          :rules="rules"
          label-position="top"
          @submit.prevent="handleRegister"
        >
          <el-form-item label="用户名" prop="username">
            <el-input
              v-model="form.username"
              placeholder="3-32 位字母、数字或中文"
              :prefix-icon="User"
              size="large"
              maxlength="32"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="邮箱" prop="email">
            <el-input
              v-model="form.email"
              placeholder="example@mail.com"
              :prefix-icon="Message"
              size="large"
            />
          </el-form-item>

          <el-form-item label="昵称（选填）" prop="nickname">
            <el-input
              v-model="form.nickname"
              placeholder="给自己起个名字吧"
              :prefix-icon="UserFilled"
              size="large"
              maxlength="64"
              show-word-limit
            />
          </el-form-item>

          <el-form-item label="角色" prop="role">
            <el-radio-group v-model="form.role">
              <el-radio value="user">普通用户 — 浏览、互动、问答</el-radio>
              <el-radio value="monitor">监测人员 — 垃圾监测与上报</el-radio>
            </el-radio-group>
          </el-form-item>

          <el-form-item label="密码" prop="password">
            <el-input
              v-model="form.password"
              type="password"
              placeholder="至少 8 位"
              :prefix-icon="Lock"
              show-password
              size="large"
            />
          </el-form-item>

          <el-form-item label="确认密码" prop="confirmPassword">
            <el-input
              v-model="form.confirmPassword"
              type="password"
              placeholder="再次输入密码"
              :prefix-icon="Lock"
              show-password
              size="large"
              @keyup.enter="handleRegister"
            />
          </el-form-item>

          <el-form-item>
            <el-button
              type="primary"
              size="large"
              :loading="registering"
              class="register-btn"
              @click="handleRegister"
            >
              {{ registering ? '注册中...' : '注  册' }}
            </el-button>
          </el-form-item>
        </el-form>

        <div class="register-footer">
          <span>已有账号？</span>
          <router-link to="/login" class="register-link">立即登录</router-link>
        </div>
      </el-card>
    </div>
  </div>
</template>

<style scoped>
.register-page {
  min-height: 100vh;
  display: flex;
  align-items: center;
  justify-content: center;
  position: relative;
  overflow: hidden;
  background: linear-gradient(160deg, #0f0f1e 0%, #1a1a2e 25%, #1d2b3a 50%, #1a2a28 75%, #1a1a2e 100%);
}
.register-page::before {
  content: '';
  position: absolute;
  inset: 0;
  background:
    radial-gradient(ellipse at 80% 20%, rgba(201,184,150,0.06) 0%, transparent 50%),
    radial-gradient(ellipse at 20% 70%, rgba(201,184,150,0.04) 0%, transparent 50%),
    radial-gradient(ellipse at 50% 50%, rgba(44,62,80,0.3) 0%, transparent 70%);
  pointer-events: none;
}
.register-page::after {
  content: '';
  position: absolute;
  bottom: 0;
  left: 0;
  right: 0;
  height: 30vh;
  background:
    radial-gradient(ellipse 160% 100% at 10% 100%, rgba(44,62,80,0.5) 0%, transparent 70%),
    radial-gradient(ellipse 150% 100% at 50% 100%, rgba(26,26,46,0.6) 0%, transparent 65%),
    radial-gradient(ellipse 140% 100% at 90% 100%, rgba(44,62,80,0.4) 0%, transparent 60%);
  pointer-events: none;
}

.register-bg {
  position: absolute;
  inset: 0;
  background: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' viewBox='0 0 1440 200'%3E%3Cpath fill='rgba(201,184,150,0.04)' d='M0,64L48,69.3C96,75,192,85,288,80C384,75,480,53,576,48C672,43,768,53,864,69.3C960,85,1056,107,1152,101.3C1248,96,1344,64,1392,48L1440,32L1440,0L1392,0C1344,0,1248,0,1152,0C1056,0,960,0,864,0C768,0,672,0,576,0C480,0,384,0,288,0C192,0,96,0,48,0L0,0Z'/%3E%3C/svg%3E") no-repeat bottom;
  background-size: cover;
  pointer-events: none;
  opacity: 0.6;
}

.register-container {
  position: relative;
  z-index: 1;
  width: 500px;
  max-width: 92vw;
  animation: fadeInUp 0.7s var(--ease-ink);
}

.register-header {
  text-align: center;
  margin-bottom: 28px;
}

.register-title {
  font-family: 'Ma Shan Zheng', 'ZCOOL XiaoWei', 'STKaiti', 'KaiTi', cursive;
  font-size: 30px;
  color: var(--gold-light);
  margin: 0 0 8px;
  letter-spacing: 4px;
  font-weight: 400;
  text-shadow: 0 2px 16px rgba(0,0,0,0.3);
}

.register-subtitle {
  color: rgba(255,255,255,0.5);
  font-size: 13px;
  margin: 0;
  letter-spacing: 1px;
}

.register-card {
  border-radius: var(--radius-lg);
  border: 1px solid var(--border-hairline);
  box-shadow: var(--shadow-lg);
  overflow: visible;
  position: relative;
}
.register-card::before {
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

.register-card :deep(.el-card__body) {
  padding: 24px 36px 32px;
}

.register-btn {
  width: 100%;
  letter-spacing: 6px;
  font-size: 16px;
  font-weight: 600;
  height: 46px;
  border-radius: var(--radius-md);
  background: linear-gradient(135deg, var(--ink-primary) 0%, #243445 100%);
  border: none;
  transition: all 0.3s var(--ease-ink);
  margin-top: 4px;
}
.register-btn:hover {
  background: linear-gradient(135deg, var(--ink-primary-light) 0%, #2c3e50 100%);
  box-shadow: 0 6px 24px rgba(44,62,80,0.35);
  transform: translateY(-2px);
}

.register-footer {
  text-align: center;
  color: var(--text-muted);
  font-size: 13px;
  margin-top: 4px;
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

@media (max-width: 768px) {
  .register-container { width: 420px; }
  .register-card :deep(.el-card__body) { padding: 20px 24px 28px; }
}

@media (max-width: 480px) {
  .register-container { padding: 0 8px; }
  .register-title { font-size: 24px; letter-spacing: 2px; }
  .register-subtitle { font-size: 11px; }
  .register-card :deep(.el-card__body) { padding: 16px 16px 22px; }
  .register-card::before { left: 16px; right: 16px; }
  .register-btn { font-size: 14px; letter-spacing: 3px; }
  .el-radio { font-size: 13px; }
}
</style>
