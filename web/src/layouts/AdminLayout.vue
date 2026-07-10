<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMenu, ElMenuItem, ElDropdown, ElDropdownMenu, ElDropdownItem, ElIcon, ElDrawer } from 'element-plus'
import {
  DataBoard, UserFilled, DocumentChecked, DeleteFilled, QuestionFilled, Present, ArrowDown, SwitchButton, HomeFilled, Expand,
} from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const isCollapse = ref(false)
const mobileMenuOpen = ref(false)

const menuItems = [
  { path: '/admin/dashboard', name: '数据看板', icon: DataBoard },
  { path: '/admin/users', name: '用户管理', icon: UserFilled },
  { path: '/admin/posts', name: '帖子审核', icon: DocumentChecked },
  { path: '/admin/garbage', name: '垃圾上报汇总', icon: DeleteFilled },
  { path: '/admin/questions', name: '题目管理', icon: QuestionFilled },
  { path: '/admin/shop', name: '兑换商店', icon: Present },
]

const activePath = computed(() => route.path)

function handleLogout() {
  auth.logout()
  router.push('/login')
}

function onMobileNav(path: string) {
  mobileMenuOpen.value = false
  router.push(path)
}
</script>

<template>
  <div class="admin-layout">
    <aside class="admin-sidebar" :class="{ collapsed: isCollapse }">
      <div class="sidebar-header">
        <span class="logo" v-if="!isCollapse">⚙ 管理后台</span>
        <span class="logo-small" v-else>⚙</span>
      </div>

      <el-menu
        :default-active="activePath"
        :collapse="isCollapse"
        router
        background-color="#1a1a2e"
        text-color="#a0a0b8"
        active-text-color="#c9b896"
      >
        <el-menu-item v-for="item in menuItems" :key="item.path" :index="item.path">
          <el-icon><component :is="item.icon" /></el-icon>
          <template #title>{{ item.name }}</template>
        </el-menu-item>
      </el-menu>

      <div class="sidebar-footer">
        <el-menu-item index="/home" @click="router.push('/home')" v-if="!isCollapse">
          <el-icon><HomeFilled /></el-icon>
          <template #title>返回前台</template>
        </el-menu-item>
        <el-icon v-else class="home-btn" @click="router.push('/home')"><HomeFilled /></el-icon>
      </div>
    </aside>

    <div class="admin-right">
      <header class="admin-topbar">
        <div class="topbar-left">
          <el-icon class="collapse-btn desktop-only" @click="isCollapse = !isCollapse" :size="20">
            <SwitchButton />
          </el-icon>
          <el-icon class="hamburger-btn" :size="22" @click="mobileMenuOpen = true"><Expand /></el-icon>
          <span class="page-title">系统管理后台</span>
        </div>

        <div class="topbar-right">
          <el-dropdown trigger="click">
            <span class="user-trigger">
              <el-avatar :size="28" :src="auth.user?.avatar_url" />
              <span>{{ auth.user?.nickname || auth.user?.username }}</span>
              <el-icon><ArrowDown /></el-icon>
            </span>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item @click="router.push('/monitor/dashboard')" v-if="auth.isMonitor">监测工作台</el-dropdown-item>
                <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </div>
      </header>

      <main class="admin-content">
        <router-view />
      </main>
    </div>

    <el-drawer v-model="mobileMenuOpen" direction="ltr" size="70%" :with-header="false">
      <div class="mobile-nav">
        <div class="mobile-nav-title">⚙ 管理后台</div>
        <div class="mobile-nav-item" v-for="item in menuItems" :key="item.path" @click="onMobileNav(item.path)">
          <el-icon><component :is="item.icon" /></el-icon>
          <span>{{ item.name }}</span>
        </div>
        <div class="mobile-nav-divider"></div>
        <div class="mobile-nav-item" @click="onMobileNav('/home')">
          <el-icon><HomeFilled /></el-icon>
          <span>返回前台</span>
        </div>
      </div>
    </el-drawer>
  </div>
