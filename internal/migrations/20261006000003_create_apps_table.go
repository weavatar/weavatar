package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000003_create_apps_table", func(s *migrate.Schema) {
		s.Create("apps", func(t *migrate.Table) {
			t.String("id", 10).Primary()
			t.Timestamps()
			t.String("user_id", 10).Index()
			t.String("name")
			t.String("secret")
		})
	})
}
