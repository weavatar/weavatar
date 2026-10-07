# CLAUDE.md

WeAvatar 后端：类似 Gravatar 的头像服务，用户通过邮箱或手机号上传头像，另提供 Gravatar 与 QQ 头像回退、程序化头像生成（identicon、monsterid、robohash、wavatar、retricon）、AI 内容审核和多云 CDN 缓存刷新。品牌与用户可见文案一律写"WeAvatar"。

## 结构

单 Go module `github.com/weavatar/weavatar`（Go 1.27），基于 [fiber-skeleton](https://github.com/libtnb/fiber-skeleton) 的模块化单体。

- `cmd/app`：HTTP 服务；`cmd/cli`：管理命令（`migrate`、`hash`）；`cmd/gen`：CRUD 模块与迁移骨架生成器，模板在 `cmd/gen/templates`。
- `internal/app`：组合根。`wire.go` 装配所有模块，`arch_test.go` 守模块边界，`injector_test.go` 构建 app 与 cli 两张图并取 `/openapi.json`。
- `internal/platform`：`bootstrap`（日志、数据库、迁移、校验器、缓存、队列、QQ 哈希表、JWT、各外部客户端）、`conf`（类型化配置）、`server`（路由注册、全局中间件、探针、OpenAPI）。
- `internal/shared`：业务中立的契约：`transport`（Bind、响应信封、Endpoint、MustLogin、Throttle、分页）、`apperr`、`appinfo`（配置值的命名类型）、`registry`、`rule`（自定义校验规则）、`job`、`database`（ctx 事务）。
- 业务模块 `internal/<module>/{biz,data,service}` 加根上一个 `wire.go`：`avatar`（核心：头像解析 WeAvatar → Gravatar → QQ → 默认头像、程序化生成、头像增删改查、AI 审核队列、CDN 刷新、缓存清理定时任务、`hash` 命令）、`user`（通行证 OAuth 登录、JWT 签发、用户资料、注销账号）、`verifycode`（短信与邮件验证码）、`system`（CDN 用量统计、随机头像）。
- `internal/migrations`：PostgreSQL 迁移，一个迁移一个文件。`internal/mocks`：mockery 生成物。
- `pkg/`：与业务无关的库：图片处理、MPHF、QQ 哈希表、CDN、审核、短信、邮件、OAuth、极验、队列等。
- `config/config.example.yml`：全部配置项；`config/config.yml` 不入库。
- `docs/`：手写文档。
- `web/`：Vue + TypeScript 前端，独立的 pnpm 工程，CI 为 `web.yml`，部署为 `deploy-frontend.yml`，约定见「前端约定」。

## 命令

| 命令 | 用途 |
| --- | --- |
| `make init` | 由 `config.example.yml` 生成 `config/config.yml` |
| `make run` / `make dev` | 启动服务（默认 `:3000`）/ air 热重载 |
| `make generate` | 重生成 wire 与 mock（`go tool wire generate ./...`、`go tool mockery`），改了 `wire.go` 或 biz 接口后必跑 |
| `make wire-check` | 校验生成的装配代码是最新的 |
| `make gen name=article` / `make gen-migration name=…` | 生成模块骨架 / 迁移文件 |
| `make gen-check` | 在临时改动里生成一个模块并编译，改模板后必跑 |
| `make lint` / `make test` / `make build` | golangci-lint / `go test -race` / `bin/app`、`bin/cli` |

迁移：`go run ./cmd/cli migrate [up|plan|status|rollback --step N]`。QQ 哈希表：`go run ./cmd/cli hash {build,verify,lookup,stat}`，见下文「QQ 头像回退」。配置文件由 `APP_CONFIG` 指定，默认 `config/config.yml`；配置值不支持环境变量覆盖。启动时校验配置（如 `app.key` 必须 32 字节），并先执行未应用的迁移。`app.debug: true` 会跳过极验校验，生产必须关闭。

开发中只跑自己改动的包（`go build/vet/test -race ./internal/<module>/...`），全仓门禁交付前跑一次。

## 架构要点

- **HTTP**：Fiber v3，接口前缀 `/api`。各模块 `service/route.go` 返回 `transport.Endpoints`，中间件放 `Middlewares`（登录用 `transport.MustLogin(jwt)`，限流用 `transport.Throttle`），`Document` 用 `transport.Describe[Req, transport.Envelope[Resp]]`，没有请求参数时 `Req` 写 `openapi.NoBody`；不写 `Document` 的端点（如 `GET /api/avatar/:hash`）不进文档。核心端点是 `GET /api/avatar/:hash`，直接返回图片。另有探针 `/healthz`、`/readyz`，实时监控页 `/api/monitor`（Fiber contrib monitor，无鉴权），`/` 与 `/api` 302 跳转到 `https://<http.domain>`；`http.docs` 打开时提供 `/openapi.json` 与 `/docs`。
- **限流与客户端 IP**：`transport.Throttle` 按 `c.IP()` 计数，每次调用一份独立额度，多个端点共用就复用同一个实例（如短信与邮件验证码共用 5/min）。部署在 nginx 后必须把 `http.proxy_header` 设为 `X-Real-IP`，否则全站共用一份额度；设置后任何来源的该头都会被采用（不是合法 IP 时退回连接地址），所以服务只能经 nginx 访问。
- **时间**：一律 UTC。`cmd/app` 与 `cmd/cli` 启动时设 `time.Local = time.UTC`，驱动读回的 timestamptz 与日志都按 UTC 呈现。
- **接口契约**：前端（`web/`）依赖现有的路径与响应形状，改动前先确认兼容；前端路由 `/oauth/callback` 被后端拼进 OAuth 回调地址，登录与注销确认共用它。
- **响应**：成功 `{"msg":"success","data":…}`，`data` 始终输出；失败 `{"msg":"…","data":null}`，apperr 错误另带 `code`（错误 key，如 `avatar.not_square`），框架级错误（404、405、413、panic）也走同一信封；两者共用 `transport.Envelope`。前端失败一律弹 toast，401 另外清除登录态。
- **handler**：只做 `transport.Bind[Req](c, validate)` → usecase → `transport.Success` / `transport.ErrorFrom`。Bind 依次绑定 query、body、路径参数，后者覆盖前者，失败返回 422。`ErrorFrom` 把 `rio.ErrNotFound` 映射为 404，apperr 按 Kind 映射（Invalid 400、Unauthorized 401、Forbidden 403、NotFound 404、Conflict 409、Unprocessable 422），其他错误记日志后返回 500 通用文案。
- **请求生命周期**：Fiber 未开 `Immutable`，从 `fiber.Ctx` 取到的值（`c.Query`、`c.Params`、`c.Get`、`c.IP`、`c.Body` 以及 `transport.Bind` 的结果）都引用请求缓冲区，只在 handler 内有效；需要在响应返回后使用的值（cache 键值、队列闭包、goroutine）在逃逸点显式 `strings.Clone`。`c.Context()` 不得越过 handler，goroutine 与后台任务用自己的 ctx，由 `arch_test.go` 的 `TestRequestContextDoesNotEscape` 检查。`adaptor.ConvertRequest` 的结果只能在请求内同步使用。
- **校验**：`libtnb/validator` 布尔 DSL（`required && email && max:255`），已开 `WithStrictRequired`（`required` 拒绝 `""`、`0`、`false`）。自定义规则在 `internal/shared/rule`：`exists`、`not_exists`、`geetest`、`verify_code`、`cn_mobile`。
- **认证**：登录走通行证的 OAuth（`oauth.base_url`、`oauth.client_id`），回调换到身份后签发 JWT（`Authorization: Bearer`），未登录 401 `未登录`、过期 401 `登录已过期`；`transport.UserID(c)` 取用户 ID。
- **注销账号**：`GET /api/user/deletion/login` 返回重新授权地址，回调后 `POST /api/user/deletion/confirm {code, state}` 确认，授权到的身份必须是当前账号。`DeletionUsecase` 在一个事务里跑 `registry.UserCleanups`（avatar 模块删除该用户的头像与图片）并软删用户，任一步失败整体回滚。`DeletionUsecase` 独立于 `UserUsecase`，因为清理来自依赖 `UserUsecase` 的模块。
- **事务**：data 层一律用 `database.Q(ctx, r.db)` 取执行目标，ctx 带着事务时自动加入调用方的事务（`internal/shared/database`）。需要原子性的写操作放在 biz 的 `Transactor.Run` 内完成（由 `*database.Runner` 满足，嵌套调用变成 savepoint），例如 avatar 上传在同一事务里写库和落盘，注销时的清理加入 user 模块开的事务。
- **配置投影**：业务模块不能 import `platform/conf`，需要的配置值通过 `appinfo` 的命名类型注入（`appinfo.Domain`、`appinfo.CodeExpire`、`appinfo.OAuthClient`）；wire 按精确类型取依赖，同类型的不同值必须各有命名类型。
- **QQ 头像回退**：依赖 `hash.dir`（默认 `storage/hash/`）下的 MPHF 映射表，文件缺失时仅关闭该回退。`hash build` 默认覆盖 10000 ~ 4294967295、md5 + sha256，`hash verify --full` 全量校验，`hash lookup <hash>` 调试查询，详见 `docs/qq-hash.md`。

## 模块依赖方向

`verifycode`、`user` 无依赖；`avatar → user`；`system → avatar`。avatar 在自己的 `wire.go` 里向 `registry.UserCleanups` 贡献注销清理（`Multibind`、`Contribute`、`Export`），由 `internal/app/wire.go` 合并后注入 user 的 `DeletionUsecase`；user 模块不能自己 `Multibind` 这个集合，否则拿到的是空集合。

- 跨模块只通过对方 `biz` 包导出的 usecase：在自己的 `biz` 里定义端口接口（如 `Users`），在自己的 `data` 里写适配器调用对方 usecase。不 import 对方的 `data`、`service`，不读写对方的表。
- 模块只 Export 被别的模块用到的 usecase（`*userbiz.UserUsecase`、`*avatarbiz.AvatarUsecase`）和它贡献的 `registry` 集合。
- 业务模块只能 import `internal/shared/*`、其他模块的 `biz` 与 `pkg/*`；`biz` 不能 import 自己的 `data`、`service`；`shared` 只依赖 `shared`；`platform` 只能 import `shared`、`platform/conf` 和自己的子包，`platform/conf` 不 import 任何 internal 包。`arch_test.go` 强制这些规则，没有豁免。

## 数据库约定

- PostgreSQL + `go-rio/rio` + `go-rio/migrate`。模型放在拥有它的模块的 `biz`；表名是结构体名的复数 snake_case（`AppAvatar → app_avatars`），列名是字段名的 snake_case。
- 迁移文件名 `YYYYMMDDHHMMSS_<动作>_<表>_table.go`，在 `init()` 里 `collection.Add(…, up, migrate.WithDown(down))`。上线后改表只追加新迁移，不改已应用的迁移。
- 字符串用 `String(name, n)`（VARCHAR，不用 Char）；时间列用 `Timestamps()`、`SoftDeletes()`、`TimestampTz`，前两者紧跟主键列声明（PostgreSQL 加列只能追加到末尾，这样各表时间列的位置一致）；带软删除的表上的唯一索引加 `Where("deleted_at IS NULL")`，让软删行释放键。
- 查询写成包级模板：`rio.From[T]().Where("user_id = ?").Must()`，参数在执行时绑定。未命中透传 `rio.ErrNotFound`（由 `ErrorFrom` 变 404）。data 层错误用 `oops.In("<module>").Wrapf(err, …)` 包装；avatar 的 `wrap` 不包装 `rio.ErrNotFound`，因为未命中是头像服务的热路径。
- 空列表返回 `[]` 而不是 `null`。

## 编码约定

- gofmt；声明顺序为常量、变量、类型、构造器、公开函数、公开方法、私有方法、包级私有助手。简单助手内联，不为复用一两行抽函数。
- 构造器返回端口接口（`func NewAvatarRepo(db *rio.DB) biz.AvatarRepo`），wire 直接 `Provide`。
- 客户端可见错误用 `apperr.Invalid/Unauthorized/Forbidden/NotFound/Conflict/Unprocessable("<module>.<reason>", "中文文案").In("<module>").Errorf/Wrap/New(…)`，就地构造，多处复用的收进同文件末尾的私有 `errXxx()`；key 为 snake_case，文案面向用户、不带内部细节。外部服务的原始错误只进日志（`slog.WarnContext(ctx, "…", slog.Any("err", err))`）。
- Go 代码、注释与日志用英文，`web/` 的注释用中文；文档与本文件用中文。

## 注释

注释尽量少，一两句为限，只写代码说不出的东西：不显然的原因、外部约束、并发与幂等策略、有意的取舍。导出标识符的 doc comment 一句话说它是什么或做什么。不复述代码，不写分节横幅、注释掉的代码、编号、人名、日期和修改历史。

## 测试与提交

- 测试与源码同目录，用 `testing` 和 `libtnb/assert/{must,check}`（参数顺序 `got, want`），不引入 testify。测试名是描述行为的句子（`TestSendRefusesAgainWithinCooldown`），不用 `TestType_Method`。测试文件同样按上面的声明顺序，构造器与助手放在测试函数之后。
- mock 由 mockery v3 matryer 模板生成到 `internal/mocks/<module>/biz/`（`go tool mockery`，配置见 `.mockery.yaml`），用法 `repo.FindFunc = …`、`repo.FindCalls()`；不设的 Func 被调用会 panic，用来断言"不应调用"。
- 每个模块：`service/request_test.go` 用 `validator.WithStrictRequired()` 加 `rule.Options(nil, nil, nil, true)` 对每个请求结构 `v.Check[T]()`；`biz` 用例单测每条业务规则一条（正常路径 + 拒绝路径）；`service` handler 测试由 `harness`（`newHarness`、`mount`、`newValidator`）挂真实的路由表（含 `MustLogin` 与真实 JWT），覆盖校验失败、鉴权与响应形状。
- 不写压力、循环、`time.Sleep` 型测试；不为了可测性把直观的方法拆碎。需要数据库的行为才写集成测试：带 `//go:build integration`，从 `TEST_DATABASE_URL` 取 PostgreSQL（未设置时跳过），每个测试在独立 schema 里跑全部迁移。
- 提交信息用 Conventional Commits（`feat|fix|chore|docs|test|refactor(scope): …`），描述用中文，只暂存自己的改动。

## 前端约定

`web/` 是 Vue 3 + TypeScript 工程（Naive UI、UnoCSS、Composition API + `<script setup>`），用 pnpm 管理。

- **目录**：`api/` 用具名函数导出，查询叫 `fetchX`，操作用动词；`views/` 只放路由页面与布局，其余组件按领域放 `components/<领域>`。
- **请求**：一律经 `@/utils/http` 的 `get`/`post`/`put`/`del`，失败时 reject `ApiError`（网络错误 `status` 为 0），401 清除登录态，传 `{ noAlert: true }` 关闭自动提示。WeAvatar 的 alova `Method` 要 `await` 才发出请求。
- **写法**：用 async/await + try/catch/finally；首屏数据用 `load()`，回调页用 `run()`；操作中的状态用动名词（`submitting`、`saving`），数据加载用 `loading`。
- **弹窗与极验**：弹窗 `preset="card"`，`bordered`、`mask-closable`、`auto-focus` 都为 `false`，表单弹窗宽 440；极验一律 `await verify()`。
- **状态与路由**：store 用 pinia-plugin-persistedstate 持久化并提供 `isLogin`；路由 `meta` 只有 `title`、`requiresAuth`、`guest`，回跳用 `safeRedirect`。
- **样式**：品牌 token 只在 `src/styles/main.css` 与 `uno.config.ts`，组件只用主题色和 shortcut，不用 Sass。
- **门禁**：`pnpm lint`、`pnpm type-check`、`pnpm test:unit --run`、`pnpm build-only`。
