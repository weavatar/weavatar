package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000002_create_avatars_table", func(s *migrate.Schema) {
		s.Create("avatars", func(t *migrate.Table) {
			t.String("sha256", 64).Primary()
			t.String("md5", 32).Index()
			t.String("raw", 255).Index()
			t.String("user_id", 10).Index()
			t.Timestamps()
		})
	})
}
