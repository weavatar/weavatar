package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000003_create_apps_table", createAppsTable,
		migrate.WithDown(dropAppsTable))
}

func createAppsTable(schema *migrate.Schema) {
	schema.Create("apps", func(table *migrate.Table) {
		table.String("id", 10).Primary()
		table.Timestamps()
		table.String("user_id", 10).Index()
		table.String("name")
		table.String("secret")
	})
}

func dropAppsTable(schema *migrate.Schema) {
	schema.DropIfExists("apps")
}
