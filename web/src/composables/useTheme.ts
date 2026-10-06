import { nextTick } from 'vue'
import { useDark, usePreferredReducedMotion } from '@vueuse/core'

const isDark = useDark({ storageKey: 'weavatar-theme' })
const reducedMotion = usePreferredReducedMotion()

type DocWithVT = Document & {
  startViewTransition?: (cb: () => Promise<void> | void) => { ready: Promise<void> }
}

/** 支持 View Transitions 时从点击位置圆形扩散切换主题 */
export function useTheme() {
  const toggle = (event?: MouseEvent) => {
    const next = !isDark.value
    const doc = document as DocWithVT

    if (!doc.startViewTransition || reducedMotion.value === 'reduce') {
      isDark.value = next
      return
    }

    const x = event?.clientX ?? window.innerWidth / 2
    const y = event?.clientY ?? 0
    const radius = Math.hypot(
      Math.max(x, window.innerWidth - x),
      Math.max(y, window.innerHeight - y)
    )

    const transition = doc.startViewTransition(async () => {
      isDark.value = next
      await nextTick()
    })

    transition.ready.then(() => {
      document.documentElement.animate(
        {
          clipPath: [`circle(0px at ${x}px ${y}px)`, `circle(${radius}px at ${x}px ${y}px)`]
        },
        {
          duration: 520,
          easing: 'cubic-bezier(0.4, 0, 0.2, 1)',
          pseudoElement: '::view-transition-new(root)'
        }
      )
    })
  }

  return { isDark, toggle }
}
