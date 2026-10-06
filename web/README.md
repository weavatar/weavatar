# WeAvatar 前端

WeAvatar 的 Web 前端，基于 Vue 3、Vite、UnoCSS 与 Naive UI。

## 开发

```sh
pnpm install
cp .env.example .env   # 设置 VITE_API_URL，例如 https://weavatar.com/api
pnpm dev
```

## 构建与检查

```sh
pnpm build            # 类型检查 + 生产构建
pnpm lint             # ESLint
pnpm type-check       # vue-tsc
pnpm test:unit --run  # Vitest
```

## 目录

```
src/api          接口封装（auth、avatar、system、user、verifyCode）
src/components   组件（ui、avatar、captcha、home、layout）
src/composables  useTheme、useGeetest
src/constants    站点信息与外部链接
src/stores       Pinia 用户状态（pinia-plugin-persistedstate 持久化到 localStorage）
src/views        路由页面：首页、登录与回调、头像管理、我的资料、文档、帮助、关于、隐私政策
```
