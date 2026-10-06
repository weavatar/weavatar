import { get, post } from '@/utils/http'

export interface Captcha {
  lot_number: string
  captcha_output: string
  pass_token: string
  gen_time: string
}

/** 返回树新峰通行证的授权地址 */
export const fetchLoginUrl = () => get<{ url: string }>('/user/login')

export const loginCallback = (code: string, state: string) =>
  post<{ token: string }>('/user/callback', { code, state })

export const logout = () => post('/user/logout', undefined, { noAlert: true })
