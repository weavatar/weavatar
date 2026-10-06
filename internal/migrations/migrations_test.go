package migrations_test

import (
	"strings"
	"testing"

	"github.com/go-rio/migrate"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/migrations"
)

func TestMigrationsRenderForPostgres(t *testing.T) {
	planned, err := migrations.Collection().SQL(migrate.Postgres)
	must.NoError(t, err)

	names := make([]string, 0, len(planned))
	var ddl strings.Builder
	for _, p := range planned {
		names = append(names, p.Name)
		for _, stmt := range p.Statements {
			ddl.WriteString(stmt)
			ddl.WriteString(";\n")
		}
	}
	must.DeepEqual(t, names, []string{
		"20261006000001_create_users_table",
		"20261006000002_create_avatars_table",
		"20261006000003_create_apps_table",
		"20261006000004_create_app_avatars_table",
		"20261006000005_create_images_table",
	})

	sql := strings.ToLower(ddl.String())
	// fixed-width CHAR pads values; every text key is VARCHAR
	must.NotContains(t, sql, " char(")
	must.NotContains(t, sql, "bpchar")
	for _, want := range []string{
		`create table "users"`,
		`"union_id" varchar(64) not null`,
		`create unique index "users_union_id_unique" on "users" ("union_id") where deleted_at is null`,
		`"deleted_at" timestamptz`,
		`create table "app_avatars"`,
		`primary key ("avatar_sha256", "avatar_md5", "app_id")`,
		`"remark" text not null`,
	} {
		must.Contains(t, sql, want)
	}
}
