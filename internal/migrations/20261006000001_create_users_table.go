package migrations

import "github.com/go-rio/migrate"

func init() {
	collection.Add("20261006000001_create_users_table", func(s *migrate.Schema) {
		s.Create("users", func(t *migrate.Table) {
			t.String("id", 10).Primary()
			t.String("open_id", 64)
			t.String("union_id", 64)
			t.String("nickname")
			t.String("avatar")
			t.Boolean("real_name").Default(false)
			t.Timestamps()
			t.SoftDeletes()
			// a soft-deleted user releases its union_id
			t.Unique("union_id").Where("deleted_at IS NULL")
		})
	})
}
