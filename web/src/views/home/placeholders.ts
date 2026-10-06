import { AVATAR_BASE } from '@/constants/links'

const styles = ['wavatar', 'robohash', 'monsterid', 'identicon', 'retro']

/** 接口不可用或头像不足时用程序化头像兜底。 */
export function placeholderAvatars(count: number, seed = 'weavatar') {
  return Array.from({ length: count }, (_, i) => {
    const style = styles[i % styles.length]
    return `${AVATAR_BASE}/${seed}-${i}?d=${style}&f=y&s=160`
  })
}

export function fillAvatars(list: string[], count: number) {
  if (list.length >= count) return list.slice(0, count)
  return [...list, ...placeholderAvatars(count - list.length, 'fill')]
}
