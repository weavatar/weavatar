<template>
  <div class="doc">
    <NCard title="文档">
      <NLayout has-sider>
        <NLayoutSider
          bordered
          collapse-mode="width"
          :collapsed-width="64"
          :collapsed="collapsed"
          show-trigger
          @collapse="() => (collapsed = !collapsed)"
          @expand="() => (collapsed = !collapsed)"
        >
          <NMenu
            :collapsed="collapsed"
            :collapsed-width="64"
            :collapsed-icon-size="22"
            :options="menuOptions"
            :on-update:value="onUpdate"
            :value="value"
          />
        </NLayoutSider>

        <div class="cont" v-if="value == 'basic'">
          <p>
            <NTag type="info">WeAvatar</NTag> 头像 API 可以像普通的图片 URL 一样请求，具体格式是:
            <NText code>https://weavatar.com/avatar/HASH</NText>
          </p>
          <p>
            其中 <NTag type="primary">HASH</NTag> 部分是 邮箱 / 手机号 的
            <NTag type="primary">SHA256</NTag> 或 <NTag type="primary">MD5</NTag> 哈希值，推荐使用
            <NTag type="primary">SHA256</NTag>，此电子邮箱 / 手机号须在
            <NText code>weavatar.com</NText> 上添加头像，否则会尝试返回
            <NTag type="warning">Gravatar</NTag> 头像和
            <NTag type="error">QQ</NTag>
            头像，如果都不存在，则返回默认头像
          </p>
        </div>
        <div class="cont" v-if="value == 'wordpress'">
          <p>
            安装启用
            <a target="_blank" href="https://wp-china-yes.com"> WP-China-Yes </a>
            插件并选择头像备用源，你可能还需要关闭主题、其他插件中自带的
            <NTag type="warning">Gravatar</NTag> 头像加速功能
          </p>
          <p>
            如果你不想安装插件，也可以通过添加以下代码到主题的
            <NTag type="error">functions.php</NTag>
            文件中来接入 <NTag type="info">WeAvatar</NTag>。
          </p>
          <NCode :code="code" language="php" word-wrap />
        </div>
        <div class="cont" v-if="value == 'typecho'">
          <p>
            添加
            <NCode
              code="define('__TYPECHO_GRAVATAR_PREFIX__', 'https://weavatar.com/avatar/');"
              language="php"
              inline
            />
            到站点根目录的
            <NTag type="error">config.inc.php</NTag>
            中来接入 <NTag type="info">WeAvatar</NTag>
          </p>
        </div>
        <div class="cont" v-if="value == 'emlog'">
          <p>
            <b>Pro之前(5.x 6.x民间): </b> 通过修改
            <NTag type="error">include/lib/function.base.php</NTag>
            中
            <NText code>getGravatar</NText>
            函数中的
            <NText code>http://www.gravatar.com</NText>
            为
            <NText code>https://weavatar.com</NText>
            接入 <NTag type="info">WeAvatar</NTag>，你可以参考左侧
            <b>额外的参数</b> 部分来修改默认头像
          </p>
          <p>
            <b>Pro之后(2.x): </b> 修改
            <NTag type="error">include/lib/common.php</NTag>
            中
            <NText code>getGravatar</NText>
            函数中的
            <NText code>cravatar.cn</NText>
            为
            <NText code>weavatar.com</NText>
            接入 <NTag type="info">WeAvatar</NTag>，你可以参考左侧
            <b>额外的参数</b> 部分来修改默认头像
          </p>
        </div>
        <div class="cont" v-if="value == 'zblog'">
          <p>
            后台应用中心搜索 <NTag type="info">WeAvatar</NTag> 安装插件或前往
            <a target="_blank" href="https://app.zblogcn.com/?id=38455">
              https://app.zblogcn.com/?id=38455
            </a>
            下载插件并手动安装
          </p>
        </div>
        <div class="cont" v-if="value == 'twikoo'">
          <n-alert type="success">
            Twikoo 已默认接入 WeAvatar，无需额外设置。
          </n-alert>
        </div>
        <div class="cont" v-if="value == 'artalk'">
          <n-alert type="success">
            Artalk 已默认接入 WeAvatar，无需额外设置。
          </n-alert>
        </div>
        <div class="cont" v-if="value == 'hash'">
          <p>1. 去除首位两边的空格</p>
          <p>2. 所有字母转小写</p>
          <p>3. 计算 SHA256 值</p>
        </div>
        <div class="cont" v-if="value == 'format'">
          <p>
            我们当前支持
            <NTag type="info">10</NTag>
            种图片返回格式，分别是:
            <NTag type="primary">webp</NTag> | <NTag type="primary">jpg</NTag> |
            <NTag type="primary">jpeg</NTag> | <NTag type="primary">png</NTag> |
            <NTag type="primary">gif</NTag> | <NTag type="primary">tiff</NTag> |
            <NTag type="primary">heif</NTag> | <NTag type="primary">heic</NTag> |
            <NTag type="primary">avif</NTag> | <NTag type="primary">jxl</NTag>。
          </p>
          <p>
            默认情况下，我们会返回
            <NTag type="info">WEBP</NTag> 格式的图片，但是你可以通过向图片访问
            <NTag type="primary">URL</NTag> 拼接文件后缀的方式来访问特定格式的图片，完整的请求
            <NTag type="primary">URL</NTag> 类似如下:
          </p>
          <NText code>https://weavatar.com/avatar/ff3dcd55b299b96db5e2ed195af50817.png</NText>
          <p>
            如无必要，请保持使用默认
            <NTag type="info">WEBP</NTag> 格式，这是当下兼容性、速度、大小之间的最佳选择。
          </p>
        </div>
        <div class="cont" v-if="value == 'resize'">
          <p>
            默认情况下，我们会返回 <NTag type="info">80×80</NTag> 尺寸的头像，但是你可以通过
            <NText code>s</NText> 或 <NText code>size</NText> 参数来指定要获取的头像大小（支持 10 -
            2000）
          </p>
        </div>
        <div class="cont" v-if="value == 'default'">
          <p>如果提供的哈希无法匹配到任何头像，则将会返回我们的 Logo 作为默认头像。</p>
          <p>
            <NImage
              src="https://weavatar.com/avatar/?f=y&s=60"
              alt="默认头像"
              width="60"
              height="60"
            />
          </p>
          <p>
            当然，你也可以通过
            <NText code>d</NText> 或 <NText code>default</NText> 参数指定需要返回的默认头像:
          </p>
          <NText code>
            https://weavatar.com/avatar/ff3dcd55b299b96db5e2ed195af50817.jpg?d=你的URL
          </NText>
          <p>需要注意的是，传递的默认头像地址必须经过 <NTag type="primary">URL</NTag> 编码</p>
          <p>
            除了允许你自己指定默认头像外，我们还准备了一组内置的默认头像，只需要传入
            <NText code>d=默认头像ID</NText> 即可调用:
          </p>
          <ul class="space-y-4 list-none">
            <li>
              <p><NText code>d=404</NText>: 返回 404 错误</p>
            </li>
            <li>
              <p>
                <NText code>d=mp</NText>: 简单的、卡通风格的人物剪影轮廓（不随哈希改变）
              </p>
            </li>
            <li>
              <p>
                <NText code>d=identicon</NText>: 基于哈希生成的几何图案
              </p>
            </li>
            <li>
              <p>
                <NText code>d=monsterid</NText>: 基于哈希生成的怪物，有不同的颜色、面孔等
              </p>
            </li>
            <li>
              <p>
                <NText code>d=wavatar</NText>: 基于哈希生成的具有不同特征和背景的人脸
              </p>
            </li>
            <li>
              <p>
                <NText code>d=retro</NText>: 基于哈希生成的 8 位街机风格的像素化人脸
              </p>
            </li>
            <li>
              <p>
                <NText code>d=robohash</NText>: 基于哈希生成的机器人，有不同的颜色、面孔等
              </p>
            </li>
            <li>
              <p>
                <NText code>d=initials&initials=X</NText>: 生成由 initials 指定的字母头像（最多保留 2 位）
              </p>
            </li>
            <li>
              <p>
                <NText code>d=initials&name=X</NText>: 生成由 name 指定的首字母头像（最多保留 1 位）
              </p>
            </li>
            <li>
              <p>
                <NText code>d=color</NText>: 基于哈希生成的纯色头像，颜色随哈希改变
              </p>
            </li>
            <li>
              <p>
                <NText code>d=blank</NText>: 生成一个透明的 PNG 图片
              </p>
            </li>
          </ul>
          <NSpace>
            <NImage
              src="https://weavatar.com/avatar/?d=mp&f=y&s=60"
              alt="人物轮廓"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/?d=identicon&f=y&s=60"
              alt="几何图案"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/?d=monsterid&f=y&s=60"
              alt="怪物"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/?d=wavatar&f=y&s=60"
              alt="人脸"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/?d=retro&f=y&s=60"
              alt="像素人脸"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/?d=robohash&f=y&s=60"
              alt="机器人"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/demo?d=initials&initials=WeAvatar&f=y&s=60"
              alt="字母头像"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/demo?d=initials&name=WeAvatar&f=y&s=60"
              alt="首字母头像"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/?d=color&f=y&s=60"
              alt="纯色头像"
              width="60"
              height="60"
            />
            <NImage
              src="https://weavatar.com/avatar/?d=blank&f=y&s=60"
              alt="透明图片"
              width="60"
              height="60"
            />
          </NSpace>
          <p>为方便演示，以上图片均通过 <NTag type="warning">CSS</NTag> 添加了边框</p>
        </div>
        <div class="cont" v-if="value == 'forcedefault'">
          <p>
            如果由于某种原因你想强制始终返回默认头像，您可以使用 <NText code>f</NText> 或
            <NText code>forcedefault</NText> 参数并将其值设置为 <NText code>y</NText>
          </p>
        </div>
        <div class="cont" v-if="value == 'rating'">
          <p>为符合中国法律要求，该参数暂不提供支持</p>
        </div>
        <div class="cont" v-if="value == 'combination'">
          <p>
            以上所介绍的所有关于
            <NTag type="info">WeAvatar</NTag>
            的参数都可以自由组合，比如你可以提供这样的一个头像
            <NTag type="primary">URL</NTag>:
          </p>
          <p>
            <NText code>
              https://weavatar.com/avatar/ff3dcd55b299b96db5e2ed195af50817.png?d=initials&initials=WeAvatar&s=200&f=y
            </NText>
          </p>
          <p>
            该头像 <NTag type="primary">URL</NTag>
            将始终返回格式为
            <NTag type="primary">PNG</NTag>、尺寸为
            <NTag type="primary">200</NTag>
            的字母头像，字母自动裁切为 <NTag type="primary">WE</NTag>
          </p>
        </div>
      </NLayout>
    </NCard>
  </div>
