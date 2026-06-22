<script setup lang="ts">
import { computed, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { ElMenu, ElMenuItem, ElDropdown, ElDropdownMenu, ElDropdownItem, ElIcon, ElSubMenu } from 'element-plus'
import { DataBoard, Camera, List, User, ArrowDown, SwitchButton, HomeFilled } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()
const isCollapse = ref(false)

const menuItems = [
  { path: '/monitor/dashboard', name: '数据看板', icon: DataBoard },
  { path: '/monitor/report', name: '垃圾上报', icon: Camera },
  { path: '/monitor/history', name: '上报记录', icon: List },
  { path: '/monitor/profile', name: '个人中心', icon: User },
]

const activePath = computed(() => route.path)

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <div class="monitor-layout">
    <aside class="monitor-sidebar" :class="{ collapsed: isCollapse }">
      <div class="sidebar-header">
        <span class="logo" v-if="!isCollapse">🔬 监测员工作台</span>
        <span class="logo-small" v-else>🔬</span>
      </div>

      <el-menu
        :default-active="activePath"
        :collapse="isCollapse"
        router
        background-color="#1d2b3a"
        text-color="#a8b7c7"
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
        <el-icon v-else class="back-icon" @click="router.push('/home')"><HomeFilled /></el-icon>
      </div>
    </aside>

    <div class="monitor-right">
      <header class="monitor-topbar">
        <div class="topbar-left">
          <el-icon class="collapse-btn" @click="isCollapse = !isCollapse" :size="20">
            <SwitchButton />
          </el-icon>
          <span class="page-title">监测员工作台</span>
        </div>

        <el-dropdown trigger="click">
          <span class="user-trigger">
            <el-avatar :size="28" :src="auth.user?.avatar_url" />
            <span>{{ auth.user?.nickname || auth.user?.username }}</span>
            <el-icon><ArrowDown /></el-icon>
          </span>
          <template #dropdown>
            <el-dropdown-menu>
              <el-dropdown-item @click="router.push('/monitor/profile')">个人中心</el-dropdown-item>
              <el-dropdown-item @click="router.push('/admin/dashboard')" v-if="auth.isAdmin">管理后台</el-dropdown-item>
              <el-dropdown-item divided @click="handleLogout">退出登录</el-dropdown-item>
            </el-dropdown-menu>
          </template>
        </el-dropdown>
      </header>

      <main class="monitor-content">
        <router-view v-slot="{ Component, route }">
          <transition name="page-fade" mode="out-in">
            <component :is="Component" :key="route.fullPath" />
          </transition>
        </router-view>
      </main>
    </div>
  </div>
</template>

<style scoped>
.monitor-layout {
  display: flex;
  min-height: 100vh;
  background: #ebe6dc;
}

.monitor-sidebar {
  width: 220px;
  background: #1d2b3a;
  display: flex;
  flex-direction: column;
  transition: width 0.3s;
  flex-shrink: 0;
}

.monitor-sidebar.collapsed {
  width: 64px;
}

.sidebar-header {
  height: 56px;
  display: flex;
  align-items: center;
  justify-content: center;
  border-bottom: 1px solid rgba(255, 255, 255, 0.08);
}

.sidebar-header .logo {
  color: #fff;
  font-size: 15px;
  font-weight: 600;
  white-space: nowrap;
}

.sidebar-header .logo-small {
  color: #fff;
  font-size: 20px;
}

.monitor-sidebar :deep(.el-menu) {
  border-right: none;
  flex: 1;
}

.sidebar-footer {
  border-top: 1px solid rgba(255, 255, 255, 0.08);
  padding: 0;
}

.sidebar-footer :deep(.el-menu-item) {
  background-color: #1d2b3a !important;
  color: #a8b7c7 !important;
}

.back-icon {
  color: #a8b7c7;
  cursor: pointer;
  padding: 16px;
  display: block;
  text-align: center;
  font-size: 18px;
}

.back-icon:hover {
  color: #c9b896;
}

.monitor-right {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
}

.monitor-topbar {
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

.monitor-content {
  flex: 1;
  padding: 20px;
  overflow-y: auto;
}
</style>
