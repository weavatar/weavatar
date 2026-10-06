import { get } from '@/utils/http'

export const fetchUsage = () =>
  get<{ usage: number }>('/system/count', undefined, { noAlert: true })

export const fetchRandomAvatars = () =>
  get<{ avatars: string[] }>('/system/random_avatars', undefined, { noAlert: true })
