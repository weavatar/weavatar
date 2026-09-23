package route

import (
	"fmt"
	"runtime"
	"slices"

	"github.com/urfave/cli/v3"

	"github.com/weavatar/weavatar/internal/service"
	"github.com/weavatar/weavatar/pkg/mphf"
	"github.com/weavatar/weavatar/pkg/qqhash"
)

type Cli struct {
	cli *service.CliService
}

func NewCli(cli *service.CliService) *Cli {
	return &Cli{
		cli: cli,
	}
}

func (r *Cli) Commands() []*cli.Command {
	return []*cli.Command{
		{
			Name:  "hash",
			Usage: "QQ 哈希表操作",
			Commands: []*cli.Command{
				{
					Name:   "build",
					Usage:  "构建哈希表",
					Action: r.cli.HashBuild,
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
					Action: r.cli.HashVerify,
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
					Action:    r.cli.HashLookup,
					Flags: []cli.Flag{
						dirFlag(),
					},
				},
				{
					Name:   "stat",
					Usage:  "查看哈希表信息",
					Action: r.cli.HashStat,
					Flags: []cli.Flag{
						dirFlag(),
					},
				},
			},
		},
	}
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
					return fmt.Errorf("哈希类型只能是 md5 或 sha256")
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
