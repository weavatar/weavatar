import type { Router } from 'vue-router'
import { useUserStore } from '@/stores'
import { SITE_NAME } from '@/constants/links'
import { safeRedirect } from '@/utils/redirect'

export function setupGuards(router: Router) {
  router.beforeEach((to) => {
    window.$loadingBar?.start()

    const userStore = useUserStore()
    const requiresAuth = to.matched.some((r) => r.meta.requiresAuth)

    if (requiresAuth && !userStore.isLogin) {
      return { name: 'login', query: { redirect: to.fullPath } }
    }
    if (to.meta.guest && userStore.isLogin) {
      return safeRedirect(to.query.redirect) || { name: 'user-avatar' }
    }

    document.title = to.meta.title ? `${to.meta.title} - ${SITE_NAME}` : SITE_NAME
  })

  router.afterEach(() => {
    window.$loadingBar?.finish()
  })

  // 路由组件加载失败等异常不会触发 afterEach
  router.onError(() => {
    window.$loadingBar?.error()
  })
}
