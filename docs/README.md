# docs

手写文档：

- `qq-hash.md`：QQ 哈希映射表的结构、构建、校验与部署

接口文档在运行时生成：开启 `http.docs` 后，`/openapi.json` 提供 OpenAPI 3.1 文档，`/docs` 提供浏览界面。

在路由表里给端点加上 `transport.Describe[Request, Response](status)` 即可进入文档，示例见 `internal/user/service/route.go`；没有 `Document` 的端点不会出现。
