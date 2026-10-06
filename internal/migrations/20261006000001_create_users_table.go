package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000001_create_users_table", createUsersTable,
		migrate.WithDown(dropUsersTable))
}

func createUsersTable(schema *migrate.Schema) {
	schema.Create("users", func(table *migrate.Table) {
		table.String("id", 10).Primary()
		table.Timestamps()
		table.SoftDeletes()
		table.String("open_id", 64)
		table.String("union_id", 64)
		table.String("nickname")
		table.String("avatar")
		table.Boolean("real_name").Default(false)
		// a soft-deleted user releases its union_id
		table.Unique("union_id").Where("deleted_at IS NULL")
	})
}

func dropUsersTable(schema *migrate.Schema) {
	schema.DropIfExists("users")
}
