package queuejob

import (
	"errors"
	"fmt"
	"github.com/go-rat/utils/debug"
	"log/slog"
	"path/filepath"
	"time"

	"github.com/go-rat/cache"
	"github.com/go-rat/utils/convert"
	"github.com/go-rat/utils/file"
	"github.com/go-rat/utils/str"
	"github.com/knadh/koanf/v2"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/pkg/audit"
	"github.com/weavatar/weavatar/pkg/cdn"
)

type ProcessAvatarAudit struct {
	cache      cache.Cache
	conf       *koanf.Koanf
	db         *gorm.DB
	log        *slog.Logger
	avatarRepo biz.AvatarRepo
}

func NewProcessAvatarAudit(cache cache.Cache, conf *koanf.Koanf, db *gorm.DB, log *slog.Logger, avatarRepo biz.AvatarRepo) *ProcessAvatarAudit {
	return &ProcessAvatarAudit{
		cache:      cache,
		conf:       conf,
		db:         db,
		log:        log,
		avatarRepo: avatarRepo,
	}
}

func (r *ProcessAvatarAudit) Handle(args ...any) error {
	if len(args) < 2 {
		return errors.New("arguments are not enough")
	}

	hash, ok := args[0].(string)
	if !ok {
		return errors.New("failed to assert hash")
	}

	appID, ok2 := args[1].(string)
	if !ok2 {
		return errors.New("failed to assert appID")
	}

	// 防止并发下重复审核
	if r.cache.Has("avatar:check:" + hash) {
		return nil
	}
	if err := r.cache.Put("avatar:check:"+hash, true, 30*time.Second); err != nil {
		return fmt.Errorf("%w, hash: %s", err, hash)
	}

	_, img, _, err := r.avatarRepo.GetWeAvatar(hash, appID)
	if err != nil {
		img, _, err = r.avatarRepo.GetGravatarByHash(hash)
	}
	if err != nil {
		return fmt.Errorf("%w, hash: %s", err, hash)
	}

	imgHash := str.SHA256(convert.UnsafeString(img))
	if err = file.Write(filepath.Join("storage", "checker", imgHash[:2], imgHash), img); err != nil {
		return fmt.Errorf("%w, hash: %s, imgHash: %s", err, hash, imgHash)
	}

	image := new(biz.Image)
	image.Hash = imgHash
	if err = r.db.Where("hash = ?", imgHash).First(image).Error; err != nil {
		auditor := audit.NewAudit(r.conf)
		image.Banned, image.Remark, err = auditor.Check("https://weavatar.com/avatar/" + hash + ".png?s=600&d=404")
		if err != nil {
			return fmt.Errorf("%w, hash: %s, imgHash: %s", err, hash, imgHash)
		}
		if err = r.db.Create(image).Error; err != nil {
			return fmt.Errorf("%w, hash: %s, imgHash: %s", err, hash, imgHash)
		}
	}

	debug.Dump("[ProcessAvatarAudit] image", slog.String("hash", hash), slog.String("imgHash", imgHash), slog.Bool("banned", image.Banned), slog.String("remark", image.Remark))

	if image.Banned {
		debug.Dump("[ProcessAvatarAudit] image banned", slog.String("hash", hash), slog.String("imgHash", imgHash), slog.String("remark", image.Remark))
		if err = cdn.NewCdn(r.conf).RefreshUrl([]string{fmt.Sprintf("https://%s/avatar/%s", r.conf.MustString("http.domain"), hash)}); err != nil {
			r.log.Error("[ProcessAvatarAudit] failed to refresh url", slog.String("url", fmt.Sprintf("https://%s/avatar/%s", r.conf.MustString("http.domain"), hash)), slog.Any("err", err))
		}
	}

	return nil
}

func (r *ProcessAvatarAudit) ErrHandle(err error) {
	if err != nil {
		r.log.Error("[ProcessAvatarAudit] failed to process avatar audit", slog.Any("err", err))
	}
}
