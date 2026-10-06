package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000005_create_images_table", createImagesTable,
		migrate.WithDown(dropImagesTable))
}

func createImagesTable(schema *migrate.Schema) {
	schema.Create("images", func(table *migrate.Table) {
		table.String("hash", 64).Primary()
		table.Timestamps()
		table.Boolean("banned").Default(false)
		table.Text("remark")
	})
}

func dropImagesTable(schema *migrate.Schema) {
	schema.DropIfExists("images")
}
