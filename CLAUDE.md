# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

WeAvatar 是一个头像服务（类似 Gravatar 的中国替代品），支持用户通过邮箱或手机号上传头像，同时提供 Gravatar/QQ 头像回退、程序化头像生成（identicon、monsterid、robohash、wavatar、retricon）、AI 内容审核和多云 CDN 缓存刷新。

## 常用命令

### 构建

```bash
# HTTP 服务
go build -o app ./cmd/app

# CLI 工具
go build -o cli ./cmd/cli
```

### 开发（热重载）

```bash
air  # 使用 .air.toml 配置，监听 :3000
```

### 测试

```bash
go test -v -coverprofile="coverage.out" ./...
```

构建和测试需要系统安装 libvips（macOS: `brew install vips`，Ubuntu: `apt-get install libvips-dev`），因为项目使用 CGO 进行图片处理。

### Lint

```bash
golangci-lint run --timeout=30m ./...
govulncheck ./...
```

### 依赖注入代码生成

修改 `cmd/app/wire.go` 或 `cmd/cli/wire.go` 后需重新生成：

```bash
wire ./cmd/app
wire ./cmd/cli
```

### 配置

首次运行需复制配置文件：`cp config/config.example.yml config/config.yml`

## 架构

项目采用分层架构，通过 Google Wire 进行编译期依赖注入。

### 分层结构

```
route (路由注册) → service (业务逻辑) → biz (实体/接口定义) ← data (数据访问/仓库实现)
```

- **`cmd/app`** — HTTP 服务入口，Wire 注入配置
- **`cmd/cli`** — CLI 工具入口（如批量生成 QQ 头像哈希映射）
- **`internal/bootstrap`** — 基础设施初始化（配置、数据库、HTTP 服务器、缓存、队列、定时任务、加密、校验器），统一通过 `ProviderSet` 暴露给 Wire
- **`internal/biz`** — 实体定义和 Repo 接口（`UserRepo`、`AvatarRepo`），不含具体实现
- **`internal/service`** — 业务逻辑层（`AvatarService`、`UserService`、`VerifyCodeService`、`SystemService`、`CliService`）
- **`internal/data`** — Repo 接口的 GORM 实现，包含头像图片生成逻辑
- **`internal/route`** — Fiber 路由注册（`http.go` 注册 HTTP 路由，`cli.go` 注册 CLI 命令）
- **`internal/http`** — 中间件（认证、限流）、请求 DTO、自定义校验规则
- **`internal/migration`** — 数据库迁移（gormigrate）
- **`internal/cronjob`** — 定时任务（如每小时更新过期头像）
- **`internal/queuejob`** — 异步队列任务（如头像审核）

### 外部集成包（`pkg/`）

- **`pkg/avatars`** — Gravatar 和 QQ 头像获取
- **`pkg/cdn`** — 多 CDN 缓存刷新（11+ 驱动：Cloudflare、华为云、又拍云、EdgeOne 等）
- **`pkg/audit`** — 内容审核（阿里云、腾讯 COS）
- **`pkg/sms`** — 短信发送（阿里云、腾讯云）
- **`pkg/mail`** — 邮件发送
- **`pkg/oauth`** — OAuth 认证
- **`pkg/queue`** — 内存任务队列
- **`pkg/geetest`** — 极验验证码

### 关键技术栈

- **HTTP**: Fiber v3
- **ORM**: GORM + MySQL
- **图片处理**: govips (libvips)
- **配置**: koanf (YAML)
- **DI**: Google Wire
- **日志**: log/slog

### CGO 注意事项

项目依赖 CGO（govips），部分文件有 `_nocgo.go` 变体（`internal/app/app_nocgo.go`、`internal/service/avatar_nocgo.go`、`internal/data/avatar_nocgo.go`），在无 CGO 环境下提供降级实现。

### API 路由结构

所有 API 在 `/api` 前缀下，主要端点：
- `GET /api/avatar/:hash` — 获取头像（核心端点）
- `/api/avatars` — 头像 CRUD（需登录）
- `/api/user` — 用户认证与管理
- `/api/verify_code` — 短信/邮件验证码
- `/api/system/count` — 系统统计

## 前端（`web/`）

独立的 Vue.js + TypeScript 前端，使用 Naive UI 组件库和 UnoCSS，采用 Composition API + `<script setup>` 风格。
