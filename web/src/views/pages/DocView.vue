<template>
  <div class="wrap">
    <div class="lg:grid lg:grid-cols-[14rem_minmax(0,1fr)] lg:gap-16">
      <!-- 目录 -->
      <aside class="hidden lg:block">
        <nav
          class="sticky top-24 py-16 max-h-[calc(100vh-6rem)] overflow-y-auto"
          aria-label="文档目录"
        >
          <div class="relative pl-4 border-l border-line">
            <span
              class="absolute -left-px w-px bg-brand transition-[top,height] duration-300"
              :style="indicatorStyle"
            />
            <div v-for="group in toc" :key="group.id" class="mb-5">
              <router-link
                :to="{ name: 'doc', hash: `#${group.id}` }"
                :data-toc="group.id"
                class="block py-1 text-sm font-600 transition-colors"
                :class="active === group.id ? 'text-brand' : 'text-fg hover:text-brand'"
              >
                {{ group.label }}
              </router-link>
              <router-link
                v-for="child in group.children"
                :key="child.id"
                :to="{ name: 'doc', hash: `#${child.id}` }"
                :data-toc="child.id"
                class="block py-1 pl-3 text-[13px] transition-colors"
                :class="active === child.id ? 'text-brand' : 'text-fg2 hover:text-fg'"
              >
                {{ child.label }}
              </router-link>
            </div>
          </div>
        </nav>
      </aside>

      <div class="min-w-0 py-14 sm:py-16">
        <h1 class="display-text text-[2rem] sm:text-[2.75rem] text-fg">文档</h1>
        <p class="mt-3 text-base sm:text-[17px] text-fg2 max-w-2xl">
          接口完全兼容 Gravatar，已接入 Gravatar 的站点只需替换域名。
        </p>

        <!-- 移动端目录 -->
        <div
          class="lg:hidden sticky top-16 z-30 -mx-6 sm:-mx-10 mt-8 bg-bg/90 backdrop-blur border-y border-line"
        >
          <div class="flex gap-2 overflow-x-auto px-6 sm:px-10 py-3 [scrollbar-width:none]">
            <router-link
              v-for="group in toc"
              :key="group.id"
              :to="{ name: 'doc', hash: `#${group.id}` }"
              class="chip whitespace-nowrap"
              :class="activeGroup === group.id ? 'chip-brand' : ''"
            >
              {{ group.label }}
            </router-link>
          </div>
        </div>

        <article class="prose mt-6">
          <section id="basic">
            <h2>基本概念</h2>
            <p>WeAvatar 头像 API 可以像普通的图片 URL 一样请求，具体格式是：</p>
            <code-block
              code="https://weavatar.com/avatar/HASH"
              lang="text"
              title="URL"
              class="not-prose my-4"
            />
            <p>
              其中 <code>HASH</code> 部分是邮箱 / 手机号的 <code>SHA256</code> 或
              <code>MD5</code> 哈希值，推荐使用 <code>SHA256</code>。此邮箱 / 手机号须在
              <code>weavatar.com</code> 上添加头像，否则会依次尝试返回 Gravatar
              头像和社交头像，如果都不存在，则返回默认头像。
            </p>
            <div
              class="not-prose my-6 flex flex-col items-start gap-2 text-sm font-600 sm:flex-row sm:flex-wrap sm:items-center sm:gap-3"
            >
              <template v-for="(node, i) in chain" :key="node">
                <span
                  class="inline-flex items-center h-9 px-3.5 rounded-xl border"
                  :class="
                    i === 0
                      ? 'bg-brand text-white border-brand'
                      : 'bg-elev border-line-strong text-fg'
                  "
                >
                  {{ node }}
                </span>
                <template v-if="i < chain.length - 1">
                  <span class="dash-line-y block h-4 ml-4 sm:hidden" />
                  <span class="dash-line hidden sm:block flex-1 min-w-5" />
                </template>
              </template>
            </div>
          </section>

          <section id="hash">
            <h2>邮箱 / 手机号的哈希</h2>
            <ol>
              <li>去除首尾两边的空格</li>
              <li>所有字母转小写</li>
              <li>计算 SHA256 值（或 MD5）</li>
            </ol>
            <code-block :code="hashCode" lang="javascript" title="hash.js" class="not-prose my-4" />
            <p>
              不想自己算？去
              <router-link :to="{ name: 'home', hash: '#playground' }">首页的在线体验</router-link>
              输入邮箱即可得到完整地址。
            </p>
          </section>

          <section id="cms">
            <h2>在 CMS 中使用</h2>

            <section id="wordpress">
              <h3>WordPress</h3>
              <p>
                安装启用
                <a target="_blank" rel="noreferrer" href="https://wp-china-yes.com">WP-China-Yes</a>
                插件并选择头像备用源，你可能还需要关闭主题、其他插件中自带的 Gravatar 头像加速功能。
              </p>
              <p>
                如果你不想安装插件，也可以通过添加以下代码到主题的
                <code>functions.php</code> 文件中来接入 WeAvatar。
              </p>
              <code-block
                :code="wordpressCode"
                lang="php"
                title="functions.php"
                class="not-prose my-4"
              />
            </section>

            <section id="typecho">
              <h3>Typecho</h3>
              <p>添加以下代码到站点根目录的 <code>config.inc.php</code> 中来接入 WeAvatar：</p>
              <code-block
                code="define('__TYPECHO_GRAVATAR_PREFIX__', 'https://weavatar.com/avatar/');"
                lang="php"
                title="config.inc.php"
                class="not-prose my-4"
              />
            </section>

            <section id="emlog">
              <h3>Emlog</h3>
              <p>
                <b>Pro 之前（5.x、6.x 民间版）：</b>通过修改
                <code>include/lib/function.base.php</code> 中 <code>getGravatar</code> 函数里的
                <code>http://www.gravatar.com</code> 为 <code>https://weavatar.com</code> 接入
                WeAvatar，你可以参考
                <router-link :to="{ name: 'doc', hash: '#params' }">额外的参数</router-link>
                部分来修改默认头像。
              </p>
              <p>
                <b>Pro 之后（2.x）：</b>修改 <code>include/lib/common.php</code> 中
                <code>getGravatar</code> 函数里的 <code>cravatar.cn</code> 为
                <code>weavatar.com</code>
                接入 WeAvatar。
              </p>
            </section>

            <section id="zblog">
              <h3>Z-Blog</h3>
              <p>
                后台应用中心搜索 WeAvatar 安装插件，或前往
                <a target="_blank" rel="noreferrer" href="https://app.zblogcn.com/?id=38455">
                  app.zblogcn.com/?id=38455
                </a>
                下载插件并手动安装。
              </p>
            </section>
          </section>

          <section id="comments">
            <h2>在评论系统中使用</h2>
            <section id="twikoo">
              <h3>Twikoo</h3>
              <doc-callout>Twikoo 已默认接入 WeAvatar，无需额外设置。</doc-callout>
            </section>
            <section id="artalk">
              <h3>Artalk</h3>
              <doc-callout>Artalk 已默认接入 WeAvatar，无需额外设置。</doc-callout>
            </section>
          </section>

          <section id="format">
            <h2>指定图片格式</h2>
            <p>我们当前支持 10 种图片返回格式：</p>
            <div class="not-prose my-4 flex flex-wrap gap-1.5">
              <span
                v-for="f in formats"
                :key="f"
                class="chip font-mono"
                :class="f === 'webp' ? 'chip-brand' : ''"
              >
                .{{ f }}
              </span>
            </div>
            <p>
              默认情况下，我们会返回 WebP 格式的图片，但是你可以通过向图片访问 URL
              拼接文件后缀的方式来访问特定格式的图片，完整的请求 URL 类似如下：
            </p>
            <code-block
              code="https://weavatar.com/avatar/ff3dcd55b299b96db5e2ed195af50817.png"
              lang="text"
              title="URL"
              class="not-prose my-4"
            />
            <p>如无必要，请保持使用默认的 WebP 格式，这是当下兼容性、速度、大小之间的最佳选择。</p>
          </section>

          <section id="params">
            <h2>额外的参数</h2>

            <section id="resize">
              <h3>调整头像大小</h3>
              <p>
                默认情况下，我们会返回 <code>80×80</code> 尺寸的头像，但是你可以通过
                <code>s</code> 或 <code>size</code> 参数来指定要获取的头像大小（支持 10 - 2000）。
              </p>
              <code-block
                code="https://weavatar.com/avatar/ff3dcd55b299b96db5e2ed195af50817?s=200"
                lang="text"
                title="URL"
                class="not-prose my-4"
              />
            </section>

            <section id="default">
              <h3>自定义默认头像</h3>
              <p>如果提供的哈希无法匹配到任何头像，则将会返回我们的 Logo 作为默认头像。</p>
              <div class="not-prose my-4">
                <avatar-image
                  :src="`${AVATAR_BASE}/?f=y&s=120`"
                  :size="60"
                  rounded="xl"
                  alt="默认头像"
                  class="ring-1 ring-line"
                />
              </div>
              <p>
                当然，你也可以通过 <code>d</code> 或
                <code>default</code> 参数指定需要返回的默认头像：
              </p>
              <code-block
                code="https://weavatar.com/avatar/ff3dcd55b299b96db5e2ed195af50817.jpg?d=你的URL"
                lang="text"
                title="URL"
                class="not-prose my-4"
              />
              <p>需要注意的是，传递的默认头像地址必须经过 URL 编码。</p>
              <p>
                除了允许你自己指定默认头像外，我们还准备了一组内置的默认头像，只需要传入
                <code>d=默认头像ID</code> 即可调用：
              </p>
              <div class="not-prose my-6 grid grid-cols-2 sm:grid-cols-3 gap-3">
                <div
                  v-for="item in defaultStyles"
                  :key="item.code"
                  class="surface p-4 flex items-start gap-3"
                >
                  <avatar-image
                    v-if="item.src"
                    :src="item.src"
                    :size="48"
                    rounded="xl"
                    :alt="item.name"
                    class="ring-1 ring-line"
                  />
                  <div
                    v-else
                    class="w-12 h-12 shrink-0 rounded-xl border border-dashed border-line-strong flex items-center justify-center text-fg3 text-xs font-mono"
                  >
                    404
                  </div>
                  <div class="min-w-0">
                    <code class="block text-xs text-brand break-all">{{ item.code }}</code>
                    <div class="mt-1 text-xs text-fg2 leading-snug">{{ item.desc }}</div>
                  </div>
                </div>
              </div>
            </section>

            <section id="forcedefault">
              <h3>强制加载默认头像</h3>
              <p>
                如果由于某种原因你想强制始终返回默认头像，可以使用 <code>f</code> 或
                <code>forcedefault</code> 参数并将其值设置为 <code>y</code>。
              </p>
            </section>

            <section id="rating">
              <h3>指定要显示的头像级别</h3>
              <p>为符合中国法律要求，<code>r</code> / <code>rating</code> 参数暂不提供支持。</p>
            </section>

            <section id="combination">
              <h3>组合参数</h3>
              <p>以上所介绍的所有参数都可以自由组合，比如你可以提供这样的一个头像 URL：</p>
              <code-block
                code="https://weavatar.com/avatar/ff3dcd55b299b96db5e2ed195af50817.png?d=initials&initials=WeAvatar&s=200&f=y"
                lang="text"
                title="URL"
                class="not-prose my-4"
              />
              <p>
                该头像 URL 将始终返回格式为 PNG、尺寸为 200 的字母头像，字母自动裁切为
                <code>WE</code>。
              </p>
            </section>
          </section>
        </article>

        <!-- 底部导航 -->
        <div
          class="mt-16 pt-8 border-t border-line flex flex-wrap items-center justify-between gap-4 text-sm"
        >
          <router-link :to="{ name: 'help' }" class="link inline-flex items-center gap-1">
            遇到问题？查看帮助
            <span class="i-lucide-arrow-right text-sm" />
          </router-link>
          <a
            :href="`${GITHUB_URL}/issues`"
            target="_blank"
            rel="noreferrer"
            class="text-fg2 hover:text-fg inline-flex items-center gap-1"
          >
            在 GitHub 反馈
            <span class="i-lucide-arrow-up-right text-sm" />
          </a>
        </div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref, watch } from 'vue'
