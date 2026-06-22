import { onMounted, onUnmounted, type Ref } from 'vue'

type AnimationName = 'fade-up' | 'fade-left' | 'fade-right' | 'zoom-in' | 'fade-in'

const defaultOptions = {
  threshold: 0.15,
  rootMargin: '0px 0px -50px 0px',
}

export function useScrollReveal(containerRef: Ref<HTMLElement | null>) {
  let observer: IntersectionObserver | null = null

  function init(_animations: Record<string, AnimationName> = {}) {
    if (!containerRef.value) return

    observer = new IntersectionObserver(
      (entries) => {
        for (const entry of entries) {
          const el = entry.target as HTMLElement
          const animClass = (el.dataset.reveal as AnimationName) || 'fade-up'
          if (entry.isIntersecting) {
            el.classList.add('reveal-active', `${animClass}-active`)
            el.classList.remove(animClass)
          } else {
            el.classList.remove('reveal-active', `${animClass}-active`)
            el.classList.add(animClass)
          }
        }
      },
      defaultOptions,
    )

    const elements = containerRef.value.querySelectorAll('[data-reveal]')
    elements.forEach((el) => {
      const anim = (el as HTMLElement).dataset.reveal as AnimationName
      if (anim) {
        el.classList.add('reveal', anim)
      }
      observer?.observe(el)
    })
  }

  onMounted(() => {
    init()
  })

  onUnmounted(() => {
    observer?.disconnect()
  })

  return { init }
}
