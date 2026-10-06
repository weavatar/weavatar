import { onScopeDispose } from 'vue'
import type { Captcha } from '@/api/auth'

export class CaptchaCancelled extends Error {
  constructor() {
    super('captcha cancelled')
    this.name = 'CaptchaCancelled'
  }
}

/** 极验 GT4 实例中用到的部分 */
interface GeetestInstance {
  showCaptcha(): void
  getValidate(): Captcha | false
  onSuccess(cb: () => void): GeetestInstance
  onError(cb: (e: { msg?: string }) => void): GeetestInstance
  onClose(cb: () => void): GeetestInstance
  destroy(): void
}

/** 需配合模板里的 <geetest-captcha @initialized="onInit" />；verify() 关闭、出错或被新调用顶替时 reject CaptchaCancelled */
export function useGeetest() {
  let instance: GeetestInstance | null = null
  let pending: { resolve: (v: Captcha) => void; reject: (e: Error) => void } | null = null
  let disposed = false

  const cancel = () => {
    pending?.reject(new CaptchaCancelled())
    pending = null
  }

  const onInit = (i: GeetestInstance) => {
    // 脚本加载较慢时组件可能已卸载
    if (disposed) return i.destroy()
    instance = i
    i.onSuccess(() => {
      const result = i.getValidate()
      if (result) pending?.resolve(result)
      else pending?.reject(new CaptchaCancelled())
      pending = null
    })
    i.onError((e) => {
      window.$message.error(e?.msg || '验证失败，请重试')
      cancel()
    })
    i.onClose(cancel)
  }

  const verify = () =>
    new Promise<Captcha>((resolve, reject) => {
      if (!instance) {
        window.$message.error('验证组件尚未加载，请稍后重试')
        reject(new CaptchaCancelled())
        return
      }
      cancel()
      pending = { resolve, reject }
      instance.showCaptcha()
    })

  onScopeDispose(() => {
    disposed = true
    cancel()
    instance?.destroy()
    instance = null
  })

  return { onInit, verify }
}
