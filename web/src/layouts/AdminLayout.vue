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
  background: #1a1a2e;
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
  flex-shrink: 0;
}

.admin-sidebar.collapsed {
  width: 64px;
}

.sidebar-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.06);
}

.sidebar-header .logo {
  color: #e0e0ff;
  font-size: 15px;
  font-weight: 600;
}

.sidebar-header .logo-small {
  color: #e0e0ff;
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
  background: #fff;
  border-bottom: 1px solid #e4e7ed;
  box-shadow: 0 1px 2px rgba(0, 0, 0, 0.04);
}

.topbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
}

.collapse-btn {
  cursor: pointer;
  color: #606266;
}

.collapse-btn:hover {
  color: #c9b896;
}

.hamburger-btn {
  display: none;
  cursor: pointer;
  color: #606266;
}

.hamburger-btn:hover {
  color: #c9b896;
}

.mobile-nav {
  padding: 12px 0;
}

.mobile-nav-title {
  padding: 12px 20px;
  font-size: 15px;
  font-weight: 600;
  color: #1a1a2e;
  border-bottom: 1px solid #ebeef5;
  margin-bottom: 8px;
}

.mobile-nav-item {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 14px 20px;
  font-size: 15px;
  color: #303133;
  cursor: pointer;
  transition: background 0.2s;
}

.mobile-nav-item:hover {
  background: #f5f3ef;
  color: #2c3e50;
}

.mobile-nav-divider {
  height: 1px;
  background: #ebeef5;
  margin: 8px 20px;
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

.page-title {
  font-size: 16px;
  font-weight: 500;
  color: #303133;
}

.user-trigger {
  display: flex;
  align-items: center;
  gap: 8px;
  cursor: pointer;
  font-size: 13px;
  color: #606266;
}

.admin-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}
</style>
