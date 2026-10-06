import { defineStore } from 'pinia'
import user from '@/api/user'
import { HttpError } from '@/utils/http'

export const useUserStore = defineStore('user', {
  state: () => ({
    info: {
      id: '',
      avatar: 'https://weavatar.com/avatar/?d=mp',
      nickname: '未登录',
      real_name: false,
      created_at: ''
    },
    auth: {
      token: '',
      login: false
    }
  }),
  actions: {
    async freshUserInfo() {
      // meta 要设在 Method 上才会传到响应拦截器，useRequest 的第二个参数不会转交
      const method = user.info()
      method.meta = { noAlert: true }
      try {
        const data = await method
        this.info = { ...this.info, ...data }
      } catch (error) {
        // 401 已在 http 层清理；404 表示账号已不存在，同样清理登录态
        if (error instanceof HttpError && error.status === 404) this.clearToken()
      }
    },
    resetUserInfo() {
      this.info = {
        id: '',
        avatar: 'https://weavatar.com/avatar/?d=mp',
        nickname: '未登录',
        real_name: false,
        created_at: ''
      }
    },
    updateToken(token: string) {
      this.auth.token = token
      this.auth.login = true
    },
    clearToken() {
      this.auth.token = ''
      this.auth.login = false
      this.resetUserInfo()
    }
  },
  persist: true
})
