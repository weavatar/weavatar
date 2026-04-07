import { useUserStore } from '@/stores'

export function resolveResError(code: number | string | undefined, msg = ''): string {
  switch (code) {
    case 400:
    case 422:
      msg = msg || '请求参数错误'
      break
    case 401:
      msg = msg || '登录已过期'
      useUserStore().clearToken()
      break
    case 403:
      msg = msg || '没有权限'
      break
    case 404:
      msg = msg || '资源或接口不存在'
      break
    case 500:
      msg = msg || '服务器异常'
      break
    default:
      msg = msg || `【${code}】: 未知异常!`
      break
  }
  return msg
}
