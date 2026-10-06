import { describe, expect, test } from 'vitest'
import { isEmail, isHttpUrl, isPhone, isPhoneOrEmail } from '../validate'

describe('isPhone', () => {
  test('接受大陆手机号', () => {
    expect(isPhone('13800138000')).toBe(true)
  })

  test('拒绝位数或号段错误', () => {
    expect(isPhone('1380013800')).toBe(false)
    expect(isPhone('12800138000')).toBe(false)
    expect(isPhone('+8613800138000')).toBe(false)
  })
})

describe('isEmail', () => {
  test('接受常见邮箱', () => {
    expect(isEmail('hi@weavatar.com')).toBe(true)
    expect(isEmail('first.last+tag@sub.example.co')).toBe(true)
  })

  test('拒绝缺少域名或 @ 的输入', () => {
    expect(isEmail('hi@weavatar')).toBe(false)
    expect(isEmail('weavatar.com')).toBe(false)
  })
})

describe('isPhoneOrEmail', () => {
  test('手机号或邮箱任一满足即可', () => {
    expect(isPhoneOrEmail('13800138000')).toBe(true)
    expect(isPhoneOrEmail('hi@weavatar.com')).toBe(true)
    expect(isPhoneOrEmail('nope')).toBe(false)
  })
})

describe('isHttpUrl', () => {
  test('只接受 http 与 https', () => {
    expect(isHttpUrl('https://weavatar.com')).toBe(true)
    expect(isHttpUrl('http://example.com/path')).toBe(true)
    expect(isHttpUrl('javascript:alert(1)')).toBe(false)
    expect(isHttpUrl('ftp://example.com')).toBe(false)
  })
})
