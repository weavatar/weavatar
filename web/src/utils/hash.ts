/**
 * 计算 WeAvatar 头像哈希：去除首尾空格 → 转小写 → SHA256（十六进制）
 */
export function normalizeRaw(raw: string) {
  return raw.trim().toLowerCase()
}

export async function sha256(input: string): Promise<string> {
  const data = new TextEncoder().encode(input)
  const digest = await crypto.subtle.digest('SHA-256', data)
  return Array.from(new Uint8Array(digest))
    .map((b) => b.toString(16).padStart(2, '0'))
    .join('')
}

export async function avatarHash(raw: string) {
  return sha256(normalizeRaw(raw))
}
