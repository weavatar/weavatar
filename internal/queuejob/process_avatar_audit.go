package queuejob

import (
	"errors"
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
)

type ProcessAvatarAudit struct {
	cache      cache.Cache
	conf       *koanf.Koanf
	db         *gorm.DB
	avatarRepo biz.AvatarRepo
}

func NewProcessAvatarAudit(cache cache.Cache, conf *koanf.Koanf, db *gorm.DB, avatarRepo biz.AvatarRepo) *ProcessAvatarAudit {
	return &ProcessAvatarAudit{
		cache:      cache,
		conf:       conf,
		db:         db,
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
	_ = r.cache.Put("avatar:check:"+hash, true, 30*time.Second)
	defer r.cache.Forget("avatar:check:" + hash)

	_, img, _, err := r.avatarRepo.GetWeAvatar(hash, appID)
	if err != nil {
		img, _, err = r.avatarRepo.GetGravatarByHash(hash)
	}
	if err != nil {
		return err
	}

	imgHash := str.SHA256(convert.UnsafeString(img))
	if err = file.Write(filepath.Join("storage", "checker", imgHash[:2], imgHash), img); err != nil {
		return err
	}

	image := new(biz.Image)
	image.Hash = imgHash
	if err = r.db.Where("hash = ?", imgHash).First(image).Error; err != nil {
		auditor := audit.NewAudit(r.conf)
		image.Banned, err = auditor.Check("https://weavatar.com/avatar/" + hash + ".png?s=600&d=404")
		if err != nil {
			return err
		}
		if err = r.db.Create(image).Error; err != nil {
			return err
		}
	}

	// TODO 刷新CDN
	if image.Banned {

	}

	return nil
}
