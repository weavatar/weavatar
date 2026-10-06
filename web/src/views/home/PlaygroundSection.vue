<template>
  <section id="playground" class="relative section border-t border-line">
    <div class="wrap">
      <section-title title="在线体验" desc="输入邮箱或手机号，即可生成头像地址。" />

      <div v-reveal class="mt-12 surface overflow-hidden grid lg:grid-cols-[1fr_22rem]">
        <!-- 参数 -->
        <div class="p-6 sm:p-8 flex flex-col gap-7">
          <label class="block">
            <span class="block text-sm font-600 text-fg mb-2">邮箱 / 手机号</span>
            <span class="relative block">
              <span
                class="i-lucide-at-sign absolute left-4 top-1/2 -translate-y-1/2 text-fg3 text-base"
              />
              <input
                v-model="raw"
                type="text"
                autocomplete="off"
                spellcheck="false"
                placeholder="hi@weavatar.com"
                class="w-full h-12 pl-11 pr-4 rounded-xl bg-bg border border-line-strong text-[15px] font-mono outline-none transition-[border-color,box-shadow] focus:border-brand focus:ring-4 focus:ring-brand/15 placeholder:text-fg3"
              />
            </span>
          </label>

          <div>
            <div class="flex items-center justify-between mb-2">
              <span class="text-sm font-600 text-fg">默认头像样式</span>
              <button
                type="button"
                role="switch"
                :aria-checked="force"
                class="inline-flex items-center gap-2 text-xs text-fg2 hover:text-fg"
                @click="force = !force"
              >
                <span
                  class="relative w-8 h-[18px] rounded-full transition-colors"
                  :class="force ? 'bg-brand' : 'bg-fg/20'"
                >
                  <span
                    class="absolute left-0 top-[2px] w-[14px] h-[14px] rounded-full bg-white shadow transition-[translate] duration-200"
                    :class="force ? 'translate-x-[16px]' : 'translate-x-[2px]'"
                  />
                </span>
                强制默认头像
              </button>
            </div>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="opt in defaults"
                :key="opt.key"
                type="button"
                :class="d === opt.key ? 'chip-brand' : 'chip hover:border-fg/30 hover:text-fg'"
                @click="d = opt.key"
              >
                {{ opt.label }}
              </button>
            </div>
          </div>

          <div>
            <span class="block text-sm font-600 text-fg mb-2">尺寸</span>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="s in sizes"
                :key="s"
                type="button"
                class="font-mono"
                :class="size === s ? 'chip-brand' : 'chip hover:border-fg/30 hover:text-fg'"
                @click="size = s"
              >
                {{ s }}
              </button>
            </div>
          </div>

          <div>
            <span class="block text-sm font-600 text-fg mb-2">格式</span>
            <div class="flex flex-wrap gap-2">
              <button
                v-for="f in formats"
                :key="f"
                type="button"
                class="font-mono"
                :class="format === f ? 'chip-brand' : 'chip hover:border-fg/30 hover:text-fg'"
                @click="format = f"
              >
                {{ f }}
              </button>
            </div>
          </div>

          <div class="grid gap-4 pt-2 border-t border-line">
            <div class="pt-4">
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-xs font-600 text-fg3 uppercase tracking-wider">SHA256</span>
                <copy-button v-if="hash" :text="hash" />
              </div>
              <code class="block text-[13px] font-mono text-fg2 break-all leading-relaxed min-h-5">
                {{ hash || '请输入邮箱或手机号' }}
              </code>
            </div>
            <div>
              <div class="flex items-center justify-between mb-1.5">
                <span class="text-xs font-600 text-fg3 uppercase tracking-wider">头像地址</span>
                <copy-button v-if="url" :text="url" label="复制地址" />
              </div>
              <code
                class="block text-[13px] font-mono text-fg break-all leading-relaxed min-h-[2lh]"
              >
                {{ url || '—' }}
              </code>
            </div>
          </div>
        </div>

        <!-- 预览 -->
        <div
          class="relative border-t lg:border-t-0 lg:border-l border-line wa-dots p-8 flex flex-col items-center justify-center gap-5 min-h-72"
        >
          <div class="relative">
            <div class="absolute -inset-6 rounded-[2rem] bg-brand/10 blur-2xl" aria-hidden="true" />
            <avatar-image
              v-if="url"
              :src="url"
              :size="176"
              rounded="2xl"
              alt="头像预览"
              class="relative ring-1 ring-line shadow-card-hover"
            />
            <div
              v-else
              class="relative w-44 h-44 rounded-2xl border border-dashed border-line-strong flex items-center justify-center text-fg3"
            >
              <span class="i-lucide-image text-3xl" />
            </div>
          </div>
          <div class="text-xs font-mono text-fg3 tabular">
            {{ size }} × {{ size }} · {{ format }}
          </div>
          <div v-if="previewUnsupported" class="text-xs text-fg3">当前浏览器可能无法预览该格式</div>
          <a v-if="url" :href="url" target="_blank" rel="noreferrer" class="btn-secondary btn-sm">
            在新窗口打开
            <span class="i-lucide-arrow-up-right text-sm" />
          </a>
        </div>
      </div>
    </div>
  </section>
</template>

<script setup lang="ts">
import { computed, ref, watch } from 'vue'
import { refDebounced } from '@vueuse/core'
import SectionTitle from '@/components/ui/SectionTitle.vue'
import CopyButton from '@/components/ui/CopyButton.vue'
import AvatarImage from '@/components/ui/AvatarImage.vue'
import { avatarHash } from '@/utils/hash'
import { AVATAR_BASE } from '@/constants/links'

const defaults = [
  { key: '', label: '系统默认' },
  { key: 'mp', label: '人物剪影' },
  { key: 'identicon', label: '几何图案' },
  { key: 'monsterid', label: '小怪物' },
  { key: 'wavatar', label: '卡通人脸' },
  { key: 'retro', label: '像素风' },
  { key: 'robohash', label: '机器人' },
  { key: 'color', label: '纯色' },
  { key: 'initials', label: '字母' }
]
const sizes = [80, 160, 240, 400]
const formats = ['webp', 'png', 'jpg', 'jpeg', 'gif', 'avif', 'heic', 'heif', 'jxl', 'tiff']
const browserUnsupported = ['heic', 'heif', 'jxl', 'tiff']

const raw = ref('hi@weavatar.com')
const debounced = refDebounced(raw, 160)
const hash = ref('')
const d = ref('')
const size = ref(160)
const format = ref('webp')
const previewUnsupported = computed(() => browserUnsupported.includes(format.value))
const force = ref(false)

watch(
  debounced,
  async (value) => {
    const v = value.trim()
    if (!v) {
      hash.value = ''
      return
    }
    const h = await avatarHash(v)
    // 防止快速输入时旧结果覆盖新结果
    if (debounced.value.trim() === v) hash.value = h
  },
  { immediate: true }
)

const url = computed(() => {
  if (!hash.value) return ''
  const params = new URLSearchParams({ s: String(size.value) })
  if (d.value) params.set('d', d.value)
  if (d.value === 'initials') params.set('initials', raw.value.trim().slice(0, 2))
  if (force.value) params.set('f', 'y')
  const ext = format.value === 'webp' ? '' : `.${format.value}`
  return `${AVATAR_BASE}/${hash.value}${ext}?${params.toString()}`
})
</script>