</template>

<style scoped>
.admin-layout {
  display: flex;
  min-height: 100vh;
  background: #ebe6dc;
}

.admin-sidebar {
  width: 220px;
  background: linear-gradient(180deg, #1a1a2e 0%, #151528 100%);
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
  flex-shrink: 0;
  border-right: 1px solid rgba(201, 184, 150, 0.08);
}

.admin-sidebar.collapsed {
  width: 64px;
}

.sidebar-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(201, 184, 150, 0.1);
}

.sidebar-header .logo {
  font-family: 'Noto Serif SC', 'STSong', serif;
  color: var(--gold-light);
  font-size: 15px;
  font-weight: 600;
  letter-spacing: 1px;
}

.sidebar-header .logo-small {
  color: var(--gold-light);
  font-size: 20px;
}

.admin-sidebar :deep(.el-menu) {
  border-right: none;
  flex: 1;
}

.sidebar-footer {
  border-top: 1px solid rgba(255, 255, 255, 0.06);
}

.sidebar-footer :deep(.el-menu-item) {
  background-color: #1a1a2e !important;
  color: #a0a0b8 !important;
}

.home-btn {
  color: #a0a0b8;
  cursor: pointer;
  padding: 16px;
  display: block;
  text-align: center;
  font-size: 18px;
}

.home-btn:hover {
  color: #c9b896;
}

.admin-right {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.admin-topbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 56px;
  padding: 0 20px;
  background: var(--surface-card);
  border-bottom: 1px solid var(--border-hairline);
  box-shadow: var(--shadow-xs);
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.collapse-btn {
  cursor: pointer;
  color: var(--text-secondary);
  transition: color 0.2s;
}
.collapse-btn:hover {
  color: var(--gold-dark);
}

.hamburger-btn {
  display: none;
  cursor: pointer;
  color: var(--text-secondary);
  transition: color 0.2s;
}
.hamburger-btn:hover {
  color: var(--gold-dark);
}

.mobile-nav {
  padding: 8px 0;
}

.mobile-nav-title {
  font-family: 'Noto Serif SC', 'STSong', serif;
  padding: 14px 24px;
  font-size: 16px;
  font-weight: 600;
  color: var(--ink-deep);
  border-bottom: 1px solid var(--border-light);
  margin-bottom: 4px;
  letter-spacing: 1px;
}

.mobile-nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 13px 24px;
  font-size: 15px;
  color: var(--text-primary);
  cursor: pointer;
  transition: all 0.2s;
  border-left: 3px solid transparent;
  margin: 2px 8px;
  border-radius: 0 var(--radius-md) var(--radius-md) 0;
}

.mobile-nav-item:hover {
  background: var(--paper-light);
  color: var(--ink-primary);
  border-left-color: var(--gold);
}

.mobile-nav-divider {
  height: 1px;
  background: var(--border-light);
  margin: 8px 24px;
}

@media (max-width: 768px) {
  .admin-sidebar {
    display: none;
  }

  .hamburger-btn {
    display: inline-flex;
  }

  .collapse-btn.desktop-only {
    display: none;
  }

  .admin-topbar {
    padding: 0 12px;
    height: 50px;
  }

  .admin-content {
    padding: 12px 8px;
  }

  .user-trigger span {
    display: none;
  }
}

/* safe area for notch phones */
@supports (padding-bottom: env(safe-area-inset-bottom)) {
  .admin-content {
    padding-bottom: calc(12px + env(safe-area-inset-bottom));
  }
}

.page-title {
  font-family: 'Noto Serif SC', 'STSong', serif;
  font-size: 16px;
  font-weight: 600;
  color: var(--text-primary);
  letter-spacing: 0.5px;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: var(--text-secondary);
  padding: 4px 8px;
  border-radius: var(--radius-sm);
  transition: background 0.2s;
}
.user-trigger:hover {
  background: var(--paper-light);
}

.admin-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
  background: var(--paper);
}
</style>
