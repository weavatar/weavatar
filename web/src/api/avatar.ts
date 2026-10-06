import { del, get, post, put } from '@/utils/http'

export interface Avatar {
  sha256: string
  md5: string
  raw: string
  created_at: string
}

export const fetchAvatars = (page: number, limit: number) =>
  get<{ total: number; items: Avatar[] }>('/avatars', { page, limit })

/** data 为表单：raw、verify_code、avatar，以及 JSON 字符串形式的 captcha */
export const createAvatar = (data: FormData) => post<Avatar>('/avatars', data)

/** data 为表单：avatar，以及 JSON 字符串形式的 captcha */
export const updateAvatar = (hash: string, data: FormData) => put<Avatar>(`/avatars/${hash}`, data)

export const deleteAvatar = (hash: string) => del(`/avatars/${hash}`)

/** bind 表示该地址已有头像 */
export const checkAvatar = (raw: string) => get<{ bind: boolean }>('/avatars/check', { raw })

/** 返回 base64 编码的 PNG */
export const fetchQqAvatar = (qq: string) => get<string>('/avatars/qq', { qq })
