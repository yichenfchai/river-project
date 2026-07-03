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
  <div class="login-page">
    <div class="login-bg"></div>
    <div class="login-container">
      <div class="login-header">
        <h1 class="login-title">创建账号</h1>
        <p class="login-subtitle">加入大运河生态与文化保护平台</p>
      </div>

      <el-card class="login-card" shadow="always">
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
              class="login-btn"
              @click="handleRegister"
            >
              {{ registering ? '注册中...' : '注  册' }}
            </el-button>
          </el-form-item>
        </el-form>

        <div class="login-footer">
          <span>已有账号？</span>
          <router-link to="/login" class="register-link">立即登录</router-link>
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
  width: 480px;
  max-width: 92vw;
}

.login-header {
  text-align: center;
  margin-bottom: 24px;
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
  padding: 28px 36px 32px;
}

.login-btn {
  width: 100%;
  letter-spacing: 4px;
  font-size: 16px;
  margin-top: 8px;
}

.login-footer {
  text-align: center;
  color: #909399;
  font-size: 13px;
}

.register-link {
  color: #409eff;
  text-decoration: none;
  font-weight: 500;
}

.register-link:hover {
  text-decoration: underline;
}

@media (max-width: 480px) {
  .login-container {
    padding: 0 8px;
  }

  .login-title {
    font-size: 20px;
    letter-spacing: 1px;
  }

  .login-subtitle {
    font-size: 12px;
  }

  .login-card :deep(.el-card__body) {
    padding: 16px 18px 24px;
  }

  .login-btn {
    font-size: 14px;
  }

  .el-radio {
    font-size: 13px;
  }
}
</style>
