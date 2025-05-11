<template>
  <main>
    <!-- Hero Section -->
    <div class="hero-section bg-white">
      <div class="mx-auto py-12 lg:py-32">
        <div class="grid grid-cols-1 lg:grid-cols-[55%_45%] items-center max-w-7xl mx-auto px-6">
          <!-- 左侧文字区域 -->
          <div class="hero-text order-2 lg:order-1">
            <div class="space-y-8 lg:space-y-12 text-center lg:text-left">
              <h1 class="text-4xl sm:text-5xl lg:text-6xl xl:text-7xl font-bold text-blue-600 !leading-[1.2]">
                您的免费网络资料卡
              </h1>
              <p class="text-lg lg:text-xl text-gray-600 max-w-2xl mx-auto lg:mx-0 leading-relaxed">
                将您的邮箱或手机号变成您的数字护照，您在互联网上发帖、评论或在线互动均可使用。
              </p>
              <p class="text-base lg:text-lg text-gray-500">
                一次设置，随处可见。
              </p>
              <div class="flex flex-col sm:flex-row gap-4 justify-center lg:justify-start pt-4">
                <NButton
                  type="info"
                  size="large"
                  class="text-base px-8 h-12"
                  @click="handleStartUse"
                >
                  开始使用
                </NButton>
                <NButton
                  type="default"
                  size="large"
                  class="text-base px-8 h-12"
                  @click="handleStartUse"
                >
                  浏览文档
                </NButton>
              </div>
              <div class="flex items-center justify-center lg:justify-start">
                <span>我们昨天共响应了</span>
                <span class="mx-2 text-2xl font-mono font-bold text-blue-600">
                  <NNumberAnimation
                  ref="usageRef"
                  :from="0"
                  :to="usage"
                  :active="false"
                  :duration="3000"
                  show-separator
                />
                </span>
                <span>次请求</span>
              </div>
            </div>
          </div>

          <!-- 右侧头像展示区域 -->
          <div class="hidden lg:flex order-1 lg:order-2 relative h-[400px] lg:h-[600px] items-center justify-center overflow-hidden">
            <div class="avatar-container relative w-full max-w-lg">
              <div class="avatar-columns flex gap-12 justify-center">
                <div class="avatar-column">
                  <div class="scroll-container">
                    <div class="scroll-content">
                      <div class="flex flex-col gap-4">
                        <NAvatar
                          v-for="(avatar, index) in column1"
                          :key="avatar + index"
                          :src="avatar"
                          round
                          :size="56"
                          class="avatar-item shadow-sm"
                        />
                      </div>
                    </div>
                  </div>
                </div>
                <div class="avatar-column" style="--scroll-duration: 25s;">
                  <div class="scroll-container">
                    <div class="scroll-content scroll-reverse">
                      <div class="flex flex-col gap-4">
                        <NAvatar
                          v-for="(avatar, index) in column2"
                          :key="avatar + index"
                          :src="avatar"
                          round
                          :size="56"
                          class="avatar-item shadow-sm"
                        />
                      </div>
                    </div>
                  </div>
                </div>
                <div class="avatar-column" style="--scroll-duration: 20s;">
                  <div class="scroll-container">
                    <div class="scroll-content">
                      <div class="flex flex-col gap-4">
                        <NAvatar
                          v-for="(avatar, index) in column3"
                          :key="avatar + index"
                          :src="avatar"
                          round
                          :size="56"
                          class="avatar-item shadow-sm"
                        />
                      </div>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Features Section -->
    <div class="features-section py-16 sm:py-24 bg-gray-50 dark:bg-gray-800">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-16">
          <h2 class="text-3xl sm:text-4xl font-bold mb-4">为什么选择 WeAvatar</h2>
          <p class="text-gray-600">
            WeAvatar 是超越 Gravatar 的新一代头像服务，相比 Gravatar 具有以下优势
          </p>
        </div>
        <div class="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-8">
          <div
            v-for="(feature, index) in features"
            :key="index"
            class="feature-card bg-white dark:bg-black rounded-lg overflow-hidden border-0 shadow-sm hover:shadow-xl hover:translate-y-[-5px] transition-all duration-300 relative"
          >
            <div
              class="absolute inset-0 rounded-lg opacity-0 hover:opacity-100 transition-opacity duration-300"
            ></div>
            <div class="p-6 flex flex-col h-full relative z-10">
              <div
                class="bg-blue/10 mb-6 flex items-center justify-center w-16 h-16 rounded-full"
              >
                <NIcon :size="32">
                  <component :is="feature.icon" />
                </NIcon>
              </div>
              <h3 class="text-xl font-bold mb-3 text-gray-800">{{ feature.title }}</h3>
              <p class="text-gray-600 flex-grow" v-html="feature.description"></p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Users Section -->
    <div class="py-16 sm:py-24 bg-gray-50 dark:bg-gray-800">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-16">
          <h2 class="text-3xl sm:text-4xl font-bold mb-4">他们都在用</h2>
          <p class="text-gray-600">一些你可能认识的 TA 也在使用 WeAvatar，不妨来试试？</p>
        </div>
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
          <a
            v-for="(user, index) in users"
            :key="index"
            :href="user.url"
            target="_blank"
            class="user-card group relative p-4 rounded-xl border border-gray-200 dark:border-gray-700 backdrop-blur-sm hover:shadow-lg hover:border-primary/30 transition-all duration-300 flex items-center justify-center w-full mx-auto h-30"
          >
            <div
              class="absolute top-2 right-2 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <LinkExternalIcon class="h-4 w-4" />
            </div>
            <div>
              <NImage
                :src="user.logo"
                class="w-full h-16 object-contain mx-auto"
                preview-disabled
                fallback-src="/placeholder-logo.png"
              />
            </div>
          </a>
        </div>
      </div>
    </div>

    <!-- Sponsors Section -->
    <div class="py-16 sm:py-24 bg-gray-50 dark:bg-gray-800">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
        <div class="text-center mb-16">
          <h2 class="text-3xl sm:text-4xl font-bold mb-4">我们的赞助商</h2>
          <p class="text-gray-600">作为公益性质的项目，WeAvatar 的稳定运行离不开它们的帮助</p>
        </div>
        <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-6">
          <a
            v-for="(sponsor, index) in sponsors"
            :key="index"
            :href="sponsor.url"
            target="_blank"
            class="sponsor-card group relative p-4 rounded-xl border border-gray-200 dark:border-gray-700 backdrop-blur-sm hover:shadow-lg hover:border-primary/30 transition-all duration-300 flex items-center justify-center w-full mx-auto h-30"
          >
            <div
              class="absolute top-2 right-2 text-gray-400 opacity-0 group-hover:opacity-100 transition-opacity"
            >
              <LinkExternalIcon class="h-4 w-4" />
            </div>
            <div>
              <NImage
                :src="sponsor.logo"
                class="w-full h-16 object-contain mx-auto"
                preview-disabled
                fallback-src="/placeholder-logo.png"
              />
            </div>
          </a>
        </div>
      </div>
    </div>
  </main>
