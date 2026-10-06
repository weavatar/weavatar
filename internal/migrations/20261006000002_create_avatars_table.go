package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000002_create_avatars_table", createAvatarsTable,
		migrate.WithDown(dropAvatarsTable))
}

func createAvatarsTable(schema *migrate.Schema) {
	schema.Create("avatars", func(table *migrate.Table) {
		table.String("sha256", 64).Primary()
		table.Timestamps()
		table.String("md5", 32).Index()
		table.String("raw", 255).Index()
		table.String("user_id", 10).Index()
	})
}

func dropAvatarsTable(schema *migrate.Schema) {
	schema.DropIfExists("avatars")
}
