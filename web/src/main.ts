import { createApp } from 'vue'
import { MotionPlugin } from '@vueuse/motion'
import { createGtag } from 'vue-gtag'
import { Geetest } from 'vue3-geetest'

import App from './App.vue'
import router from './router'
import { setupStore } from './stores'

import '@/styles/main.scss'
import 'virtual:uno.css'

const app = createApp(App)

setupStore(app)
app.use(router)
app.use(MotionPlugin)
app.use(
  createGtag({
    tagId: 'G-BL3JX2SWLP',
    pageTracker: {
      router
    }
  })
)
app.use(Geetest, {
  captchaId: 'dda5bf0b265affca8bddca9b647ce16b'
})

app.mount('#app')
