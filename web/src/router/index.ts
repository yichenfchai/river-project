import { createRouter, createWebHistory } from 'vue-router'
import { useAuthStore } from '@/stores/auth'

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/',
      component: () => import('@/components/Layout.vue'),
      children: [
        { path: '', name: 'home', component: () => import('@/views/Home.vue') },
        { path: 'posts', name: 'posts', component: () => import('@/views/Posts.vue') },
        { path: 'posts/create', name: 'post-create', component: () => import('@/views/PostCreate.vue'), meta: { auth: true } },
        { path: 'posts/:id', name: 'post-detail', component: () => import('@/views/PostDetail.vue') },
        { path: 'quiz', name: 'quiz', component: () => import('@/views/Quiz.vue') },
        { path: 'leaderboard', name: 'leaderboard', component: () => import('@/views/Leaderboard.vue') },
        { path: 'profile', name: 'profile', component: () => import('@/views/Profile.vue'), meta: { auth: true } },
        { path: 'admin', name: 'admin', component: () => import('@/views/Admin.vue'), meta: { auth: true, role: 'admin' } },
      ],
    },
    { path: '/login', name: 'login', component: () => import('@/views/Login.vue'), meta: { guest: true } },
    { path: '/register', name: 'register', component: () => import('@/views/Register.vue'), meta: { guest: true } },
  ],
})

router.beforeEach((to, _from, next) => {
  const auth = useAuthStore()

  if (to.meta.auth && !auth.isLoggedIn) {
    return next('/login')
  }
  if (to.meta.guest && auth.isLoggedIn) {
    return next('/')
  }
  if (to.meta.role && !auth.isAdmin) {
    return next('/')
  }
  next()
})

export default router
