# shared

所有模块共用的契约：

- `transport`：请求绑定、响应信封、端点声明、登录与限流中间件
- `apperr`：带类型的应用错误
- `registry`：Wire 多绑定集合（路由、命令、定时任务、健康检查、注销清理）
- `database`：经 ctx 传递事务，各模块的数据适配器加入调用方的事务
- `job`：定时任务的贡献类型
- `appinfo`：注入给模块的命名配置值（域名、哈希表目录、Gravatar 地址、验证码有效期、OAuth 客户端）
- `rule`：自定义校验规则（exists、not_exists、geetest、verify_code、cn_mobile）

这里的包只能互相依赖，由架构测试强制。
