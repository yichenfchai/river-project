<template>
  <el-container class="layout">
    <el-header class="header">
      <div class="logo" @click="$router.push('/')">🏯 大运河守护者</div>
      <el-menu mode="horizontal" :default-active="activeMenu" router class="nav">
        <el-menu-item index="/">首页</el-menu-item>
        <el-menu-item index="/posts">分享广场</el-menu-item>
        <el-menu-item index="/quiz">趣味问答</el-menu-item>
        <el-menu-item index="/leaderboard">排行榜</el-menu-item>
        <el-menu-item v-if="auth.isAdmin" index="/admin">管理后台</el-menu-item>
      </el-menu>
      <div class="user-area">
        <template v-if="auth.isLoggedIn">
          <el-dropdown @command="handleCommand">
            <span class="user-dropdown">
              {{ auth.user?.nickname || auth.user?.username }}
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="profile">个人中心</el-dropdown-item>
                <el-dropdown-item command="logout" divided>退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
        <template v-else>
          <el-button type="primary" size="small" @click="$router.push('/login')">登录</el-button>
          <el-button size="small" @click="$router.push('/register')">注册</el-button>
        </template>
      </div>
    </el-header>
    <el-main class="main">
      <router-view />
    </el-main>
  </el-container>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()

const activeMenu = computed(() => {
  if (route.path.startsWith('/posts')) return '/posts'
  if (route.path.startsWith('/quiz')) return '/quiz'
  if (route.path.startsWith('/leaderboard')) return '/leaderboard'
  if (route.path.startsWith('/admin')) return '/admin'
  return '/'
})

function handleCommand(cmd: string) {
  if (cmd === 'profile') router.push('/profile')
  if (cmd === 'logout') {
    auth.logout()
    router.push('/')
  }
}
</script>

<style scoped>
.layout { min-height: 100vh; background: #f5f7fa; }
.header {
  display: flex; align-items: center; background: #fff;
  border-bottom: 1px solid #e4e7ed; padding: 0 20px; height: 60px;
}
.logo { font-size: 18px; font-weight: bold; color: #409eff; cursor: pointer; margin-right: 30px; white-space: nowrap; }
.nav { flex: 1; border-bottom: none !important; }
.user-area { display: flex; align-items: center; gap: 8px; }
.user-dropdown { cursor: pointer; display: flex; align-items: center; gap: 4px; }
.main { padding: 24px; max-width: 1200px; margin: 0 auto; width: 100%; }
</style>
