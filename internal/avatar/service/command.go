package service

import (
	"context"
	"errors"
	"fmt"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/urfave/cli/v3"

	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/pkg/mphf"
	"github.com/weavatar/weavatar/pkg/qqhash"
)

const defaultHashDir = "storage/hash"

type hashCommand struct {
	dir string // hash.dir; the --dir flag wins
}

// HashCommand contributes `cli hash ...`, which builds and inspects the QQ
// hash tables; it needs no database.
func HashCommand(dir appinfo.HashDir) *cli.Command {
	h := hashCommand{dir: string(dir)}

	return &cli.Command{
		Name:  "hash",
		Usage: "QQ 哈希表操作",
		Commands: []*cli.Command{
			{
				Name:   "build",
				Usage:  "构建哈希表",
				Action: h.build,
				Flags: []cli.Flag{
					dirFlag(),
					typeFlag(),
					&cli.Uint64Flag{
						Name:  "start",
						Value: qqhash.DefaultStart,
						Usage: "起始 QQ 号",
					},
					&cli.Uint64Flag{
						Name:    "end",
						Aliases: []string{"sum"},
						Value:   qqhash.DefaultEnd,
						Usage:   "结束 QQ 号",
					},
					&cli.UintFlag{
						Name:  "partition-bits",
						Value: qqhash.DefaultPartBits,
						Usage: "分区位数，分区数为 2 的该次幂",
					},
					&cli.Float64Flag{
						Name:  "gamma",
						Value: mphf.DefaultGamma,
						Usage: "MPHF γ 参数，越大体积越大、查询越快",
					},
					workersFlag(),
				},
			},
			{
				Name:   "verify",
				Usage:  "校验哈希表",
				Action: h.verify,
				Flags: []cli.Flag{
					dirFlag(),
					typeFlag(),
					&cli.IntFlag{
						Name:  "sample",
						Value: 100000,
						Usage: "随机抽样查询次数，0 为跳过",
					},
					&cli.BoolFlag{
						Name:  "full",
						Usage: "按槽位全量校验",
					},
					&cli.BoolFlag{
						Name:  "no-coverage",
						Usage: "全量校验时跳过覆盖检查以节省内存",
					},
					workersFlag(),
				},
			},
			{
				Name:      "lookup",
				Usage:     "查询哈希对应的 QQ 号",
				ArgsUsage: "<hash>",
				Action:    h.lookup,
				Flags:     []cli.Flag{dirFlag()},
			},
			{
				Name:   "stat",
				Usage:  "查看哈希表信息",
				Action: h.stat,
				Flags:  []cli.Flag{dirFlag()},
			},
		},
	}
}

func (h hashCommand) build(ctx context.Context, cmd *cli.Command) error {
	start := time.Now()
	err := qqhash.Build(ctx, qqhash.BuildOptions{
		Dir:      h.dirOf(cmd),
		Types:    cmd.StringSlice("type"),
		Start:    cmd.Uint64("start"),
		End:      cmd.Uint64("end"),
		PartBits: uint32(cmd.Uint("partition-bits")), //nolint:gosec // qqhash rejects out-of-range bits
		Gamma:    cmd.Float64("gamma"),
		Workers:  cmd.Int("workers"),
		Logf:     logf,
	})
	if err != nil {
		return err
	}

	color.Greenf("全部构建完成，总耗时 %s\n", time.Since(start).Round(time.Second))
	return nil
}

func (h hashCommand) verify(ctx context.Context, cmd *cli.Command) error {
	tables, err := h.open(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = tables.Close() }()

	types := cmd.StringSlice("type")
	sample := cmd.Int("sample")
	for _, t := range tables.All() {
		if len(types) > 0 && !slices.Contains(types, t.Type()) {
			continue
		}
		printStats(t.Stats())

		if sample > 0 {
			if err = t.Sample(ctx, sample, uint64(time.Now().UnixNano())); err != nil {
				return err
			}
			color.Greenf("[%s] 随机抽样 %d 次查询通过\n", t.Type(), sample)
		}
		if cmd.Bool("full") {
			if err = t.Verify(ctx, qqhash.VerifyOptions{
				Workers:  cmd.Int("workers"),
				Coverage: !cmd.Bool("no-coverage"),
				Logf:     logf,
			}); err != nil {
				return err
			}
			color.Greenf("[%s] 全量校验通过\n", t.Type())
		}
	}

	return nil
}

func (h hashCommand) lookup(_ context.Context, cmd *cli.Command) error {
	hash := strings.ToLower(cmd.Args().First())
	if hash == "" {
		return errors.New("请提供要查询的哈希")
	}

	tables, err := h.open(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = tables.Close() }()

	start := time.Now()
	qq, ok := tables.Lookup(hash)
	elapsed := time.Since(start)
	if !ok {
		return fmt.Errorf("未命中（%s）", elapsed)
	}

	color.Greenf("QQ: %d（%s）\n", qq, elapsed)
	return nil
}

func (h hashCommand) stat(_ context.Context, cmd *cli.Command) error {
	tables, err := h.open(cmd)
	if err != nil {
		return err
	}
	defer func() { _ = tables.Close() }()

	for _, t := range tables.All() {
		printStats(t.Stats())
	}
	return nil
}

func (h hashCommand) dirOf(cmd *cli.Command) string {
	if dir := cmd.String("dir"); dir != "" {
		return dir
	}
	if h.dir != "" {
		return h.dir
	}
	return defaultHashDir
}

func (h hashCommand) open(cmd *cli.Command) (*qqhash.Tables, error) {
	dir := h.dirOf(cmd)
	tables, err := qqhash.Open(dir)
	if err != nil {
		return nil, err
	}
	if len(tables.All()) == 0 {
		_ = tables.Close()
		return nil, fmt.Errorf("目录 %s 中没有哈希表文件", dir)
	}
	return tables, nil
}

func dirFlag() cli.Flag {
	return &cli.StringFlag{
		Name:  "dir",
		Usage: "哈希表目录，默认取配置 hash.dir",
	}
}

func typeFlag() cli.Flag {
	return &cli.StringSliceFlag{
		Name:  "type",
		Usage: "哈希类型 md5, sha256，默认全部",
		Validator: func(types []string) error {
			for _, t := range types {
				if !slices.Contains(qqhash.Types, t) {
					return errors.New("哈希类型只能是 md5 或 sha256")
				}
			}
			return nil
		},
	}
}

func workersFlag() cli.Flag {
	return &cli.IntFlag{
		Name:  "workers",
		Value: runtime.NumCPU(),
		Usage: "并行数",
	}
}

func printStats(s qqhash.Stats) {
	color.Warnf("[%s]\n", s.Type)
	fmt.Printf("  QQ 号范围: %d ~ %d（%d 个）\n", s.Start, s.End, s.KeyCount)
	fmt.Printf("  分区数: %d，最大层数: %d，MPHF %.2f bit/键\n", s.Partitions, s.MaxLevels, s.BitsPerKey)
	fmt.Printf("  idx: %s，val: %s\n", qqhash.FormatSize(s.IdxSize), qqhash.FormatSize(s.ValSize))
}

func logf(format string, args ...any) {
	color.Greenf(format+"\n", args...)
}
