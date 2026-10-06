# platform

基础设施装配，对业务模块不可见：

- `bootstrap`：日志、数据库、迁移、定时任务、校验器、缓存、队列、QQ 哈希表、JWT，以及外部客户端（CDN、审核、短信、邮件、OAuth、极验）的 provider
- `conf`：配置加载与校验
- `server`：HTTP 服务、路由、中间件、健康检查

只有 `internal/app` 可以导入这些包，由架构测试强制。
