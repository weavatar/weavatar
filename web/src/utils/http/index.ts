import { resolveResError } from '@/utils/http/helpers'
import { createAlova } from 'alova'
import adapterFetch from 'alova/fetch'
import VueHook from 'alova/vue'
import { useUserStore } from '@/stores'

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
      const json = await response
        .json()
        .catch(() => ({ code: response.status, msg: response.statusText }))
      const { status } = response
      const { meta } = method

      if (status !== 200) {
        const code = json?.code ?? status
        const msg = resolveResError(
          code,
          (typeof json?.msg === 'string' && json.msg.trim()) || response.statusText
        )
        const noAlert = meta?.noAlert
        if (!noAlert) {
          if (code === 422) {
            window.$message.error(msg)
          } else if (code !== 401) {
            window.$dialog.error({
              title: '错误',
              content: msg,
              maskClosable: false
            })
          }
        }
        throw new Error(msg)
      }

      return json.data
    },
    onError: (error: any, method: any) => {
      const { meta } = method
      const errorMsg = error?.message || '网络请求失败'

      if (!meta?.noAlert) {
        window.$dialog.error({
          title: '请求失败',
          content: errorMsg,
          maskClosable: false
        })
      }

      throw error
    }
  }
})
