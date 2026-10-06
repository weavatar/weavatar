import { describe, expect, test } from 'vitest'
import { safeRedirect } from '../redirect'

describe('safeRedirect', () => {
  test('接受站内路径', () => {
    expect(safeRedirect('/user/avatar')).toBe('/user/avatar')
    expect(safeRedirect('/user/info?tab=a&page=2')).toBe('/user/info?tab=a&page=2')
  })

  test('拒绝协议相对地址与反斜杠绕过', () => {
    expect(safeRedirect('//evil.com')).toBe('')
    expect(safeRedirect('/\\evil.com')).toBe('')
  })

  test('拒绝绝对地址与非字符串', () => {
    expect(safeRedirect('https://evil.com')).toBe('')
    expect(safeRedirect('user/avatar')).toBe('')
    expect(safeRedirect(undefined)).toBe('')
    expect(safeRedirect(['/user/avatar'])).toBe('')
  })
})
