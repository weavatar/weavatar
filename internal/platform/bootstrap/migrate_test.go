package bootstrap

import (
	"testing"

	"github.com/go-rio/migrate"
	"github.com/libtnb/assert/must"
)

func TestMigrateCommandTree(t *testing.T) {
	cmd := MigrateCommand(nil)

	must.Equal(t, cmd.Name, "migrate")
	must.NotNil(t, cmd.Action, must.Msgf("bare migrate applies pending migrations"))
	names := make([]string, 0, len(cmd.Commands))
	for _, sub := range cmd.Commands {
		names = append(names, sub.Name)
		must.NotNil(t, sub.Action, must.Msgf("%s has an action", sub.Name))
	}
	must.DeepEqual(t, names, []string{"up", "plan", "status", "rollback"})
}

func TestRenderPlan(t *testing.T) {
	must.Equal(t, renderPlan(nil), "没有待执行的迁移\n")

	c := migrate.NewCollection()
	c.Add("20260101000001_create_things_table", func(s *migrate.Schema) {
		s.Create("things", func(t *migrate.Table) {
			t.String("id", 10).Primary()
			t.String("name").Index()
		})
	})
	c.Add("20260101000002_seed_things", func(s *migrate.Schema) {
		s.Exec("INSERT INTO things (id, name) VALUES (?, ?)", "t1", "first")
	})
	planned, err := c.SQL(migrate.Postgres)
	must.NoError(t, err)

	want := `-- 20260101000001_create_things_table
CREATE TABLE "things" (
	"id" VARCHAR(10) NOT NULL,
	"name" VARCHAR(255) NOT NULL,
	CONSTRAINT "things_pkey" PRIMARY KEY ("id")
);
CREATE INDEX "things_name_index" ON "things" ("name");

-- 20260101000002_seed_things
INSERT INTO things (id, name) VALUES (?, ?);
-- args: [t1 first]
`
	must.Equal(t, renderPlan(planned), want)
}

func TestRenderPlanWarningsAndPlaceholders(t *testing.T) {
	got := renderPlan([]migrate.Planned{{
		Name:       "20260101000003_backfill",
		Warnings:   []string{"drops column"},
		Statements: []string{"ALTER TABLE things DROP COLUMN name", "-- Go function: not renderable, runs at migration time"},
	}})
	must.Equal(t, got, `-- 20260101000003_backfill
-- 警告：drops column
ALTER TABLE things DROP COLUMN name;
-- Go function: not renderable, runs at migration time
`)
}