</template>

<script setup lang="ts">
import type { MenuOption } from 'naive-ui'
import {
  NAlert,
  NCard,
  NCode,
  NIcon,
  NImage,
  NLayout,
  NLayoutSider,
  NMenu,
  NSpace,
  NTag,
  NText
} from 'naive-ui'
import type { Component } from 'vue'
import { h, ref } from 'vue'
import {
  BrowsersOutline,
  CodeSlashSharp,
  CogOutline,
  CropOutline,
  ExtensionPuzzleOutline,
  EyeOutline,
  GridOutline,
  ImageOutline,
  LayersOutline,
  LogoWordpress,
  RocketOutline,
  ServerOutline,
  ChatbubblesOutline,
  ChatboxOutline
} from '@vicons/ionicons5'

const code = `
if ( ! function_exists( 'get_weavatar_url' ) ) {
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
}
`
const menuOptions: MenuOption[] = [
  {
    label: '基本概念',
    key: 'basic',
    icon: renderIcon(RocketOutline)
  },
  {
    label: '在 CMS 中使用',
    key: 'cms-use',
    icon: renderIcon(LayersOutline),
    children: [
      {
        label: 'WordPress',
        key: 'wordpress',
        icon: renderIcon(LogoWordpress)
      },
      {
        label: 'Typecho',
        key: 'typecho',
        icon: renderIcon(ServerOutline)
      },
      {
        label: 'Emlog',
        key: 'emlog',
        icon: renderIcon(ServerOutline)
      },
      {
        label: 'Z-Blog',
        key: 'zblog',
        icon: renderIcon(ServerOutline)
      }
    ]
  },
  {
    label: '在评论系统中使用',
    key: 'comment-use',
    icon: renderIcon(ChatbubblesOutline),
    children: [
      {
        label: 'Twikoo',
        key: 'twikoo',
        icon: renderIcon(ChatboxOutline)
      },
      {
        label: 'Artalk',
        key: 'artalk',
        icon: renderIcon(ChatboxOutline)
      },
    ]
  },
  {
    label: '邮箱 / 手机号的哈希',
    key: 'hash',
    icon: renderIcon(CodeSlashSharp)
  },
  {
    label: '指定图片格式',
    key: 'format',
    icon: renderIcon(ImageOutline)
  },
  {
    label: '额外的参数',
    key: 'additional',
    icon: renderIcon(ExtensionPuzzleOutline),
    children: [
      {
        label: '调整头像大小',
        key: 'resize',
        icon: renderIcon(CropOutline)
      },
      {
        label: '自定义默认头像',
        key: 'default',
        icon: renderIcon(BrowsersOutline)
      },
      {
        label: '强制加载默认头像',
        key: 'forcedefault',
        icon: renderIcon(CogOutline)
      },
      {
        label: '指定要显示的头像级别',
        key: 'rating',
        icon: renderIcon(EyeOutline)
      },
      {
        label: '组合参数',
        key: 'combination',
        icon: renderIcon(GridOutline)
      }
    ]
  }
]

const collapsed = ref(true)
let value = ref('basic')

const onUpdate = (key: string) => {
  value.value = key
}

function renderIcon(icon: Component) {
  return () => h(NIcon, null, { default: () => h(icon) })
}
</script>

<style scoped>
.doc {
  padding: 40px;
}

a {
  text-decoration: none;
  color: #409eff;
}

.n-image {
  border: 1px solid black;
}

@media screen and (max-width: 719px) {
  .doc {
    padding: 40px 0;
  }

  .show-menu {
    display: flex !important;
  }
}

.cont {
  padding: 0 20px 20px;
  width: 100%;
}

:deep(.n-layout-toggle-button) {
  top: 25px;
}

.n-layout-sider {
  z-index: 0;
}
</style>
