package route

import (
	"fmt"
	"slices"

	"github.com/urfave/cli/v3"

	"github.com/weavatar/weavatar/internal/service"
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
			Usage: "哈希表操作",
			Commands: []*cli.Command{
				{
					Name:   "make",
					Usage:  "生成哈希表",
					Action: r.cli.HashMake,
					Flags: []cli.Flag{
						&cli.UintFlag{
							Name:     "sum",
							Value:    4000000000,
							Usage:    "生成的QQ号最大值",
							Required: true,
						},
						&cli.StringFlag{
							Name:     "dir",
							Value:    "hash",
							Usage:    "生成的文件存放目录",
							Required: true,
						},
						&cli.StringFlag{
							Name:     "type",
							Value:    "md5",
							Usage:    "生成的哈希类型 md5, sha256",
							Required: true,
							Validator: func(v string) error {
								if !slices.Contains([]string{"md5", "sha256"}, v) {
									return fmt.Errorf("哈希类型只能是 md5 或 sha256")
								}
								return nil
							},
						},
					},
				},
				{
					Name:   "insert",
					Usage:  "插入哈希表",
					Action: r.cli.HashInsert,
					Flags: []cli.Flag{
						&cli.StringFlag{
							Name:     "dir",
							Value:    "hash",
							Usage:    "哈希文件存放目录",
							Required: true,
						},
						&cli.StringFlag{
							Name:     "type",
							Value:    "md5",
							Usage:    "哈希类型 md5, sha256",
							Required: true,
							Validator: func(v string) error {
								if !slices.Contains([]string{"md5", "sha256"}, v) {
									return fmt.Errorf("哈希类型只能是 md5 或 sha256")
								}
								return nil
							},
						},
						&cli.BoolWithInverseFlag{
							BoolFlag: &cli.BoolFlag{
								Name:  "rocksdb",
								Usage: "是否使用 rocksdb",
							},
						},
					},
				},
			},
		},
	}
}
