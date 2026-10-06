import { describe, expect, test } from 'vitest'
import { formatDate, formatNumber, shortHash } from '../format'

describe('formatDate', () => {
  test('RFC3339 时间格式化为本地日期', () => {
    const d = new Date('2026-05-06T20:08:54Z')
    const pad = (n: number) => String(n).padStart(2, '0')
    const expected = `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
    expect(formatDate('2026-05-06T20:08:54Z')).toBe(expected)
  })

  test('空值返回空字符串，无法解析的原样返回', () => {
    expect(formatDate('')).toBe('')
    expect(formatDate(null)).toBe('')
    expect(formatDate(undefined)).toBe('')
    expect(formatDate('not a date')).toBe('not a date')
  })
})

describe('formatNumber', () => {
  test('取整并加千位分隔符', () => {
    expect(formatNumber(10000000)).toBe('10,000,000')
    expect(formatNumber(1234.6)).toBe('1,235')
  })
})

describe('shortHash', () => {
  test('保留首尾，中间用省略号', () => {
    expect(shortHash('0123456789abcdef0123456789abcdef')).toBe('01234567…abcdef')
    expect(shortHash('0123456789abcdef0123456789abcdef', 4, 2)).toBe('0123…ef')
  })

  test('不比首尾更长时原样返回', () => {
    expect(shortHash('0123456789abcde')).toBe('0123456789abcde')
    expect(shortHash('')).toBe('')
  })
})
