# cmd

每个子目录一个二进制：

- `app`：HTTP 服务
- `cli`：管理命令（`migrate`、`hash` 等）
- `gen`：生成 CRUD 模块与迁移文件

入口保持精简：调用 Wire 生成的初始化函数后运行。
