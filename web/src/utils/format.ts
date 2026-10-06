/** RFC3339 时间格式化为本地日期 YYYY-MM-DD */
export function formatDate(value: string | null | undefined) {
  if (!value) return ''
  const d = new Date(value)
  if (Number.isNaN(d.getTime())) return value
  const pad = (n: number) => String(n).padStart(2, '0')
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`
}

export function formatNumber(n: number) {
  return new Intl.NumberFormat('en-US').format(Math.round(n))
}

export function shortHash(hash: string, head = 8, tail = 6) {
  if (!hash || hash.length <= head + tail + 1) return hash
  return `${hash.slice(0, head)}…${hash.slice(-tail)}`
}
