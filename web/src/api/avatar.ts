import { http } from '@/utils/http'

export default {
  list: (page: number, limit: number): any => http.Get('/avatars', { params: { page, limit } }),
  create: (data: FormData): any => http.Post('/avatars', data),
  update: (hash: string, data: FormData): any => http.Put(`/avatars/${hash}`, data),
  delete: (hash: string): any => http.Delete(`/avatars/${hash}`),
  check: (raw: string): any => http.Get('/avatars/check', { params: { raw } }),
  qq: (qq: string): any => http.Get('/avatars/qq', { params: { qq } })
}
