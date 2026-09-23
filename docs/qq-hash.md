# QQ 哈希映射表

QQ 头像回退需要把请求中的邮箱哈希反查成 QQ 号。QQ 号的取值范围是已知且可枚举的，
所以映射表用**最小完美哈希函数（MPHF）+ 值数组**实现，不存储哈希本身，也不依赖数据库。

## 结构

每种哈希类型两个文件，位于 `hash.dir`（默认 `storage/hash/`）：

| 文件 | 内容 | 40 亿键时体积 |
|---|---|---|
| `qq_md5.idx` / `qq_sha256.idx` | 文件头 + 分区表 + 256 个分区各自的 MPHF | ~1.9 GB |
| `qq_md5.val` / `qq_sha256.val` | `uint32[n]`，全局槽位 → QQ 号 | 16 GB |

- 键是**完整摘要**（MD5 16 字节、SHA256 32 字节），不截断；MPHF 体积只与键数有关。
- 分区号取摘要首字节，与哈希的前两位十六进制对应。
- 查询流程：hex 解码 → 分区 → MPHF 求槽位 → 读值数组得到 QQ 号 → **重新计算 `{qq}@qq.com` 的摘要与请求比对**。
  MPHF 对表外的键也会给出槽位，这一步校验保证不会误命中。
- 文件通过 mmap 加载：idx 常驻页缓存（`MADV_WILLNEED`），val 按需读取（`MADV_RANDOM`），
  多进程（`http.prefork`）共享同一份页缓存，内存不进 Go 堆。
- 文件缺失时应用只记录警告，QQ 头像回退失效，不影响其它功能。

## 构建

```bash
# 默认构建 md5 + sha256，QQ 号 10000 ~ 4000000000，输出到配置的 hash.dir
./cli hash build

# 常用参数
./cli hash build --dir /data/hash --type md5 --end 4000000000 --workers 16
```

流程：枚举全部 QQ 号并按分区写入桶文件（每种类型 16 GB 中间文件，位于 `hash.dir/tmp`）
→ 逐分区重算摘要、构建 MPHF、写入值数组 → 原子替换目标文件。
16 核机器上两种类型合计约十几分钟；`--workers` 越大内存占用越高（SHA256 每个 worker 约 700 MB）。

构建机需要的磁盘：中间文件 32 GB + 产物 36 GB。**不要直接在线上目录构建**，
构建完成后把四个文件放到 `hash.dir` 再重启应用。

## 校验

```bash
./cli hash verify              # 随机抽样 10 万次查询
./cli hash verify --full       # 按槽位全量扫描 + 分区校验和 + 覆盖检查（需要约 500 MB 内存）
./cli hash lookup <hash>       # 查询单个哈希
./cli hash stat                # 查看文件信息
```

每次构建后都应跑一次 `--full`。由于查询时会用完整摘要校验，文件损坏的后果只会是"查不到"，不会是"查错人"。

## 部署

- `hash.dir` 建议挂载独立的 volume，`.dockerignore` 已排除 `storage/hash`，产物不会被打进镜像。
- 页缓存越大命中越快：idx 全部常驻需要约 4 GB，val 全部常驻再加 32 GB；
  内存不足时每次查询最多一次 SSD 随机读。
- QQ 号范围扩大时需要全量重建，MPHF 不支持增量。

## 代码

- `pkg/mphf`：通用 BBHash 风格 MPHF，构建、序列化、零拷贝加载、查询。
- `pkg/qqhash`：文件格式、构建流水线、mmap 表与查询、校验。
- `internal/bootstrap/qqhash.go`：启动时加载，`internal/data/avatar.go` 的 `GetQqByHash` 使用。
