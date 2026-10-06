import { http } from '@/utils/http'

export default {
  info: (): any => http.Get('/user/info'),
  update: (data: any): any => http.Put('/user/info', data),
  qq: (qq: string): any => http.Get('/avatars/qq', { params: { qq } }),
  deletionLogin: (): any => http.Get('/user/deletion/login'),
  deletionConfirm: (code: string, state: string): any =>
    http.Post('/user/deletion/confirm', { code, state })
}
