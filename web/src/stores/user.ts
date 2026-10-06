import { defineStore } from 'pinia'
import { fetchUserInfo } from '@/api/user'
import { ApiError } from '@/utils/http'
import { DEFAULT_AVATAR } from '@/constants/links'

export interface UserInfo {
  id: string
  avatar: string
  nickname: string
  real_name: boolean
  created_at: string
}

export interface UserState {
  info: UserInfo
  auth: { token: string; login: boolean }
}

export const defaultInfo = (): UserInfo => ({
  id: '',
  avatar: DEFAULT_AVATAR,
  nickname: '未登录',
  real_name: false,
  created_at: ''
})

export const useUserStore = defineStore('user', {
  state: (): UserState => ({
    info: defaultInfo(),
    auth: { token: '', login: false }
  }),
  getters: {
    isLogin: (state) => state.auth.login && !!state.auth.token
  },
  actions: {
    async freshUserInfo() {
      try {
        const data = await fetchUserInfo({ noAlert: true })
        this.info = { ...this.info, ...data }
      } catch (e) {
        // 401 已由 http 层清理，404 表示账号已不存在
        if (e instanceof ApiError && e.status === 404) this.clearToken()
      }
    },
    updateToken(token: string) {
      this.auth.token = token
      this.auth.login = true
    },
    clearToken() {
      this.auth.token = ''
      this.auth.login = false
      this.info = defaultInfo()
    }
  },
  persist: true
})
