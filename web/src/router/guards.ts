import type { Router } from 'vue-router'
import { useUserStore } from '@/stores'

export function setupGuards(router: Router) {
  router.beforeEach((to) => {
    window.$loadingBar?.start()

    const userStore = useUserStore()

    if (to.meta.requiresAuth && !userStore.auth.login) {
      return { name: 'login' }
    }
    if (to.meta.guest && userStore.auth.login) {
      return { name: 'user-avatar' }
    }

    if (to.meta.title) {
      document.title = `${to.meta.title} - WeAvatar`
    }
  })

  router.afterEach(() => {
    window.$loadingBar?.finish()
  })
}