import { useEventListener } from '@vueuse/core'
import CodeBlock from '@/components/ui/CodeBlock.vue'
import AvatarImage from '@/components/ui/AvatarImage.vue'
import DocCallout from '@/components/ui/DocCallout.vue'
import { AVATAR_BASE, GITHUB_URL } from '@/constants/links'

interface TocItem {
  id: string
  label: string
  children?: TocItem[]
}

const toc: TocItem[] = [
  { id: 'basic', label: '基本概念' },
  { id: 'hash', label: '邮箱 / 手机号的哈希' },
  {
    id: 'cms',
    label: '在 CMS 中使用',
    children: [
      { id: 'wordpress', label: 'WordPress' },
      { id: 'typecho', label: 'Typecho' },
      { id: 'emlog', label: 'Emlog' },
      { id: 'zblog', label: 'Z-Blog' }
    ]
  },
  {
    id: 'comments',
    label: '在评论系统中使用',
    children: [
      { id: 'twikoo', label: 'Twikoo' },
      { id: 'artalk', label: 'Artalk' }
    ]
  },
  { id: 'format', label: '指定图片格式' },
  {
    id: 'params',
    label: '额外的参数',
    children: [
      { id: 'resize', label: '调整头像大小' },
      { id: 'default', label: '自定义默认头像' },
      { id: 'forcedefault', label: '强制加载默认头像' },
      { id: 'rating', label: '指定头像级别' },
      { id: 'combination', label: '组合参数' }
    ]
  }
]

