//go:build !cgo

package data

import (
	"log/slog"
	"time"

	"github.com/knadh/koanf/v2"
	"github.com/libtnb/cache"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/internal/http/request"
	"github.com/weavatar/weavatar/pkg/queue"
)

const (
	CacheThreshold = 14 * 24 * time.Hour
)

type avatarRepo struct{}

func NewAvatarRepo(cache cache.Cache, conf *koanf.Koanf, db *gorm.DB, log *slog.Logger, queue *queue.Queue) (biz.AvatarRepo, error) {
	return &avatarRepo{}, nil
}

func (r *avatarRepo) List(userID string, page, limit uint) ([]*biz.Avatar, int64, error) {
	return nil, 0, nil
}

func (r *avatarRepo) Get(userID string, hash string) (*biz.Avatar, error) {
	return nil, nil
}

func (r *avatarRepo) Create(userID string, req *request.AvatarCreate) (*biz.Avatar, error) {
	return nil, nil
}

func (r *avatarRepo) Update(userID string, req *request.AvatarUpdate) (*biz.Avatar, error) {
	return nil, nil
}

func (r *avatarRepo) Delete(userID string, hash string) error {
	return nil
}

func (r *avatarRepo) GetByRaw(raw string) (*biz.Avatar, error) {
	return nil, nil
}

func (r *avatarRepo) GetWeAvatar(hash, appID string) (string, []byte, time.Time, error) {
	return "", nil, time.Now(), nil
}

func (r *avatarRepo) GetQqByHash(hash string) (string, []byte, time.Time, error) {
	return "", nil, time.Now(), nil
}

func (r *avatarRepo) GetGravatarByHash(hash string) ([]byte, time.Time, error) {
	return nil, time.Now(), nil
}

func (r *avatarRepo) GetByType(avatarType string, options ...string) ([]byte, time.Time, error) {
	return nil, time.Now(), nil
}

func (r *avatarRepo) IsBanned(hash, appID string, img []byte) (bool, error) {
	return false, nil
}
