import { createAlova, type RequestBody } from 'alova'
import adapterFetch from 'alova/fetch'
import { useUserStore } from '@/stores'

/** 网络错误时 status 为 0；errorCode 是后端可选的业务错误码，分流只看 status */
export class ApiError extends Error {
  status: number
  errorCode?: string

  constructor(status: number, msg: string, errorCode?: string) {
    super(msg)
    this.name = 'ApiError'
    this.status = status
    this.errorCode = errorCode
  }
}

export interface RequestConfig {
  /** 失败时不弹出错误提示 */
  noAlert?: boolean
}

declare module 'alova' {
  interface AlovaCustomTypes {
    meta: RequestConfig
  }
}

const alertError = (err: ApiError) => window.$message?.error(err.message)

const service = createAlova({
  baseURL: import.meta.env.VITE_API_URL,
  requestAdapter: adapterFetch(),
  timeout: 30000,
  cacheFor: null,
  beforeRequest(method) {
    const token = useUserStore().auth.token
    if (token) method.config.headers.Authorization = `Bearer ${token}`
  },
  responded: {
    async onSuccess(response, method) {
      const payload = await response.json().catch(() => ({}))
      if (response.ok) return payload.data

      const status = response.status
      const serverMsg = typeof payload.msg === 'string' ? payload.msg.trim() : ''
      const fallback = status === 401 ? '登录已过期，请重新登录' : `请求失败（${status}）`
      const err = new ApiError(
        status,
        serverMsg || fallback,
        typeof payload.code === 'string' ? payload.code : undefined
      )

      if (status === 401) useUserStore().clearToken()
      if (!method.meta?.noAlert) alertError(err)
      throw err
    },
    // 只在请求没有得到响应时触发，onSuccess 抛出的错误不会进入这里
    onError(_error, method) {
      const err = new ApiError(0, '网络连接失败，请稍后重试')
      if (!method.meta?.noAlert) alertError(err)
      throw err
    }
  }
})

// 返回的 Method 在 await 或 then 时才发出请求
export const get = <T = unknown>(
  url: string,
  params?: Record<string, unknown>,
  config?: RequestConfig
) => service.Get<T>(url, { params, meta: config })
export const post = <T = unknown>(url: string, data?: RequestBody, config?: RequestConfig) =>
  service.Post<T>(url, data, { meta: config })
export const put = <T = unknown>(url: string, data?: RequestBody, config?: RequestConfig) =>
  service.Put<T>(url, data, { meta: config })
export const del = <T = unknown>(url: string, data?: RequestBody, config?: RequestConfig) =>
  service.Delete<T>(url, data, { meta: config })
