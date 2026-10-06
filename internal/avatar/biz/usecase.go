package biz

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/go-rio/rio"
	"github.com/libtnb/cache"
	"github.com/libtnb/utils/str"

	"github.com/weavatar/weavatar/internal/shared/apperr"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/pkg/queue"
)

const (
	// CacheTTL is how long a fetched Gravatar or QQ avatar is served from disk;
	// older files are purged hourly.
	CacheTTL = 14 * 24 * time.Hour

	// auditGuard suppresses re-enqueuing the audit of one hash for a while.
	auditGuard = 30 * time.Second
)

// ErrAvatarExists reports an avatar whose sha256 is already taken.
var ErrAvatarExists = errors.New("avatar already exists")

// AvatarUsecase is shared by HTTP, CLI, the audit queue and the scheduler.
type AvatarUsecase struct {
	repo    AvatarRepo
	images  ImageRepo
	tx      TxRunner
	users   Users
	store   Store
	fetcher Fetcher
	qq      QQHashes
	gen     Generator
	purger  Purger
	auditor Auditor
	queue   *queue.Queue
	cache   cache.Cache
	domain  string
	log     *slog.Logger
}

func NewAvatarUsecase(
	repo AvatarRepo,
	images ImageRepo,
	tx TxRunner,
	users Users,
	store Store,
	fetcher Fetcher,
	qq QQHashes,
	gen Generator,
	purger Purger,
	auditor Auditor,
	queue *queue.Queue,
	cache cache.Cache,
	domain appinfo.Domain,
	log *slog.Logger,
) *AvatarUsecase {
	return &AvatarUsecase{
		repo:    repo,
		images:  images,
		tx:      tx,
		users:   users,
		store:   store,
		fetcher: fetcher,
		qq:      qq,
		gen:     gen,
		purger:  purger,
		auditor: auditor,
		queue:   queue,
		cache:   cache,
		domain:  string(domain),
		log:     log,
	}
}

func (uc *AvatarUsecase) List(ctx context.Context, userID string, page, limit int) ([]*Avatar, int64, error) {
	return uc.repo.List(ctx, userID, page, limit)
}

// Create binds raw (an email or phone) to img for userID.
func (uc *AvatarUsecase) Create(ctx context.Context, userID, raw string, img []byte) (*Avatar, error) {
	img, err := formatUpload(img)
	if err != nil {
		return nil, err
	}

	avatar := &Avatar{
		SHA256: str.SHA256(raw),
		MD5:    str.MD5(raw),
		Raw:    raw,
		UserID: userID,
	}
	err = uc.tx.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Create(ctx, avatar); err != nil {
			return err
		}
		return uc.store.WriteAvatar(avatar.SHA256, img)
	})
	if err != nil {
		// lost the not_exists pre-check race
		if errors.Is(err, ErrAvatarExists) {
			return nil, apperr.Conflict("avatar.exists", "头像已存在").In("avatar").Wrap(err)
		}
		return nil, err
	}

	uc.enqueueRefresh(ctx, uc.avatarURL(avatar.SHA256), uc.avatarURL(avatar.MD5))
	return avatar, nil
}

// Update replaces the image of userID's avatar matching hash.
func (uc *AvatarUsecase) Update(ctx context.Context, userID, hash string, img []byte) (*Avatar, error) {
	avatar, err := uc.repo.Find(ctx, userID, hash)
	if err != nil {
		return nil, err
	}
	img, err = formatUpload(img)
	if err != nil {
		return nil, err
	}

	err = uc.tx.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Touch(ctx, avatar); err != nil {
			return err
		}
		return uc.store.WriteAvatar(avatar.SHA256, img)
	})
	if err != nil {
		return nil, err
	}

	uc.enqueueRefresh(ctx, uc.avatarURL(avatar.SHA256), uc.avatarURL(avatar.MD5))
	return avatar, nil
}

// Delete removes userID's avatar matching hash together with its image.
func (uc *AvatarUsecase) Delete(ctx context.Context, userID, hash string) error {
	avatar, err := uc.repo.Find(ctx, userID, hash)
	if err != nil {
		return err
	}

	err = uc.tx.Run(ctx, func(ctx context.Context) error {
		if err := uc.repo.Delete(ctx, avatar); err != nil {
			return err
		}
		return uc.store.RemoveAvatar(avatar.SHA256)
	})
	if err != nil {
		return err
	}

	uc.enqueueRefresh(ctx, uc.avatarURL(avatar.SHA256), uc.avatarURL(avatar.MD5))
	return nil
}

