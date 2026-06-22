<script setup lang="ts">
import { computed, ref, onMounted, onUnmounted } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMenu, ElMenuItem, ElDropdown, ElDropdownMenu, ElDropdownItem, ElIcon } from 'element-plus'
import { HomeFilled, MapLocation, Reading, ChatLineSquare, TrophyBase, Present, ArrowDown } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const allNavItems = [
  { path: '/home', name: '首页', icon: HomeFilled, guest: true },
  { path: '/map', name: '时空地图', icon: MapLocation, guest: true },
  { path: '/story', name: '科普故事', icon: Reading, guest: true },
  { path: '/plaza', name: '分享广场', icon: ChatLineSquare, guest: false },
  { path: '/quiz', name: '趣味问答', icon: TrophyBase, guest: false },
  { path: '/leaderboard', name: '排行榜', icon: TrophyBase, guest: false },
  { path: '/shop', name: '兑换商店', icon: Present, guest: true },
]

const navItems = computed(() =>
  auth.isGuest ? allNavItems.filter((i) => i.guest) : allNavItems,
)

const activePath = computed(() => route.path)

const isHome = computed(() => route.path === '/home')
const isFullWidth = computed(() => route.path === '/home' || route.path === '/map')
const headerScrolled = ref(false)

function onScroll() {
  headerScrolled.value = window.scrollY > 60
}

onMounted(() => {
  window.addEventListener('scroll', onScroll, { passive: true })
})

onUnmounted(() => {
  window.removeEventListener('scroll', onScroll)
})

function navigate(path: string) {
  router.push(path)
}

function handleLogout() {
  auth.logout()
  router.push('/login')
}

function goToWorkbench() {
  if (auth.isAdmin) router.push('/admin/dashboard')
  else if (auth.isMonitor) router.push('/monitor/dashboard')
}
</script>

<template>
  <div class="user-layout">
    <header class="user-header" :class="{ scrolled: headerScrolled, 'home-overlay': isHome }">
      <div class="header-left">
        <span class="logo" @click="router.push('/')">🏛 大运河保护平台</span>
      </div>
      <div class="header-center">
        <el-menu
          :default-active="activePath"
          mode="horizontal"
          :ellipsis="false"
          @select="navigate"
        >
          <el-menu-item v-for="item in navItems" :key="item.path" :index="item.path">
            <el-icon><component :is="item.icon" /></el-icon>
            <span>{{ item.name }}</span>
          </el-menu-item>
        </el-menu>
      </div>
      <div class="header-right">
        <template v-if="auth.isGuest">
          <el-button type="primary" size="small" round @click="router.push('/login')">
            登录 / 注册
          </el-button>
        </template>
        <template v-else>
          <el-dropdown trigger="click">
            <span class="user-trigger">
              <el-avatar :size="32" :src="auth.user?.avatar_url" />
              <span class="username">{{ auth.user?.nickname || auth.user?.username }}</span>
              <el-icon class="arrow"><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="goToWorkbench" v-if="auth.isAdmin || auth.isMonitor">
                  工作台
                </el-dropdown-item>
                <el-dropdown-item @click="navigate('/leaderboard')">个人中心</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </div>
    </header>

    <main class="user-main" :class="{ 'full-width': isFullWidth }">
      <router-view v-slot="{ Component, route }">
        <transition name="page-fade" mode="out-in">
          <component :is="Component" :key="route.fullPath" />
        </transition>
      </router-view>
    </main>

    <footer class="user-footer">
      <span>&copy; 2026 大运河生态与文化保护平台 — Grand Canal Guardian</span>
    </footer>
  </div>
</template>

<style scoped>
.user-layout {
  min-height: 100vh;
  display: flex;
  flex-direction: column;
}

.user-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 0 24px;
  height: 60px;
  background: #2c3e50;
  border-bottom: 1px solid rgba(255,255,255,0.06);
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  position: sticky;
  top: 0;
  z-index: 100;
  transition: background 0.3s, box-shadow 0.3s;
}

.user-header :deep(.el-menu) {
  background: transparent !important;
  border-bottom: none !important;
}

.user-header :deep(.el-menu .el-menu-item) {
  color: rgba(255, 255, 255, 0.8) !important;
  border-bottom-color: transparent !important;
  transition: color 0.2s, border-color 0.2s;
}

.user-header :deep(.el-menu .el-menu-item:hover) {
  color: #c9b896 !important;
}

.user-header :deep(.el-menu .el-menu-item.is-active) {
  color: #c9b896 !important;
  border-bottom-color: #c9b896 !important;
}

.header-left .logo {
  font-size: 16px;
  font-weight: 600;
  color: #e0d6c0;
  cursor: pointer;
  white-space: nowrap;
}

.header-left .logo:hover {
  color: #c9b896;
}

.header-center {
  flex: 1;
  display: flex;
  justify-content: center;
}

.header-right {
  display: flex;
  align-items: center;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  padding: 4px 8px;
  border-radius: 6px;
  transition: background 0.2s;
  color: rgba(255, 255, 255, 0.85);
}

.user-trigger:hover {
  background: rgba(255, 255, 255, 0.08);
}

.username {
  font-size: 14px;
  color: rgba(255, 255, 255, 0.85);
}

.arrow {
  font-size: 12px;
  color: rgba(255, 255, 255, 0.55);
}

/* Home overlay — transparent on top of banner */
.user-header.home-overlay {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  background: transparent;
  border-bottom: 1px solid transparent;
  box-shadow: none;
}

.user-header.home-overlay .logo {
  color: #fff;
  text-shadow: 0 1px 6px rgba(0, 0, 0, 0.4);
}

.user-header.home-overlay.scrolled {
  background: rgba(44, 62, 80, 0.92);
  backdrop-filter: blur(12px);
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
  box-shadow: 0 2px 16px rgba(0, 0, 0, 0.2);
}

.user-header.home-overlay.scrolled .logo {
  color: #e0d6c0;
  text-shadow: none;
}

.user-header.home-overlay :deep(.el-menu) {
  background: transparent !important;
  border-bottom: none !important;
}

.user-header.home-overlay :deep(.el-menu .el-menu-item) {
  color: rgba(255, 255, 255, 0.9) !important;
  border-bottom-color: transparent !important;
}

.user-header.home-overlay.scrolled :deep(.el-menu .el-menu-item) {
  color: rgba(255, 255, 255, 0.8) !important;
}

.user-header.home-overlay.scrolled :deep(.el-menu .el-menu-item.is-active) {
  color: #c9b896 !important;
  border-bottom-color: #c9b896 !important;
}

.user-header.home-overlay .user-trigger {
  color: rgba(255, 255, 255, 0.9);
}

.user-header.home-overlay.scrolled .user-trigger {
  color: rgba(255, 255, 255, 0.85);
}

.user-header.home-overlay .user-trigger:hover {
  background: rgba(255, 255, 255, 0.12);
}

.user-header.home-overlay.scrolled .user-trigger:hover {
  background: rgba(255, 255, 255, 0.08);
}

.user-header.home-overlay .username,
.user-header.home-overlay .arrow {
  color: inherit;
}

.user-main {
  flex: 1;
  padding: 24px;
  max-width: 1400px;
  width: 100%;
  margin: 0 auto;
  box-sizing: border-box;
}

.user-main.full-width {
  padding: 0;
  max-width: 100%;
  display: flex;
  flex-direction: column;
}

.user-footer {
  text-align: center;
  padding: 16px;
  color: #889099;
  font-size: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.06);
  background: #2c3e50;
}
</style>
