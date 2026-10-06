import { get, post, put, type RequestConfig } from '@/utils/http'
import type { UserInfo } from '@/stores'

export const fetchUserInfo = (config?: RequestConfig) =>
  get<UserInfo>('/user/info', undefined, config)

export const updateUserInfo = (data: { nickname: string; avatar: string }) =>
  put('/user/info', data)

/** 注销前需重新经树新峰通行证授权，返回授权地址 */
export const fetchDeletionUrl = () => get<{ url: string }>('/user/deletion/login')

export const confirmDeletion = (code: string, state: string) =>
  post('/user/deletion/confirm', { code, state })
