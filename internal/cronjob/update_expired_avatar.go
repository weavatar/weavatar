package cronjob

import (
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

const (
	ExpiredThreshold = 14 * 24 * time.Hour
)

type UpdateExpiredAvatar struct {
	log *slog.Logger
}

func NewUpdateExpiredAvatar(log *slog.Logger) *UpdateExpiredAvatar {
	return &UpdateExpiredAvatar{
		log: log,
	}
}

func (r *UpdateExpiredAvatar) Run() {
	cachePath := filepath.Join("storage", "cache")

	if _, err := os.Stat(cachePath); os.IsNotExist(err) {
		r.log.Info("[UpdateExpiredAvatar] cache directory doesn't exist", slog.String("path", cachePath))
		return
	}

	err := filepath.Walk(cachePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			r.log.Error("[UpdateExpiredAvatar] error accessing path", slog.String("path", path), slog.Any("err", err))
			return nil
		}

		if !info.IsDir() {
			if info.ModTime().Add(ExpiredThreshold).Before(time.Now()) {
				if err = os.Remove(path); err != nil {
					r.log.Error("[UpdateExpiredAvatar] failed to delete expired avatar", slog.String("path", path), slog.Any("err", err))
				}
			}
		}

		return nil
	})

	if err != nil {
		r.log.Error("[UpdateExpiredAvatar] failed to walk cache directory", slog.Any("err", err))
	}
}