const chain = ['WeAvatar', 'Gravatar', '社交头像', '默认头像']
const formats = ['webp', 'jpg', 'jpeg', 'png', 'gif', 'tiff', 'heif', 'heic', 'avif', 'jxl']

const hashCode = `const raw = ' Hi@WeAvatar.com '
const hash = sha256(raw.trim().toLowerCase())
const url = \`https://weavatar.com/avatar/\${hash}\``

const defaultStyles = [
  { code: 'd=404', desc: '返回 404 错误', src: '' },
  {
    code: 'd=mp',
    desc: '简单的、卡通风格的人物剪影轮廓（不随哈希改变）',
    src: `${AVATAR_BASE}/?d=mp&f=y&s=96`
  },
  {
    code: 'd=identicon',
    desc: '基于哈希生成的几何图案',
    src: `${AVATAR_BASE}/?d=identicon&f=y&s=96`
  },
  {
    code: 'd=monsterid',
    desc: '基于哈希生成的怪物，有不同的颜色、面孔等',
    src: `${AVATAR_BASE}/?d=monsterid&f=y&s=96`
  },
  {
    code: 'd=wavatar',
    desc: '基于哈希生成的具有不同特征和背景的人脸',
    src: `${AVATAR_BASE}/?d=wavatar&f=y&s=96`
  },
  {
    code: 'd=retro',
    desc: '基于哈希生成的 8 位街机风格的像素化人脸',
    src: `${AVATAR_BASE}/?d=retro&f=y&s=96`
  },
  {
    code: 'd=robohash',
    desc: '基于哈希生成的机器人，有不同的颜色、面孔等',
    src: `${AVATAR_BASE}/?d=robohash&f=y&s=96`
  },
  {
    code: 'd=initials&initials=X',
    desc: '生成由 initials 指定的字母头像（最多保留 2 位）',
    src: `${AVATAR_BASE}/demo?d=initials&initials=WeAvatar&f=y&s=96`
  },
  {
    code: 'd=initials&name=X',
    desc: '生成由 name 指定的首字母头像（最多保留 1 位）',
    src: `${AVATAR_BASE}/demo?d=initials&name=WeAvatar&f=y&s=96`
  },
  {
    code: 'd=color',
    desc: '基于哈希生成的纯色头像，颜色随哈希改变',
    src: `${AVATAR_BASE}/?d=color&f=y&s=96`
  },
  { code: 'd=blank', desc: '生成一个透明的 PNG 图片', src: `${AVATAR_BASE}/?d=blank&f=y&s=96` }
].map((i) => ({ ...i, name: i.code }))

