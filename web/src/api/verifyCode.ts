import { post } from '@/utils/http'
import type { Captcha } from './auth'

export type VerifyCodeUse = 'avatar'

export const sendSms = (phone: string, use_for: VerifyCodeUse, captcha: Captcha) =>
  post('/verify_code/sms', { phone, use_for, captcha })

export const sendEmail = (email: string, use_for: VerifyCodeUse, captcha: Captcha) =>
  post('/verify_code/email', { email, use_for, captcha })
