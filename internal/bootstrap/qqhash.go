package bootstrap

import (
	"log/slog"

	"github.com/knadh/koanf/v2"

	"github.com/weavatar/weavatar/pkg/qqhash"
)

// NewQqHash 加载 QQ 哈希映射表，文件缺失时只记录警告，此时 QQ 头像回退不可用。
func NewQqHash(conf *koanf.Koanf, log *slog.Logger) (*qqhash.Tables, error) {
	dir := conf.String("hash.dir")
	if dir == "" {
		dir = "storage/hash"
	}

	tables, err := qqhash.Open(dir)
	if err != nil {
		return nil, err
	}

	if len(tables.All()) == 0 {
		log.Warn("[QqHash] no hash table found, qq avatar fallback disabled", slog.String("dir", dir))
	}
	for _, t := range tables.All() {
		s := t.Stats()
		log.Info("[QqHash] table loaded",
			slog.String("type", s.Type),
			slog.Uint64("keys", s.KeyCount),
			slog.Uint64("start", s.Start),
			slog.Uint64("end", s.End),
			slog.Int("partitions", s.Partitions),
			slog.Float64("bits_per_key", s.BitsPerKey),
			slog.Time("build_time", s.BuildTime),
		)
	}

	return tables, nil
}