const wordpressCode = `if ( ! function_exists( 'get_weavatar_url' ) ) {
    /**
     * 替换 Gravatar 头像为 WeAvatar 头像
     *
     * WeAvatar 是新一代头像服务解决方案，可在 https://weavatar.com 修改头像
     */
    function get_weavatar_url( $url ) {
        $sources = array(
            'www.gravatar.com',
            '0.gravatar.com',
            '1.gravatar.com',
            '2.gravatar.com',
            'secure.gravatar.com',
            'cn.gravatar.com',
            'gravatar.com',
            'sdn.geekzu.org',
            'gravatar.duoshuo.com',
            'gravatar.loli.net',
            'cravatar.cn',
        );
        return str_replace( $sources, 'weavatar.com', $url );
    }
    add_filter( 'um_user_avatar_url_filter', 'get_weavatar_url', 1 );
    add_filter( 'bp_gravatar_url', 'get_weavatar_url', 1 );
    add_filter( 'get_avatar_url', 'get_weavatar_url', 1 );
    add_filter( 'um_user_avatar_url_filter', 'get_weavatar_url', PHP_INT_MAX );
    add_filter( 'bp_gravatar_url', 'get_weavatar_url', PHP_INT_MAX );
    add_filter( 'get_avatar_url', 'get_weavatar_url', PHP_INT_MAX );
}
if ( ! function_exists( 'set_defaults_for_weavatar' ) ) {
    /**
     * 替换 WordPress 讨论设置中的默认头像
     */
    function set_defaults_for_weavatar( $avatar_defaults ) {
        $avatar_defaults['gravatar_default'] = 'WeAvatar 头像';
        return $avatar_defaults;
    }
    add_filter( 'avatar_defaults', 'set_defaults_for_weavatar', 1 );
}
if ( ! function_exists( 'set_user_profile_picture_for_weavatar' ) ) {
    /**
     * 替换个人资料卡中的头像上传地址
     */
    function set_user_profile_picture_for_weavatar() {
        return '<a href="https://weavatar.com" target="_blank">您可以在 WeAvatar 修改您的资料图片</a>';
    }
    add_filter( 'user_profile_picture_description', 'set_user_profile_picture_for_weavatar', 1 );
}`

