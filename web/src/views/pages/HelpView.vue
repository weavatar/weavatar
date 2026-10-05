<template>
  <div>
    <page-header title="帮助" desc="常见问题解答。" />

    <div class="wrap py-12 sm:py-16 grid gap-12 lg:grid-cols-[minmax(0,1fr)_18rem] lg:gap-20">
      <div>
        <faq-item question="WeAvatar 是什么？" default-open>
          WeAvatar 完美兼容所有 Gravatar 头像 API 接口。如果你未在 WeAvatar 设置头像，则会先尝试调用
          Gravatar 上的头像数据，其后是社交头像，最后会返回我们为你准备的一组默认头像。
        </faq-item>
        <faq-item question="为什么我的头像是红色的 WeAvatar 标志？">
          为了符合有关法律规定要求，WeAvatar 上的全部头像均会进行 AI 检测，您的头像被 AI
          判定违规。如果您认为这是误报，请与我们联系。
        </faq-item>
        <faq-item question="WeAvatar 支持哪些额外功能？">
          WeAvatar 目前支持手机号调用头像，方式同邮箱。计划在下个大版本更新为不同 App
          设置自定义头像的功能。
        </faq-item>
        <faq-item question="修改头像后多久生效？">
          添加、修改或删除头像后，全网缓存会在 3
          小时内刷新。如果本地仍显示旧头像，请清除浏览器缓存。
        </faq-item>
        <faq-item question="这里没有我的问题，需要寻求额外帮助">
          请加群交流，项目 QQ 群：
          <a :href="QQ_GROUP_URL" target="_blank" rel="noreferrer" class="link">{{
            QQ_GROUP_ID
          }}</a>
          。你也可以在
          <a :href="`${GITHUB_URL}/issues`" target="_blank" rel="noreferrer" class="link">GitHub</a>
          提交 Issue。
        </faq-item>
      </div>

      <aside class="space-y-4">
        <div class="surface p-6">
          <div class="w-10 h-10 rounded-xl bg-brand/10 text-brand flex items-center justify-center">
            <span class="i-lucide-book-open text-xl" />
          </div>
          <h3 class="mt-4 font-700 text-fg">接入文档</h3>
          <p class="mt-1.5 text-sm text-fg2 leading-relaxed">
            WordPress、Typecho、Emlog、Z-Blog 以及评论系统的接入方式。
          </p>
          <router-link
            :to="{ name: 'doc' }"
            class="mt-4 link text-sm font-600 inline-flex items-center gap-1"
          >
            查看文档
            <span class="i-lucide-arrow-right text-sm" />
          </router-link>
        </div>
        <div class="surface p-6">
          <div class="w-10 h-10 rounded-xl bg-brand/10 text-brand flex items-center justify-center">
            <span class="i-lucide-message-circle text-xl" />
          </div>
          <h3 class="mt-4 font-700 text-fg">QQ 交流群</h3>
          <p class="mt-1.5 text-sm text-fg2 leading-relaxed">
            群号 {{ QQ_GROUP_ID }}，问题反馈与商务合作。
          </p>
          <a
            :href="QQ_GROUP_URL"
            target="_blank"
            rel="noreferrer"
            class="mt-4 link text-sm font-600 inline-flex items-center gap-1"
          >
            加入群聊
            <span class="i-lucide-arrow-up-right text-sm" />
          </a>
        </div>
      </aside>
    </div>
  </div>
</template>

<script setup lang="ts">
import PageHeader from '@/components/ui/PageHeader.vue'
import FaqItem from '@/components/ui/FaqItem.vue'
import { GITHUB_URL, QQ_GROUP_ID, QQ_GROUP_URL } from '@/constants/links'
</script>
