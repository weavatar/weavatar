export const isPhone = (val: string) => /^1[3-9]\d{9}$/.test(val)
export const isEmail = (val: string) => /^[\w.+-]+@[\w-]+(\.[\w-]+)+$/.test(val)
export const isPhoneOrEmail = (val: string) => isPhone(val) || isEmail(val)
