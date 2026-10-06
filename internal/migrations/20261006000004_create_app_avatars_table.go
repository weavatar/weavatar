package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000004_create_app_avatars_table", func(s *migrate.Schema) {
		s.Create("app_avatars", func(t *migrate.Table) {
			t.String("avatar_sha256", 64)
			t.String("avatar_md5", 32)
			t.String("app_id", 10)
			t.Timestamps()
			t.Primary("avatar_sha256", "avatar_md5", "app_id")
			t.Index("avatar_md5")
		})
	})
}
