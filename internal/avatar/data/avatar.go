package data

import (
	"context"
	"errors"

	"github.com/go-rio/rio"
	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/database"
)

var (
	avatarsByUser        = rio.From[biz.Avatar]().Where("user_id = ?").Must()
	orderedAvatarsByUser = rio.From[biz.Avatar]().Where("user_id = ?").OrderBy("created_at DESC").Must()
	userAvatarByHash     = rio.From[biz.Avatar]().Where("(sha256 = ? OR md5 = ?) AND user_id = ?").Must()
	avatarByRaw          = rio.From[biz.Avatar]().Where("raw = ?").Must()
	avatarByHash         = rio.From[biz.Avatar]().Where("sha256 = ? OR md5 = ?").Must()
	randomAvatars        = rio.From[biz.Avatar]().OrderBy("random()").Must()
)

type avatarRepo struct {
	db *rio.DB
}

func NewAvatarRepo(db *rio.DB) biz.AvatarRepo {
	return &avatarRepo{db: db}
}

func (r *avatarRepo) List(ctx context.Context, userID string, page, limit int) ([]*biz.Avatar, int64, error) {
	if page < 1 { // guard the callers that skip HTTP validation
		page = 1
	}
	q := database.Q(ctx, r.db)
	total, err := avatarsByUser.Count(ctx, q, userID)
	if err != nil {
		return nil, 0, wrap(err, "count avatars of %s", userID)
	}

	list, err := orderedAvatarsByUser.Offset((page-1)*limit).Limit(limit).All(ctx, q, userID)
	if err != nil {
		return nil, 0, wrap(err, "list avatars of %s", userID)
	}

	avatars := make([]*biz.Avatar, len(list))
	for i := range list {
		avatars[i] = &list[i]
	}
	return avatars, total, nil
}

func (r *avatarRepo) ListAllByUser(ctx context.Context, userID string) ([]*biz.Avatar, error) {
	list, err := avatarsByUser.All(ctx, database.Q(ctx, r.db), userID)
	if err != nil {
		return nil, wrap(err, "list all avatars of %s", userID)
	}

	avatars := make([]*biz.Avatar, len(list))
	for i := range list {
		avatars[i] = &list[i]
	}
	return avatars, nil
}

func (r *avatarRepo) Find(ctx context.Context, userID, hash string) (*biz.Avatar, error) {
	avatar, err := userAvatarByHash.First(ctx, database.Q(ctx, r.db), hash, hash, userID)
	if err != nil {
		return nil, wrap(err, "find avatar %s of %s", hash, userID)
	}
	return avatar, nil
}

func (r *avatarRepo) ExistsByRaw(ctx context.Context, raw string) (bool, error) {
	exists, err := avatarByRaw.Exists(ctx, database.Q(ctx, r.db), raw)
	if err != nil {
		return false, wrap(err, "check avatar by raw")
	}
	return exists, nil
}

func (r *avatarRepo) FindForServe(ctx context.Context, hash, appID string) (*biz.Avatar, error) {
	query := avatarByHash
	if appID != "" {
		query = query.
			With("AppSHA256", rio.RelWhere("app_id = ?", appID)).
			With("AppMD5", rio.RelWhere("app_id = ?", appID))
	}

	avatar, err := query.First(ctx, database.Q(ctx, r.db), hash, hash)
	if err != nil {
		return nil, wrap(err, "find avatar %s", hash)
	}
	if appID == "" { // no app can override it
		avatar.AppSHA256.Set(nil)
		avatar.AppMD5.Set(nil)
	}
	return avatar, nil
}

func (r *avatarRepo) Create(ctx context.Context, avatar *biz.Avatar) error {
	if err := rio.Insert(ctx, database.Q(ctx, r.db), avatar); err != nil {
		if errors.Is(err, rio.ErrDuplicateKey) {
			return biz.ErrAvatarExists
		}
		return wrap(err, "create avatar %s", avatar.SHA256)
	}
	return nil
}

func (r *avatarRepo) Touch(ctx context.Context, avatar *biz.Avatar) error {
	if err := rio.Update(ctx, database.Q(ctx, r.db), avatar); err != nil {
		return wrap(err, "touch avatar %s", avatar.SHA256)
	}
	return nil
}

func (r *avatarRepo) Delete(ctx context.Context, avatar *biz.Avatar) error {
	if err := rio.Delete(ctx, database.Q(ctx, r.db), avatar); err != nil {
		return wrap(err, "delete avatar %s", avatar.SHA256)
	}
	return nil
}

func (r *avatarRepo) RandomHashes(ctx context.Context, n int) ([]string, error) {
	hashes, err := randomAvatars.Limit(n).Pluck[string](ctx, database.Q(ctx, r.db), "sha256")
	if err != nil {
		return nil, wrap(err, "pick random avatars")
	}
	return hashes, nil
}

// wrap leaves rio.ErrNotFound bare: misses are the hot path of avatar
// serving, and a stack trace per miss is wasted work.
func wrap(err error, format string, args ...any) error {
	if errors.Is(err, rio.ErrNotFound) {
		return err
	}
	return oops.In("avatar").Wrapf(err, format, args...)
}
