<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h2>注册</h2>
      <el-form :model="form" :rules="rules" ref="formRef" label-width="0" @submit.prevent="handleRegister">
        <el-form-item prop="username">
          <el-input v-model="form.username" placeholder="用户名 (3-32位)" />
        </el-form-item>
        <el-form-item prop="email">
          <el-input v-model="form.email" placeholder="邮箱" />
        </el-form-item>
        <el-form-item prop="nickname">
          <el-input v-model="form.nickname" placeholder="昵称 (选填)" />
        </el-form-item>
        <el-form-item prop="password">
          <el-input v-model="form.password" type="password" placeholder="密码 (至少8位)" show-password />
        </el-form-item>
        <el-form-item prop="confirm">
          <el-input v-model="form.confirm" type="password" placeholder="确认密码" show-password />
        </el-form-item>
        <el-form-item>
          <el-button type="primary" native-type="submit" :loading="auth.loading" style="width:100%">
            注册
          </el-button>
        </el-form-item>
      </el-form>
      <p class="switch">已有账号？<router-link to="/login">去登录</router-link></p>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'
import { ElMessage } from 'element-plus'

const router = useRouter()
const auth = useAuthStore()
const formRef = ref()

const form = reactive({ username: '', email: '', nickname: '', password: '', confirm: '' })
const validateConfirm = (_rule: unknown, value: string, cb: (err?: Error) => void) => {
  if (value !== form.password) cb(new Error('两次密码不一致'))
  else cb()
}
const rules = {
  username: [{ required: true, min: 3, max: 32, message: '用户名 3-32 位', trigger: 'blur' }],
  email: [{ required: true, type: 'email', message: '请输入有效邮箱', trigger: 'blur' }],
  password: [{ required: true, min: 8, message: '密码至少 8 位', trigger: 'blur' }],
  confirm: [{ required: true, validator: validateConfirm, trigger: 'blur' }],
}

async function handleRegister() {
  await formRef.value?.validate()
  const ok = await auth.register({
    username: form.username,
    password: form.password,
    email: form.email,
    nickname: form.nickname,
  })
  if (ok) {
    ElMessage.success('注册成功，请登录')
    router.push('/login')
  } else {
    ElMessage.error('注册失败，用户名或邮箱可能已被使用')
  }
}
</script>

<style scoped>
.auth-page { display: flex; justify-content: center; align-items: center; min-height: 100vh; background: #f5f7fa; }
.auth-card { width: 420px; }
.auth-card h2 { text-align: center; margin-bottom: 24px; }
.switch { text-align: center; color: #999; font-size: 14px; }
.switch a { color: #409eff; }
</style>
