package app

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"github.com/urfave/cli/v3"

	"github.com/weavatar/weavatar/internal/shared/registry"
)

var localizeOnce sync.Once

type Cli struct {
	cmd *cli.Command
}

func NewCli(cmd *cli.Command) *Cli {
	return &Cli{cmd: cmd}
}

// Run executes the command; SIGINT/SIGTERM cancel the context handed to it.
func (r *Cli) Run(version string) error {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	r.cmd.Version = version

	return r.cmd.Run(ctx, os.Args)
}

func newRootCommand(commands registry.Commands) *cli.Command {
	localizeHelp()

	return &cli.Command{
		Name:     "cli",
		Usage:    "WeAvatar 管理命令",
		Commands: commands,
	}
}

// localizeHelp translates the section headings of urfave/cli's help templates.
func localizeHelp() {
	localizeOnce.Do(func() {
		cli.RootCommandHelpTemplate = strings.NewReplacer(
			"GLOBAL OPTIONS", "全局选项",
			"NAME", "名称",
			"USAGE", "用法",
			"VERSION", "版本",
			"DESCRIPTION", "描述",
			"AUTHOR", "作者",
			"COMMANDS", "命令",
			"COPYRIGHT", "版权",
		).Replace(cli.RootCommandHelpTemplate)
		cli.CommandHelpTemplate = strings.NewReplacer(
			"NAME", "名称",
			"USAGE", "用法",
			"CATEGORY", "分类",
			"DESCRIPTION", "描述",
			"OPTIONS", "选项",
		).Replace(cli.CommandHelpTemplate)
		cli.SubcommandHelpTemplate = strings.NewReplacer(
			"NAME", "名称",
			"USAGE", "用法",
			"DESCRIPTION", "描述",
			"COMMANDS", "命令",
			"OPTIONS", "选项",
		).Replace(cli.SubcommandHelpTemplate)
	})
}