// DeleteByUser removes all of userID's avatars and images, joining the
// caller's transaction, and purges the CDN once.
func (uc *AvatarUsecase) DeleteByUser(ctx context.Context, userID string) error {
	avatars, err := uc.repo.ListAllByUser(ctx, userID)
	if err != nil {
		return err
	}
	if len(avatars) == 0 {
		return nil
	}

	err = uc.tx.Run(ctx, func(ctx context.Context) error {
		for _, avatar := range avatars {
			if err := uc.repo.Delete(ctx, avatar); err != nil {
				return err
			}
			if err := uc.store.RemoveAvatar(avatar.SHA256); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}

	urls := make([]string, 0, 2*len(avatars))
	for _, avatar := range avatars {
		urls = append(urls, uc.avatarURL(avatar.SHA256), uc.avatarURL(avatar.MD5))
	}
	uc.enqueueRefresh(ctx, urls...)
	return nil
}

// Bound reports whether raw already has an avatar.
func (uc *AvatarUsecase) Bound(ctx context.Context, raw string) (bool, error) {
	return uc.repo.ExistsByRaw(ctx, raw)
}

// FetchQQ downloads the current avatar of a QQ number.
func (uc *AvatarUsecase) FetchQQ(ctx context.Context, qq string) ([]byte, error) {
	return uc.fetcher.QQ(ctx, qq)
}

// Random returns the sha256 of up to n random avatars.
func (uc *AvatarUsecase) Random(ctx context.Context, n int) ([]string, error) {
	return uc.repo.RandomHashes(ctx, n)
}

// PurgeExpiredCache removes fetched avatars older than CacheTTL.
func (uc *AvatarUsecase) PurgeExpiredCache(ctx context.Context) error {
	return uc.store.PurgeCache(ctx, CacheTTL)
}

// Audit moderates the image served for hash (upload, else Gravatar), reusing a
// verdict recorded under its digest, and purges the CDN copy when banned.
func (uc *AvatarUsecase) Audit(ctx context.Context, hash, appID string) error {
	img, err := uc.auditSource(ctx, hash, appID)
	if err != nil {
		return fmt.Errorf("audit avatar %s: %w", hash, err)
	}

	imgHash := digest(img)
	if err = uc.store.WriteChecker(imgHash, img); err != nil {
		return fmt.Errorf("audit avatar %s: keep image %s: %w", hash, imgHash, err)
	}

	image, err := uc.images.Find(ctx, imgHash)
	switch {
	case errors.Is(err, rio.ErrNotFound):
		image = &Image{Hash: imgHash}
		image.Banned, image.Remark, err = uc.auditor.Check(ctx, uc.auditURL(hash, appID))
		if err != nil {
			return fmt.Errorf("audit avatar %s: check image %s: %w", hash, imgHash, err)
		}
		if err = uc.images.Create(ctx, image); err != nil {
			return fmt.Errorf("audit avatar %s: record image %s: %w", hash, imgHash, err)
		}
	case err != nil:
		return fmt.Errorf("audit avatar %s: find image %s: %w", hash, imgHash, err)
	}

	// Audit already runs on the queue, so it purges inline
	if image.Banned {
		if err = uc.purger.Refresh(ctx, []string{uc.avatarURL(hash)}); err != nil {
			return fmt.Errorf("audit avatar %s: purge cdn: %w", hash, err)
		}
	}
	return nil
}

func (uc *AvatarUsecase) auditSource(ctx context.Context, hash, appID string) ([]byte, error) {
	if avatar, err := uc.repo.FindForServe(ctx, hash, appID); err == nil {
		if img, _, err := uc.readUpload(avatar, appID); err == nil && len(img) > 0 {
			return img, nil
		}
	}

	img, _, err := uc.cached(ctx, kindGravatar, hash, uc.fetcher.Gravatar)
	return img, err
}

// enqueueAudit schedules Audit once per auditGuard window per hash, so a
// burst of requests for a new avatar cannot flood the queue.
func (uc *AvatarUsecase) enqueueAudit(ctx context.Context, hash, appID string) {
	key := "avatar:check:" + hash
	if !uc.cache.Add(key, true, auditGuard) {
		return
	}

	// the job outlives the request, whose buffers hash and appID may alias
	hash, appID = strings.Clone(hash), strings.Clone(appID)
	err := uc.queue.Push(func(ctx context.Context) error {
		return uc.Audit(ctx, hash, appID)
	})
	if err != nil {
		uc.cache.Forget(key)
		uc.log.WarnContext(ctx, "enqueue avatar audit failed",
			slog.String("hash", hash),
			slog.Any("err", err),
		)
	}
}

// enqueueRefresh purges the CDN in the background: CDN APIs can take seconds
// per driver, and the write has already succeeded.
func (uc *AvatarUsecase) enqueueRefresh(ctx context.Context, urls ...string) {
	err := uc.queue.Push(func(ctx context.Context) error {
		if err := uc.purger.Refresh(ctx, urls); err != nil {
			return fmt.Errorf("refresh cdn %v: %w", urls, err)
		}
		return nil
	})
	if err != nil {
		uc.log.ErrorContext(ctx, "enqueue cdn refresh failed",
			slog.Any("urls", urls),
			slog.Any("err", err),
		)
	}
}

func (uc *AvatarUsecase) avatarURL(hash string) string {
	return "https://" + uc.domain + "/avatar/" + hash
}

// auditURL carries appID so the moderator sees the image whose digest the
// verdict is recorded under.
func (uc *AvatarUsecase) auditURL(hash, appID string) string {
	u := uc.avatarURL(hash) + ".png?s=600&d=404"
	if appID != "" {
		u += "&app=" + url.QueryEscape(appID)
	}
	return u
}

// digest is the hex SHA256 of an image, the key of its audit verdict.
func digest(img []byte) string {
	sum := sha256.Sum256(img)
	return hex.EncodeToString(sum[:])
}