</template>

<script setup lang="ts">
import type { NumberAnimationInst } from 'naive-ui'
import { ref, nextTick } from 'vue'
import { useRouter } from 'vue-router'
import { NButton, NNumberAnimation, NIcon, NImage, NAvatar } from 'naive-ui'
import {
  AlbumsOutline as AlbumsIcon,
  CloudOutline as CloudOutlineIcon,
  IdCardOutline as IdCardIcon,
  ImageOutline as ImageIcon,
  SpeedometerOutline as SpeedometerIcon,
  ShieldCheckmarkOutline as ShieldCheckmarkIcon,
  OpenOutline as LinkExternalIcon
} from '@vicons/ionicons5'
import { fetchCdnUsage } from '@/api/system'

// 导入图片资源
import logo_zmingcx from '@/assets/logo-zmingcx.png'
import logo_nicetheme from '@/assets/logo-nicetheme.png'
import logo_artalk from '@/assets/logo-artalk.png'
import logo_iro from '@/assets/logo-iro.png'
import logo_lxtx from '@/assets/logo-lxtx.png'
import logo_liyang from '@/assets/logo-liyang.png'
import logo_twikoo from '@/assets/logo-twikoo.png'
import logo_dami from '@/assets/logo-dami.png'
import logo_dunyun from '@/assets/logo-dunyun.png'
import logo_wafpro from '@/assets/logo-wafpro.png'
import logo_wuwei from '@/assets/logo-wuwei.png'
import logo_anycast from '@/assets/logo-anycast.png'

const router = useRouter()
const usage = ref(0)
const usageRef = ref<NumberAnimationInst | null>(null)

fetchCdnUsage()
  .then((res) => {
    usage.value = res.data.usage
    nextTick(() => {
      usageRef.value?.play()
    })
  })
  .catch((err) => {
    if (err.code != 422) {
      window.$message.error(err.msg)
    }
    console.log(err)
  })

// 生成随机头像的函数
const generateRandomAvatar = () => {
  const styles = ['identicon', 'monsterid', 'wavatar', 'retro', 'robohash', 'color', 'initials']
  const style = styles[Math.floor(Math.random() * styles.length)]
  const seed = Math.random().toString(36).substring(2)
  if (style == 'initials') {
    const letters = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ'
    const firstLetter = letters[Math.floor(Math.random() * letters.length)]
    const secondLetter = letters[Math.floor(Math.random() * letters.length)]
    // 再随机一次判断是两个字母还是一个字母
    const isName = Math.random() < 0.5
    if (isName) {
      return `https://weavatar.com/avatar/${seed}?d=${style}&name=${firstLetter}`
    }
    return `https://weavatar.com/avatar/${seed}?d=${style}&initials=${firstLetter}${secondLetter}`
  }
  return `https://weavatar.com/avatar/${seed}?d=${style}`
}

