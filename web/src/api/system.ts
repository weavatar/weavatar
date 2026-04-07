import { http } from '@/utils/http'

export default {
  count: (): any => http.Get('/system/count'),
  randomAvatars: (): any => http.Get('/system/random_avatars')
}
