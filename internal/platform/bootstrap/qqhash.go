package bootstrap

import (
	"log/slog"

	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/pkg/qqhash"
)

// NewQqHash maps the QQ hash tables; missing tables only disable the QQ
// avatar fallback.
func NewQqHash(config *conf.Config, log *slog.Logger) (*qqhash.Tables, func() error, error) {
	dir := config.Hash.Dir
	tables, err := qqhash.Open(dir)
	if err != nil {
		return nil, nil, err
	}

	if len(tables.All()) == 0 {
		log.Warn("qq hash table not found, qq avatar fallback disabled", slog.String("dir", dir))
	}
	for _, t := range tables.All() {
		s := t.Stats()
		log.Info("qq hash table loaded",
			slog.String("dir", dir),
			slog.String("type", s.Type),
			slog.Uint64("keys", s.KeyCount),
			slog.Uint64("start", s.Start),
			slog.Uint64("end", s.End),
			slog.Int("partitions", s.Partitions),
			slog.Float64("bits_per_key", s.BitsPerKey),
		)
	}

	return tables, tables.Close, nil
}
