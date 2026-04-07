import { http } from '@/utils/http'

export default {
  sms: (phone: string, use_for: string, captcha: any): any =>
    http.Post('/verify_code/sms', { phone, use_for, captcha }),
  email: (email: string, use_for: string, captcha: any): any =>
    http.Post('/verify_code/email', { email, use_for, captcha })
}
