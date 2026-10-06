export const isPhone = (val: string) => /^1[3-9]\d{9}$/.test(val)
export const isEmail = (val: string) => /^[\w.+-]+@[\w-]+(\.[\w-]+)+$/.test(val)
export const isPhoneOrEmail = (val: string) => isPhone(val) || isEmail(val)

/** 仅 http(s) 地址可作为外链，避免 javascript: 等协议 */
export const isHttpUrl = (val: string) => /^https?:\/\//i.test(val)
