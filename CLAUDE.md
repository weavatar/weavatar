# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

WeAvatar 是一个头像服务（类似 Gravatar 的中国替代品），支持用户通过邮箱或手机号上传头像，同时提供 Gravatar/QQ 头像回退、程序化头像生成（identicon、monsterid、robohash、wavatar、retricon）、AI 内容审核和多云 CDN 缓存刷新。

后端是基于 [fiber-skeleton](https://github.com/libtnb/fiber-skeleton) 的模块化单体，单 Go module `github.com/weavatar/weavatar`（Go 1.27），数据库为 PostgreSQL。

## 结构

- `cmd/app`：HTTP 服务。`cmd/cli`：管理命令（`migrate`、`hash`）。`cmd/gen`：模块与迁移生成器。
- `internal/app`：组合根。`wire.go` 汇总所有模块，生成 app 与 cli 的注入器；`arch_test.go` 是架构测试。
- `internal/platform`：基础设施。`bootstrap` 是各 provider，`conf` 是配置，`server` 是 Fiber、中间件、健康检查与 OpenAPI。
- `internal/shared`：模块共用的契约。`transport` 负责绑定、响应、端点声明、登录与限流；`apperr` 是带类型的错误；`registry` 是 Wire 多绑定集合；`job` 是定时任务；`appinfo` 是注入给模块的配置值；`rule` 是自定义校验规则；`database` 把事务放进 ctx，让跨模块的写入加入同一事务。
- `internal/migrations`：数据库迁移。`internal/mocks`：mockery 生成物。
- `internal/<模块>`：业务模块，按 `biz/data/service` 分层，根上一个 `wire.go`。
  - `avatar`：核心模块。负责头像解析（WeAvatar → Gravatar → QQ → 默认头像）、程序化生成、头像增删改查、AI 审核队列、CDN 刷新、缓存清理定时任务和 `hash` 命令。
  - `user`：OAuth 登录、JWT 签发、用户资料、注销账号。
  - `verifycode`：短信与邮件验证码，含发送冷却与限流。
  - `system`：CDN 用量统计、随机头像。
- `pkg/`：与业务无关的库，包括图片处理、MPHF、QQ 哈希表、CDN、审核、短信、邮件、OAuth、极验、队列等。
- `config/config.example.yml`：全部配置项。`config/config.yml` 不入库。
- `docs/`：手写文档。`web/`：Vue 前端。

## 命令

仓库根执行：

| 命令 | 用途 |
| --- | --- |
| `make init` | 复制 `config/config.example.yml` 为 `config/config.yml` |
| `make run` / `make dev` | 启动服务（默认 `:3000`）/ air 热重载。启动时先执行未应用的迁移 |
| `make build` | 构建 `bin/app`、`bin/cli`（`CGO_ENABLED=0`，注入版本号） |
| `make test` | `go test -race`，输出 `coverage.out` |
| `make lint` | golangci-lint |
| `make generate` | 重新生成 Wire 注入代码与 mockery mock。改了任何 `wire.go` 或 biz 接口后必跑 |
| `make wire-check` | 检查 `wire_gen.go` 是否最新 |
| `make gen name=article` | 生成 CRUD 模块骨架。之后把 `article.Module` 加进 `internal/app/wire.go` 的 `Include`，再 `make generate` |
| `make gen-migration name=add_email_to_users_table` | 生成迁移文件 |
| `make gen-check` | 验证生成器产物能编译，改了 `cmd/gen` 模板后必跑 |

CLI：`go run ./cmd/cli migrate {up,plan,status,rollback --step N}`（不带子命令等同 `up`），以及 `hash {build,verify,lookup,stat}`。

配置：

- 配置值只从 YAML 文件读取，**不支持环境变量覆盖**。
- 唯一的环境变量是 `APP_CONFIG`，用于指定配置文件路径（默认 `config/config.yml`），主要供测试使用。
- 启动时解析为类型化的 `conf.Config` 并校验，例如 `app.key` 必须是 32 字节。
- `app.debug: true` 会跳过极验校验，生产环境必须关闭。
- 部署在 nginx 之后时，要把 `http.proxy_header` 设为 `X-Real-IP`，否则 `c.IP()` 是代理地址，所有经代理的请求共用一份限流额度。设置后该请求头来自任何连接都会被采用，所以服务只能经代理访问；值不是合法 IP 时回退到连接地址。

## 架构要点

- **分层**：biz 放模型、Repo 与端口接口、`XxxUsecase`，HTTP、CLI、定时任务共用同一个 usecase。data 放接口实现，包括 rio 仓库、文件存储、`pkg/` 客户端适配和对其他模块 usecase 的适配器。service 放 Fiber handler、`request.go`、`route.go`、CLI 命令与定时任务。
- **装配**：每个模块的 `wire.go` 用 libtnb/wire `Provide` 构造器，把路由、命令、定时任务 `Contribute` 到 `registry` 集合，只 `Export` 被其他模块使用的 usecase。
- **模块边界**：由 `internal/app/arch_test.go` 的 `TestModuleBoundaries` 强制。
  - 业务模块只能 import `internal/shared/*` 和其他模块的 `biz` 包，不能 import `app`、`platform`、`migrations`。
  - 模块内的 `biz` 不能 import 自己的 `data` 与 `service`。
  - `shared` 只能 import `shared`。`platform` 只能 import `shared`、`platform/conf` 和自己的子包。`platform/conf` 不 import 任何 internal 包。
  - `pkg/` 不受限制。
- **跨模块调用**：在自己的 biz 里声明端口接口，在 data 里适配对方的 usecase。例如 avatar 的 `Users` 端口适配 user 模块的 `UserUsecase`。
- **注销账号**：二次确认是重新走一次树新峰通行证 OAuth，state 带 `delete-` 前缀且缓存值为发起人的 userID，回调换到的 `UnionID` 必须与当前用户一致。各模块通过 `registry.UserCleanups` 贡献自己的清理（avatar 的 `data.NewUserCleanup` 删头像行、文件并刷新 CDN），user 模块的 `DeletionUsecase` 在一个事务里依次执行清理再软删 users 行（只打 `deleted_at`，不改其他列）。`Multibind[registry.UserCleanups]()` 只在 `internal/app/wire.go` 声明，user 模块自己不要声明，否则注入的是空集合；`DeletionUsecase` 独立于 `UserUsecase`，因为 avatar 依赖 `UserUsecase`，合在一起会成 wire 环。
- **配置注入**：模块不能 import `platform/conf`。需要的配置值通过 `internal/shared/appinfo` 的命名类型注入，例如 `appinfo.Domain`、`appinfo.CodeExpire`。
- **HTTP**：Fiber v3。路由表返回 `transport.Endpoints`，每个端点带 OpenAPI `Document`。需要登录的端点加 `Middlewares: {transport.MustLogin(jwt)}`，限流加 `transport.Throttle`。注意 `Throttle` 每次调用都是独立配额，几个端点要共享配额时得共用同一个 handler。handler 只做三件事：`transport.Bind`、调用 usecase、`transport.Success` 或 `transport.ErrorFrom`。
- **请求生命周期**：Fiber 未开 `Immutable`，从 `fiber.Ctx` 取到的值（`c.Query`、`c.Params`、`c.Get`、`c.IP`、`c.Body` 以及 `transport.Bind` 的结果）都引用请求缓冲区，只在 handler 内有效；需要在响应返回后使用的值（cache 键值、队列闭包、goroutine）在逃逸点显式 `strings.Clone`。`c.Context()` 不得越过 handler，goroutine 与后台任务用自己的 ctx，由 `arch_test.go` 的 `TestRequestContextDoesNotEscape` 检查。`adaptor.ConvertRequest` 的结果只能在请求内同步使用。
- **响应**：成功为 `{"msg":"success","data":...}`。失败为 `{"msg":"...","data":null}`，apperr 错误另带机器码 `code`（如 `avatar.not_square`）；框架级错误（404、405、413、panic）也走同一信封。登录态通过 `Authorization: Bearer <jwt>` 传递，未登录返回 401。前端按 HTTP 状态分流：422 弹 toast，401 清除登录态，其余弹对话框。
- **错误**：客户端可见的错误由 biz 的错误构造器返回，例如 `ErrStateExpired()`，它基于 `apperr`，再由 `transport.ErrorFrom` 映射为状态码。未命中统一透传 `rio.ErrNotFound`，映射为 404。没有 kind 的错误返回通用 500，详情只进日志。
- **持久化**：PostgreSQL + go-rio/rio。查询写成包级模板 `rio.From[T]().Where("...").Must()`，参数延迟绑定。仓库一律用 `database.Q(ctx, db)` 取执行目标，各模块的 `TxRunner` 端口由 `database.Runner` 实现，嵌套调用落成 savepoint。迁移用 go-rio/migrate，一个文件一个迁移。
- **校验**：libtnb/validator，标签用布尔 DSL，例如 `required && email && max:255`。`required` 开启了严格模式，会拒绝零值。自定义规则 `geetest`、`verify_code`、`cn_mobile`、`exists`、`not_exists` 在 `internal/shared/rule`。
- **路由**：所有 API 在 `/api` 前缀下，覆盖 `avatar`、`avatars`、`user`、`verify_code`、`system`。核心端点是 `GET /api/avatar/:hash`。另有探针 `/healthz`、`/readyz`，`/` 与 `/api` 302 跳转到 `https://<http.domain>`；`http.docs` 开启时，`/openapi.json` 与 `/docs` 提供接口文档。
- **其他技术栈**：配置 koanf，日志 log/slog + libtnb/logrotate，定时任务 libtnb/cron，生命周期 libtnb/graceful，错误 samber/oops，图片处理纯 Go、无 CGO。

## 编码约定

- 声明顺序为常量、变量、类型、构造器、公开函数、公开方法、私有方法、包级私有助手。简单助手内联。
- 测试文件同样按这个顺序：类型和它的导出方法、测试函数，最后是私有方法与助手。
- 错误 key 格式为 `<模块>.<原因>`（snake_case），用户可见文案用中文。
- 空列表返回 `[]` 而不是 `null`，时间一律 UTC。pgx 按进程本地时区返回 timestamptz，所以 `cmd/app` 与 `cmd/cli` 的 main 开头设置了 `time.Local = time.UTC`，日志时间也是 UTC。
- 代码和注释用英文，本文件与 `docs/` 用中文。

## 注释

注释尽可能少，一两句为限。只写代码说不出的东西，例如不显然的原因、外部约束和有意的取舍。不复述代码，不写分节横幅、注释掉的代码和修改记录，也不写编号、名字、日期或历史叙述。

## 测试纪律

- 用 `testing` + `libtnb/assert`。`must` 失败即停，`check` 继续执行，参数顺序为 `(got, want)`。
- 每条业务规则一条用例，覆盖正常路径与拒绝路径。不写压力或循环型测试，不用 `time.Sleep` 等待，不为可测性拆碎方法。
- `service/request_test.go` 用 `v.Check[T]()` 保证校验标签能编译。validator 要带上 `rule.Options(...)` 与 `validator.WithStrictRequired()`，与生产一致。
- service 层用 mockery mock 加 `app.Test` 测 handler。biz 层 mock 端口测 usecase。mock 的 func 字段留空时，意外调用会直接 panic。
- 集成测试只覆盖必须碰数据库的行为。

## QQ 哈希表

QQ 头像回退依赖 `hash.dir`（默认 `storage/hash/`）下的 MPHF 映射表，文件缺失时仅关闭该回退。详见 `docs/qq-hash.md`。

```bash
./cli hash build            # 构建，默认 10000 ~ 4294967295，md5 + sha256
./cli hash verify --full    # 全量校验
./cli hash lookup <hash>    # 调试查询
```

## 前端（`web/`）

独立的 Vue.js + TypeScript 前端，使用 Naive UI 组件库和 UnoCSS，采用 Composition API + `<script setup>` 风格。
