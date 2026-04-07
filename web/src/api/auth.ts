import { http } from '@/utils/http'

export default {
  login: (): any => http.Get('/user/login'),
  callback: (code: string, state: string): any => http.Post('/user/callback', { code, state }),
  logout: (): any => http.Post('/user/logout')
}
