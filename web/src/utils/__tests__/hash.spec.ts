import { describe, expect, test } from 'vitest'
import { avatarHash, normalizeRaw } from '../hash'

describe('normalizeRaw', () => {
  test('去除首尾空白并转为小写', () => {
    expect(normalizeRaw(' Hi@WeAvatar.com ')).toBe('hi@weavatar.com')
  })
})

describe('avatarHash', () => {
  test('与规范化后地址的 SHA256 一致', async () => {
    expect(await avatarHash(' Hi@WeAvatar.com ')).toBe(
      '8ef53e597e8b7b639ce4fc3997e88787f9a1792c5338e681f79990121b92b563'
    )
  })
})
