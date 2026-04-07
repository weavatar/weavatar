import { createRouter, createWebHistory } from 'vue-router'
import { setupGuards } from './guards'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    {
      path: '/',
      component: () => import('@/views/DefaultLayout.vue'),
      children: [
        {
          path: '',
          name: 'home',
          component: () => import('@/views/home/HomeView.vue'),
          meta: { title: '首页' }
        },
        {
          path: 'login',
          name: 'login',
          component: () => import('@/views/auth/LoginView.vue'),
          meta: { title: '登录', guest: true }
        },
        {
          path: 'logout',
          name: 'logout',
          component: () => import('@/views/auth/LogoutView.vue'),
          meta: { title: '退出' }
        },
        {
          path: 'oauth/callback',
          name: 'oauth-callback',
          component: () => import('@/views/auth/CallbackView.vue'),
          meta: { title: '登录中' }
        },
        {
          path: 'doc',
          name: 'doc',
          component: () => import('@/views/pages/DocView.vue'),
          meta: { title: '文档' }
        },
        {
          path: 'help',
          name: 'help',
          component: () => import('@/views/pages/HelpView.vue'),
          meta: { title: '帮助' }
        },
        {
          path: 'about',
          name: 'about',
          component: () => import('@/views/pages/AboutView.vue'),
          meta: { title: '关于' }
        },
        {
          path: 'privacy',
          name: 'privacy',
          component: () => import('@/views/pages/PrivacyView.vue'),
          meta: { title: '隐私' }
        },
        {
          path: 'user/avatar',
          name: 'user-avatar',
          component: () => import('@/views/user/AvatarListView.vue'),
          meta: { title: '头像', requiresAuth: true }
        },
        {
          path: 'user/info',
          name: 'user-info',
          component: () => import('@/views/user/InfoView.vue'),
          meta: { title: '账号信息', requiresAuth: true }
        },
        {
          path: '404',
          name: '404',
          component: () => import('@/views/pages/NotFoundView.vue'),
          meta: { title: '404' }
        },
        { path: ':catchAll(.*)', redirect: { name: '404' } }
      ]
    }
  ]
})

setupGuards(router)

export default router
