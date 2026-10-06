package bootstrap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/go-rio/migrate"
	"github.com/go-rio/rio"
	"github.com/urfave/cli/v3"
)

// NewMigrate builds the migrator; history lives in schema_migrations.
func NewMigrate(db *rio.DB, c *migrate.Collection, log *slog.Logger) (*migrate.Migrator, error) {
	return migrate.New(db.Unwrap(), migrate.Postgres,
		migrate.WithCollection(c),
		migrate.WithLogger(log),
	)
}

// MigrateCommand is the "migrate" CLI command, where bare "migrate" runs "up";
// output goes to the root command's Writer.
func MigrateCommand(m *migrate.Migrator) *cli.Command {
	up := func(ctx context.Context, cmd *cli.Command) error {
		if err := m.Up(ctx); err != nil {
			return err
		}
		return write(cmd, "数据库迁移完成\n")
	}

	return &cli.Command{
		Name:   "migrate",
		Usage:  "执行未应用的数据库迁移",
		Action: up,
		Commands: []*cli.Command{
			{
				Name:   "up",
				Usage:  "执行未应用的数据库迁移（同 migrate）",
				Action: up,
			},
			{
				Name:  "plan",
				Usage: "打印待执行迁移的 SQL，不执行",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					planned, err := m.Plan(ctx)
					if err != nil {
						return err
					}
					return write(cmd, renderPlan(planned))
				},
			},
			{
				Name:  "status",
				Usage: "查看迁移及其应用状态",
				Action: func(ctx context.Context, cmd *cli.Command) error {
					statuses, err := m.Status(ctx)
					if err != nil {
						return err
					}
					var b strings.Builder
					for _, s := range statuses {
						state := "未应用"
						switch {
						case s.Drifted:
							state = "已漂移（校验和不一致）"
						case !s.Registered:
							state = "已应用，但代码中不存在"
						case s.Applied:
							state = "已应用 " + s.AppliedAt.Local().Format(time.DateTime)
						}
						fmt.Fprintf(&b, "%s\t%s\n", s.Name, state)
					}
					return write(cmd, b.String())
				},
			},
			{
				Name:  "rollback",
				Usage: "回滚最近应用的迁移",
				Flags: []cli.Flag{
					&cli.IntFlag{Name: "step", Value: 1, Usage: "回滚的迁移数量"},
				},
				Action: func(ctx context.Context, cmd *cli.Command) error {
					if err := m.Rollback(ctx, cmd.Int("step")); err != nil {
						return err
					}
					return write(cmd, "回滚完成\n")
				},
			},
		},
	}
}

// renderPlan prints each pending migration as a commented, runnable SQL block.
func renderPlan(planned []migrate.Planned) string {
	if len(planned) == 0 {
		return "没有待执行的迁移\n"
	}

	var b strings.Builder
	for i, p := range planned {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("-- " + p.Name + "\n")
		for _, w := range p.Warnings {
			b.WriteString("-- 警告：" + w + "\n")
		}
		for _, stmt := range p.Statements {
			b.WriteString(terminate(stmt) + "\n")
		}
	}
	return b.String()
}

// terminate appends ";" except to comment-only statements, keeping a trailing
// "-- args:" line after it.
func terminate(stmt string) string {
	if strings.HasPrefix(stmt, "--") {
		return stmt
	}
	if sql, args, ok := strings.Cut(stmt, "\n-- args:"); ok {
		return sql + ";\n-- args:" + args
	}
	return stmt + ";"
}

func write(cmd *cli.Command, s string) error {
	_, err := io.WriteString(cmd.Root().Writer, s)
	return err
}
