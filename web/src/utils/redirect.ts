/** 登录后的回跳地址只接受站内路径，排除 //host 与 /\host，防止开放重定向 */
export function safeRedirect(value: unknown): string {
  return typeof value === 'string' && /^\/(?![/\\])/.test(value) ? value : ''
}
