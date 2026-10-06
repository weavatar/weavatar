# docs

这里存放手写文档：

- `qq-hash.md`：QQ 哈希映射表的结构、构建、校验与部署

接口文档在运行时生成：开启 `http.docs` 后，服务在 `/openapi.json` 提供 OpenAPI 3.1 文档，在 `/docs` 提供浏览界面。

要给端点写文档，在模块的路由表里为端点加上 `transport.Describe[Request, Response](status)`，示例见 `internal/user/service/route.go`。没有 `Document` 的端点不会出现在文档里。
