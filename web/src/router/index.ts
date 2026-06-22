import { createRouter, createWebHistory } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '@/stores/auth'
import type { UserRole } from '@/types'

const UserLayout = () => import('@/layouts/UserLayout.vue')
const MonitorLayout = () => import('@/layouts/MonitorLayout.vue')
const AdminLayout = () => import('@/layouts/AdminLayout.vue')

const LoginPage = () => import('@/pages/public/LoginPage.vue')
const NotFoundPage = () => import('@/pages/public/NotFound.vue')

function requireAuth(roles?: UserRole[]) {
  return function (_to: unknown, _from: unknown, next: (path?: string) => void) {
    const auth = useAuthStore()
    if (!auth.isLoggedIn) {
      next('/login')
      return
    }
    if (roles && roles.length > 0 && auth.user) {
      if (!roles.includes(auth.user.role)) {
        next('/')
        return
      }
    }
    next()
  }
}

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: LoginPage,
      meta: { guest: true },
    },
    {
      path: '/',
      component: UserLayout,
      redirect: '/home',
      children: [
        {
          path: 'home',
          name: 'Home',
          component: () => import('@/pages/user/Home.vue'),
        },
        {
          path: 'map',
          name: 'MapView',
          component: () => import('@/pages/user/MapView.vue'),
        },
        {
          path: 'story',
          name: 'StoryPage',
          component: () => import('@/pages/user/StoryPage.vue'),
        },
        {
          path: 'plaza',
          name: 'PlazaPage',
          component: () => import('@/pages/user/PlazaPage.vue'),
          meta: { roles: ['user', 'monitor', 'admin'] as UserRole[] },
        },
        {
          path: 'quiz',
          name: 'QuizPage',
          component: () => import('@/pages/user/QuizPage.vue'),
          meta: { roles: ['user', 'monitor', 'admin'] as UserRole[] },
        },
        {
          path: 'leaderboard',
          name: 'LeaderboardPage',
          component: () => import('@/pages/user/LeaderboardPage.vue'),
          meta: { roles: ['user', 'monitor', 'admin'] as UserRole[] },
        },
        {
          path: 'shop',
          name: 'ShopPage',
          component: () => import('@/pages/user/ShopPage.vue'),
        },
      ],
    },
    {
      path: '/monitor',
      component: MonitorLayout,
      redirect: '/monitor/dashboard',
      meta: { roles: ['monitor', 'admin'] as UserRole[] },
      beforeEnter: requireAuth(['monitor', 'admin']),
      children: [
        {
          path: 'dashboard',
          name: 'MonitorDashboard',
          component: () => import('@/pages/monitor/Dashboard.vue'),
        },
        {
          path: 'report',
          name: 'GarbageReport',
          component: () => import('@/pages/monitor/GarbageReport.vue'),
        },
        {
          path: 'history',
          name: 'ReportHistory',
          component: () => import('@/pages/monitor/ReportHistory.vue'),
        },
        {
          path: 'profile',
          name: 'MonitorProfile',
          component: () => import('@/pages/monitor/Profile.vue'),
        },
      ],
    },
    {
      path: '/admin',
      component: AdminLayout,
      redirect: '/admin/dashboard',
      meta: { roles: ['admin'] as UserRole[] },
      beforeEnter: requireAuth(['admin']),
      children: [
        {
          path: 'dashboard',
          name: 'AdminDashboard',
          component: () => import('@/pages/admin/Dashboard.vue'),
        },
        {
          path: 'users',
          name: 'UserManagement',
          component: () => import('@/pages/admin/UserManagement.vue'),
        },
        {
          path: 'posts',
          name: 'PostReview',
          component: () => import('@/pages/admin/PostReview.vue'),
        },
        {
          path: 'garbage',
          name: 'GarbageReports',
          component: () => import('@/pages/admin/GarbageReports.vue'),
        },
        {
          path: 'questions',
          name: 'QuestionManagement',
          component: () => import('@/pages/admin/QuestionManagement.vue'),
        },
        {
          path: 'shop',
          name: 'ShopManagement',
          component: () => import('@/pages/admin/ShopManagement.vue'),
        },
      ],
    },
    {
      path: '/:pathMatch(.*)*',
      name: 'NotFound',
      component: NotFoundPage,
    },
  ],
})

router.beforeEach((to, from, next) => {
  const auth = useAuthStore()

  if (to.meta.guest) {
    if (auth.isLoggedIn && !auth.isGuest) {
      const role = auth.user?.role
      if (role === 'admin') next('/admin/dashboard')
      else if (role === 'monitor') next('/monitor/dashboard')
      else next('/home')
      return
    }
    next()
    return
  }

  const roles = to.meta.roles as UserRole[] | undefined
  if (roles && roles.length > 0) {
    if (!auth.isLoggedIn) {
      next('/login')
      return
    }
    if (auth.user && !roles.includes(auth.user.role)) {
      if (auth.isGuest) {
        ElMessage.warning('此功能需要登录后使用，请先登录')
      }
      next('/home')
      return
    }
  }

  next()
})

export default router
