package service

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/gookit/color"
	"github.com/knadh/koanf/v2"
	"github.com/urfave/cli/v3"

	"github.com/weavatar/weavatar/pkg/qqhash"
)

type CliService struct {
	conf *koanf.Koanf
}

func NewCliService(conf *koanf.Koanf) *CliService {
	return &CliService{
		conf: conf,
	}
}

// hashDir 命令行参数优先于配置
func (r *CliService) hashDir(cmd *cli.Command) string {
	if dir := cmd.String("dir"); dir != "" {
		return dir
	}
	if dir := r.conf.String("hash.dir"); dir != "" {
		return dir
	}
	return "storage/hash"
}

func (r *CliService) HashBuild(ctx context.Context, cmd *cli.Command) error {
	start := time.Now()
	err := qqhash.Build(ctx, qqhash.BuildOptions{
		Dir:      r.hashDir(cmd),
		Types:    cmd.StringSlice("type"),
		Start:    cmd.Uint64("start"),
		End:      cmd.Uint64("end"),
		PartBits: uint32(cmd.Uint("partition-bits")),
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

func (r *CliService) HashVerify(ctx context.Context, cmd *cli.Command) error {
	tables, err := r.openTables(cmd)
	if err != nil {
		return err
	}
	defer func() {
		_ = tables.Close()
	}()

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

func (r *CliService) HashLookup(_ context.Context, cmd *cli.Command) error {
	hash := strings.ToLower(cmd.Args().First())
	if hash == "" {
		return errors.New("请提供要查询的哈希")
	}

	tables, err := r.openTables(cmd)
	if err != nil {
		return err
	}
	defer func() {
		_ = tables.Close()
	}()

	start := time.Now()
	qq, ok := tables.Lookup(hash)
	elapsed := time.Since(start)
	if !ok {
		return fmt.Errorf("未命中（%s）", elapsed)
	}

	color.Greenf("QQ: %d（%s）\n", qq, elapsed)
	return nil
}

func (r *CliService) HashStat(_ context.Context, cmd *cli.Command) error {
	tables, err := r.openTables(cmd)
	if err != nil {
		return err
	}
	defer func() {
		_ = tables.Close()
	}()

	for _, t := range tables.All() {
		printStats(t.Stats())
	}

	return nil
}

func (r *CliService) openTables(cmd *cli.Command) (*qqhash.Tables, error) {
	dir := r.hashDir(cmd)
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

func printStats(s qqhash.Stats) {
	color.Warnf("[%s]\n", s.Type)
	fmt.Printf("  QQ 号范围: %d ~ %d（%d 个）\n", s.Start, s.End, s.KeyCount)
	fmt.Printf("  分区数: %d，最大层数: %d，MPHF %.2f bit/键\n", s.Partitions, s.MaxLevels, s.BitsPerKey)
	fmt.Printf("  idx: %s，val: %s\n", qqhash.FormatSize(s.IdxSize), qqhash.FormatSize(s.ValSize))
	fmt.Printf("  构建时间: %s\n", s.BuildTime.Format(time.DateTime))
}

func logf(format string, args ...any) {
	color.Greenf(format+"\n", args...)
}