// 目录随滚动高亮
const active = ref('basic')
const indicatorStyle = ref({ top: '0px', height: '0px' })

const allIds = toc.flatMap((g) => [g.id, ...(g.children?.map((c) => c.id) ?? [])])
const parentOf = new Map<string, string>()
toc.forEach((g) => g.children?.forEach((c) => parentOf.set(c.id, g.id)))
const activeGroup = computed(() => parentOf.get(active.value) ?? active.value)

const updateIndicator = () => {
  const el = document.querySelector<HTMLElement>(`[data-toc="${active.value}"]`)
  if (!el) return
  indicatorStyle.value = { top: `${el.offsetTop}px`, height: `${el.offsetHeight}px` }
}

let ticking = false
const updateActive = () => {
  if (ticking) return
  ticking = true
  requestAnimationFrame(() => {
    ticking = false
    // 取视口上沿 120px 以内、最靠下的章节；滚到底部时高亮最后一节
    const threshold = 120
    let current = allIds[0]
    for (const id of allIds) {
      const el = document.getElementById(id)
      if (el && el.getBoundingClientRect().top <= threshold) current = id
    }
    const atBottom =
      window.innerHeight + window.scrollY >= document.documentElement.scrollHeight - 2
    active.value = atBottom ? allIds[allIds.length - 1] : current
  })
}

watch(active, () => nextTick(updateIndicator))
useEventListener(window, 'scroll', updateActive, { passive: true })
useEventListener(window, 'resize', updateActive, { passive: true })
onMounted(() => {
  updateIndicator()
  updateActive()
})
</script>

<style scoped>
.prose :deep(.not-prose) {
  font-size: initial;
  line-height: initial;
}

.prose :deep(.not-prose code) {
  background: none;
  border: 0;
  padding: 0;
}
</style>
