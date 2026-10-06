package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000005_create_images_table", func(s *migrate.Schema) {
		s.Create("images", func(t *migrate.Table) {
			t.String("hash", 64).Primary()
			t.Timestamps()
			t.Boolean("banned").Default(false)
			t.Text("remark")
		})
	})
}
