# WeAvatar

## 说明

这是 WeAvatar 的后端项目，使用 AGPLv3 协议开源。

WeAvatar 是超越 Gravatar 的新一代头像服务，不仅支持用户上传头像，也能从 Gravatar、QQ 上获取头像，同时支持 AI 自动化审核。

## 开发

需要 Go 1.27 与 PostgreSQL。

```bash
make init                   # 生成 config/config.yml，按需修改数据库连接等配置
go run ./cmd/cli migrate    # 创建表结构（make run 启动时也会自动执行）
make run                    # 启动 HTTP 服务，默认 :3000
```

常用命令：`make dev`（热重载）、`make test`、`make lint`、`make generate`（重新生成依赖注入代码与 mock）、`make build`（构建 `bin/app` 与 `bin/cli`）。`make help` 列出全部命令。

开启 `http.docs` 后，接口文档在 `/docs`。

## 部署

`.github/workflows/deploy-backend.yml` 通过 SSH 上传二进制、执行迁移并重启 systemd 服务，所需的变量与密钥见文件开头的注释。

服务器准备：

- 创建 `www` 用户与安装目录，放好 `config/config.yml`，systemd 服务以 `www` 运行。
- 由运维创建安装目录下的 `storage/` 并把属主设为 `www`。部署不会创建或遍历 `storage/`，迁移命令以 `www` 身份运行，日志也归 `www` 所有。
- `DEPLOY_USER` 用 root，或者能免密执行 `sudo -u www`、`sudo chown`、`systemctl` 与 `journalctl` 的用户。
- 部署在 nginx 之后时，在配置里把 `http.proxy_header` 设为 `X-Real-IP`，并确保服务端口只对 nginx 开放。

## 文档

- [docs/qq-hash.md](docs/qq-hash.md)：QQ 头像回退使用的哈希映射表
- [CLAUDE.md](CLAUDE.md)：项目结构与开发约定
