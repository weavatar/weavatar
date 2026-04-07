import { defineStore } from 'pinia'
import user from '@/api/user'
import { useRequest } from 'alova/client'

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
    freshUserInfo() {
      useRequest(user.info()).onSuccess(({ data }: any) => {
        this.info = { ...this.info, ...data }
      })
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
