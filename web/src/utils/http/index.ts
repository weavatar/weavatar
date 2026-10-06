import { resolveResError } from '@/utils/http/helpers'
import { createAlova } from 'alova'
import adapterFetch from 'alova/fetch'
import VueHook from 'alova/vue'
import { useUserStore } from '@/stores'

/**
 * 请求失败时抛出的错误：按 HTTP 状态码分流，机器可读的业务码在 errorCode（可选）。
 * 后端错误信封为 { msg, code? }，code 为字符串形式的业务错误码，仅业务错误携带，不用于分流。
 */
export class HttpError extends Error {
  status: number
  errorCode?: string

  constructor(status: number, msg: string, errorCode?: string) {
    super(msg)
    this.name = 'HttpError'
    this.status = status
    this.errorCode = errorCode
  }
}

export const http = createAlova({
  baseURL: import.meta.env.VITE_API_URL,
  statesHook: VueHook,
  requestAdapter: adapterFetch(),
  cacheFor: null,
  beforeRequest: (method) => {
    const userStore = useUserStore()
    if (userStore.auth.token) {
      method.config.headers['Authorization'] = `Bearer ${userStore.auth.token}`
    }
  },
  responded: {
    onSuccess: async (response: any, method: any) => {
      const json = await response.json().catch(() => ({}))
      const { status } = response
      const { meta } = method

      if (status !== 200) {
        const msg = resolveResError(
          status,
          (typeof json?.msg === 'string' && json.msg.trim()) || response.statusText
        )
        if (!meta?.noAlert) {
          if (status === 422) {
            window.$message.error(msg)
          } else if (status !== 401) {
            window.$dialog.error({
              title: '错误',
              content: msg,
              maskClosable: false
            })
          }
        }
        throw new HttpError(status, msg, typeof json?.code === 'string' ? json.code : undefined)
      }

      return json.data
    },
    // 仅在请求未得到响应（网络异常）时触发，onSuccess 中抛出的错误不会进入这里
    onError: (error: any, method: any) => {
      if (!method.meta?.noAlert) {
        window.$dialog.error({
          title: '请求失败',
          content: '网络连接失败，请稍后重试',
          maskClosable: false
        })
      }

      throw error
    }
  }
})
