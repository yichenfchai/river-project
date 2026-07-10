<script setup lang="ts">
import { computed } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { HomeFilled, MapLocation, ChatLineSquare, UserFilled } from '@element-plus/icons-vue'
import { useAuthStore } from '@/stores/auth'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

const tabs = [
  { path: '/home', label: '首页', icon: HomeFilled },
  { path: '/map', label: '地图', icon: MapLocation },
  { path: '/plaza', label: '广场', icon: ChatLineSquare },
  { path: '/leaderboard', label: '我的', icon: UserFilled },
]

const activeTab = computed(() => {
  const p = route.path
  if (p.startsWith('/home')) return '/home'
  if (p.startsWith('/map')) return '/map'
  if (p.startsWith('/plaza')) return '/plaza'
  if (p.startsWith('/leaderboard') || p.startsWith('/quiz') || p.startsWith('/shop') || p.startsWith('/story')) return '/leaderboard'
  return '/home'
})

function navigate(path: string) {
  if (path === '/leaderboard') {
    // "我的" tab goes to leaderboard which serves as profile hub
  }
  router.push(path)
}
</script>

<template>
  <nav class="mobile-tab-bar">
    <button
      v-for="tab in tabs"
      :key="tab.path"
      class="tab-item"
      :class="{ active: activeTab === tab.path }"
      @click="navigate(tab.path)"
    >
      <component :is="tab.icon" class="tab-icon" />
      <span class="tab-label">{{ tab.label }}</span>
    </button>
  </nav>
</template>

<style scoped>
.mobile-tab-bar {
  display: none;
  position: fixed;
  bottom: 0;
  left: 0;
  right: 0;
  z-index: 200;
  background: #fffdf9;
  border-top: 1px solid rgba(201, 184, 150, 0.2);
  box-shadow: 0 -4px 20px rgba(0, 0, 0, 0.06);
  justify-content: space-around;
  align-items: center;
  height: 56px;
  padding-bottom: env(safe-area-inset-bottom, 0);
}
.mobile-tab-bar::before {
  content: '';
  position: absolute;
  top: 0;
  left: 20%;
  right: 20%;
  height: 1px;
  background: linear-gradient(90deg, transparent, var(--gold), transparent);
  opacity: 0.4;
}

.tab-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  min-width: 48px;
  min-height: 44px;
  border: none;
  background: none;
  cursor: pointer;
  color: var(--text-muted);
  transition: all 0.25s var(--ease-ink);
  -webkit-tap-highlight-color: transparent;
  position: relative;
}
.tab-item::after {
  content: '';
  position: absolute;
  top: 0;
  left: 50%;
  transform: translateX(-50%) scaleX(0);
  width: 20px;
  height: 2px;
  background: var(--gold);
  border-radius: 0 0 2px 2px;
  transition: transform 0.25s var(--ease-ink);
}
.tab-item.active {
  color: var(--ink-primary);
}
.tab-item.active::after {
  transform: translateX(-50%) scaleX(1);
}

.tab-icon {
  font-size: 22px;
  transition: transform 0.25s var(--ease-ink);
}
.tab-item.active .tab-icon {
  transform: translateY(-1px);
}

.tab-label {
  font-size: 11px;
  line-height: 1;
  font-weight: 500;
}

@media (max-width: 768px) {
  .mobile-tab-bar {
    display: flex;
  }
}
</style>
