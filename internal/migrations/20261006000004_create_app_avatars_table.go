package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000004_create_app_avatars_table", createAppAvatarsTable,
		migrate.WithDown(dropAppAvatarsTable))
}

func createAppAvatarsTable(schema *migrate.Schema) {
	schema.Create("app_avatars", func(table *migrate.Table) {
		table.String("avatar_sha256", 64)
		table.String("avatar_md5", 32)
		table.String("app_id", 10)
		table.Timestamps()
		table.Primary("avatar_sha256", "avatar_md5", "app_id")
		table.Index("avatar_md5")
	})
}

func dropAppAvatarsTable(schema *migrate.Schema) {
	schema.DropIfExists("app_avatars")
}
