# WeAvatar

WeAvatar 的后端服务：超越 Gravatar 的新一代头像服务，支持用户通过邮箱或手机号上传头像，
也能从 Gravatar、QQ 上获取头像，提供程序化头像生成、AI 自动化审核与多云 CDN 缓存刷新。

基于 [fiber-skeleton](https://github.com/libtnb/fiber-skeleton) 的模块化单体结构，
HTTP 框架为 [Fiber](https://gofiber.io/) v3，数据库为 PostgreSQL。

## 快速开始

需要 Go 1.27 与 PostgreSQL。

```bash
make init                    # 由 config/config.example.yml 生成 config/config.yml
# 编辑 config/config.yml：至少修改 app.key（32 字节）、http.domain、database.dsn
go run ./cmd/cli migrate     # 创建 / 升级表结构
make run                     # 或 make dev（air 热重载）
```

服务默认监听 `:3000`，接口前缀为 `/api`。`http.docs` 为 `true` 时，
`/openapi.json` 提供 OpenAPI 3.1 文档，`/docs` 提供在线文档页面。

## 配置

配置文件默认为 `config/config.yml`，可用环境变量 `APP_CONFIG` 指定其他路径，配置值不支持环境变量覆盖。
全部键见 [config/config.example.yml](config/config.example.yml)，启动时解析并校验。

部署在 nginx 后面时，把 `http.proxy_header` 设为 `X-Real-IP`，并在 nginx 中 `proxy_set_header X-Real-IP $remote_addr;`，
否则限流按代理地址计数，全站共用一份额度。服务会无条件采用该头，所以端口不能直接暴露到公网。

QQ 头像回退依赖 `hash.dir`（默认 `storage/hash/`）下的哈希映射表，文件缺失时仅关闭该回退，构建方法见 [docs/qq-hash.md](docs/qq-hash.md)。

## 目录结构

```
cmd/            入口：app（HTTP 服务）、cli（管理命令）、gen（代码生成器）
config/         配置文件
deploy/scripts/ 部署脚本：probe.sh（探活）、rollback.sh（回滚）
docs/           手写文档
internal/
  app/          组装根：把各模块组合成 app 与 cli 两个注入器
  migrations/   表结构迁移，每个迁移一个文件
  mocks/        mockery 生成的 mock
  platform/     基础设施：bootstrap（providers）、conf、server
  shared/       各模块共享的契约：transport、apperr、appinfo、registry、rule、job、database
  avatar/       头像：解析与回退、程序化生成、上传管理、AI 审核、CDN 刷新、hash 命令
  user/         用户：通行证登录、资料、注销账号
  verifycode/   短信 / 邮件验证码
  system/       CDN 用量统计、随机头像
pkg/            与业务无关的库：图片处理、MPHF、QQ 哈希表、CDN、审核、短信、邮件、OAuth、极验、队列等
storage/        运行时文件：日志、哈希表、上传的头像与缓存
web/            前端（Vue + TypeScript），独立的 pnpm 工程
```

## 架构

每个业务模块按 [Kratos](https://go-kratos.dev/) 的三层划分：

- **biz**：领域模型、仓库接口与用例，不含传输层与数据库代码
- **data**：仓库与外部客户端的实现
- **service**：传输层适配：绑定并校验请求，调用用例，组装响应

依赖注入使用 [libtnb/wire](https://github.com/libtnb/wire) 在编译期生成构造代码。
`internal/app/arch_test.go` 约束模块边界：模块之间只能通过对方的 `biz` 包交互，
且业务模块不能依赖 `app`、`platform`、`migrations`。

## 常用命令

| 命令 | 说明 |
|---|---|
| `make init` | 生成本地配置文件 |
| `make run` | 启动 HTTP 服务 |
| `make dev` | 热重载开发（需要 [air](https://github.com/air-verse/air)） |
| `make generate` | 重新生成 Wire 注入代码与 mock |
| `make gen name=article` | 生成一个 CRUD 模块骨架 |
| `make gen-migration name=add_xxx_to_users_table` | 生成一个迁移文件 |
| `make lint` | 运行 golangci-lint |
| `make test` | 带 race 检测与覆盖率运行测试 |
| `make build` | 构建静态二进制到 `bin/` |
| `make help` | 列出全部目标 |

迁移与哈希表命令：

```bash
go run ./cmd/cli migrate                    # 执行未应用的迁移
go run ./cmd/cli migrate status             # 查看已应用与待执行的迁移
go run ./cmd/cli migrate rollback --step 1  # 回滚最近一次迁移
go run ./cmd/cli hash build                 # 构建 QQ 哈希映射表
```

## 部署

服务以 systemd 单元运行，由 GitHub Actions 的 `Deploy Backend` 工作流（`.github/workflows/deploy-backend.yml`）手动触发部署：

1. 在 runner 上 `make build`，生成注入版本号的 `bin/app` 与 `bin/cli` 静态二进制，并计算 sha256。
2. 通过 SSH 与 rsync 上传为 `app.new`、`cli.new`，同时上传 `config/config.example.yml` 与 `deploy/scripts/*`。
3. 在服务器上校验 sha256，以 `www` 身份依次执行 `cli.new migrate status`、`migrate plan`、`migrate up`。
4. 备份当前二进制为 `*.prev`，切换新二进制，写入 `VERSION`，把本次写入的文件属主改为 `www`，不遍历 `storage/`。
5. `systemctl reload-or-restart` 让新二进制零停机接管（单元未运行时直接启动），等主进程 PID 切换后用 `probe.sh`
   探测 `/healthz` 与 `/readyz`；失败时 `rollback.sh` 换回上一版并同样重载。数据库迁移不会回滚，迁移必须保持向后兼容。

### 仓库配置

变量与密钥都配置在仓库级的 Actions 设置里，不使用部署环境。变量都可选。

| 类型 | 名称 | 说明 |
|---|---|---|
| 变量 | `API_DIR` | 安装目录，默认 `/opt/ace/projects/weavatar` |
| 变量 | `API_PORT` | 服务在本机监听的端口，默认 `3100` |
| 变量 | `SERVICE_NAME` | systemd 单元名，默认 `weavatar` |
| 密钥 | `DEPLOY_HOST` | SSH 主机 |
| 密钥 | `DEPLOY_USER` | SSH 用户：root，或能写入 `API_DIR` 且可免密 `sudo chown`、`sudo -u www`、`systemctl`、`journalctl` 的用户 |
| 密钥 | `DEPLOY_SSH_KEY` | 该用户的 SSH 私钥 |

### 服务器准备

服务器需要 `rsync`、`curl`、`sha256sum` 与 `www` 用户。`$API_DIR/storage` 存放日志、哈希表与头像，
由运维预先创建并把属主设为 `www`，部署不会修改它。首次部署前，把
`config/config.example.yml` 复制为 `$API_DIR/config/config.yml` 并填好，然后安装 systemd 单元，例如
`/etc/systemd/system/weavatar.service`：

```ini
[Unit]
Description=WeAvatar
After=network-online.target postgresql.service
Wants=network-online.target

[Service]
Type=notify-reload
NotifyAccess=all
ExitType=cgroup
User=www
Group=www
WorkingDirectory=/opt/ace/projects/weavatar
Environment=APP_CONFIG=/opt/ace/projects/weavatar/config/config.yml
ExecStart=/opt/ace/projects/weavatar/app
Restart=on-failure
TimeoutStopSec=40

[Install]
WantedBy=multi-user.target
```

`Type=notify-reload` 让 `systemctl reload` 成为零停机升级，`NotifyAccess=all` 与 `ExitType=cgroup`
是 [graceful](https://github.com/libtnb/graceful) 接管主进程的要求。

```bash
sudo systemctl daemon-reload && sudo systemctl enable weavatar
```

| 信号 | 行为 |
|---|---|
| SIGINT / SIGTERM | 等待请求与任务结束（最长 30 秒），再按逆序关闭资源 |
| SIGHUP | 零停机升级二进制 |

## 文档

- [docs/qq-hash.md](docs/qq-hash.md)：QQ 头像回退使用的哈希映射表
- [CLAUDE.md](CLAUDE.md)：项目结构与开发约定

## 许可证

[AGPLv3](LICENSE)