// 为每列生成足够多的头像以确保滚动效果
const generateColumnAvatars = () => Array.from({ length: 15 }, generateRandomAvatar)

const column1 = ref(generateColumnAvatars())
const column2 = ref(generateColumnAvatars())
const column3 = ref(generateColumnAvatars())

const handleStartUse = () => {
  router.push({ name: 'login' })
}

// 特性数据
const features = [
  {
    icon: AlbumsIcon,
    title: '多级头像匹配',
    description:
      'WeAvatar 除用户上传的头像外，同时支持从 Gravatar、QQ 获取头像，这可为 <b>70%</b> 以上的请求提供准确的头像'
  },
  {
    icon: IdCardIcon,
    title: '手机号、字母头像',
    description:
      'WeAvatar 首家支持手机号头像及字母默认头像，手机号头像更符合国内的使用习惯，字母头像可为没有头像的用户提供更好的体验'
  },
  {
    icon: ImageIcon,
    title: 'WEBP 支持',
    description: 'WeAvatar 默认使用新一代的图像格式 WEBP，这可减少约 <b>80%</b> 的流量消耗'
  },
  {
    icon: ShieldCheckmarkIcon,
    title: '安全',
    description: 'WeAvatar 拥有 AI 自动化审核，确保不会有违规内容被输出'
  },
  {
    icon: SpeedometerIcon,
    title: '更快的速度',
    description:
      'WeAvatar 使用 Go 开发，比同类产品速度更快。WeAvatar 还采用多级缓存机制，以尽可能提高头像的加载速度'
  },
  {
    icon: CloudOutlineIcon,
    title: '开放平台（规划中）',
    description:
      'WeAvatar 未来将为开发者提供开放平台和配套的 SDK，可将自己的应用无缝对接至 WeAvatar'
  }
]

// 用户数据
const users = [
  { url: 'https://zmingcx.com/', logo: logo_zmingcx },
  { url: 'https://www.nicetheme.cn/', logo: logo_nicetheme },
  { url: 'https://www.ilxtx.com/', logo: logo_lxtx },
  { url: 'https://github.com/mirai-mamori/Sakurairo', logo: logo_iro },
  { url: 'https://blog.qqsuu.cn/', logo: logo_dami },
  { url: 'https://artalk.js.org/', logo: logo_artalk },
  { url: 'https://twikoo.js.org/', logo: logo_twikoo },
  { url: 'https://www.liblog.cn/', logo: logo_liyang }
]

// 赞助商数据
const sponsors = [
  { url: 'https://www.ddunyun.com/aff/PNYAXMKI', logo: logo_dunyun },
  { url: 'https://waf.pro/', logo: logo_wafpro },
  { url: 'https://su.sctes.com/register?code=8st689ujpmm2p', logo: logo_wuwei },
  { url: 'https://www.anycast.ai', logo: logo_anycast }
]
</script>

<style scoped>
.hero-section {
  min-height: auto;
  padding: 2rem 0;
}

@media (min-width: 1024px) {
  .hero-section {
    min-height: 100vh;
    padding: 0;
  }
}

.avatar-container {
  mask-image: linear-gradient(to bottom, transparent 0%, black 10%, black 90%, transparent 100%);
}

.avatar-columns {
  position: relative;
}

.avatar-column {
  --scroll-duration: 30s;
  height: 500px;
  overflow: hidden;
}

.scroll-container {
  height: 100%;
  overflow: hidden;
}

.scroll-content {
  animation: scroll var(--scroll-duration) linear infinite;
}

.scroll-reverse {
  animation-direction: reverse;
}

.avatar-item {
  transition: all 0.3s ease;
}

.avatar-item:hover {
  transform: scale(1.1);
  z-index: 10;
}

@keyframes scroll {
  0% {
    transform: translateY(0);
  }
  100% {
    transform: translateY(-50%);
  }
}

@media (max-width: 768px) {
  .hero-section {
    min-height: auto;
    padding-top: 2rem;
  }
}

@media (prefers-reduced-motion: reduce) {
  .scroll-content {
    animation: none;
  }
}

.user-card,
.sponsor-card {
  background-color: rgba(255, 255, 255, 0.05);
  backdrop-filter: blur(10px);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.05);
}

.dark .user-card,
.dark .sponsor-card {
  background-color: rgba(0, 0, 0, 0.2);
  box-shadow: 0 2px 10px rgba(0, 0, 0, 0.2);
}

/* 标签卡悬停效果 */
.user-card:hover,
.sponsor-card:hover {
  transform: translateY(-5px);
  background-color: rgba(255, 255, 255, 0.1);
}

.dark .user-card:hover,
.dark .sponsor-card:hover {
  background-color: rgba(30, 30, 30, 0.4);
}
</style>
