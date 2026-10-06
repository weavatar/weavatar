package biz

import (
	"context"
	"errors"
	"log/slog"
	"math/rand/v2"
	"regexp"
	"strconv"
	"time"

	"github.com/go-rio/rio"
	"github.com/libtnb/utils/str"

	"github.com/weavatar/weavatar/pkg/embed"
	"github.com/weavatar/weavatar/pkg/imaging"
)

// Where a served avatar came from, sent as X-Avatar-From.
const (
	FromWeAvatar = "weavatar"
	FromGravatar = "gravatar"
	FromQQ       = "qq"
)

const (
	kindGravatar = "gravatar"
	kindQQ       = "qq"

	defaultExt  = "webp"
	defaultSeed = "weavatar" // keeps drawings stable when no hash was given
	maxSize     = 2048
)

var (
	hashPattern = regexp.MustCompile(`^(?:[a-f0-9]{64}|[a-f0-9]{32})$`)

	// zodiac stands in for initials when there is no name to take them from.
	zodiac = []string{"🐭", "🐮", "🐯", "🐰", "🐲", "🐍", "🐴", "🐏", "🐵", "🐔", "🐶", "🐷"}
)

// ResolveRequest is one avatar request after query normalization.
type ResolveRequest struct {
	Hash  string
	AppID string
	Ext   string // output format, e.g. webp or png
	Size  int    // output edge in pixels, 1 ~ 2048
	// Force skips every lookup and draws Default; an invalid Hash implies it.
	Force bool
	// Default applies when nothing is found: "404", a URL to redirect to, or
	// a Generator kind ("" draws the WeAvatar default).
	Default  string
	Name     string // initials source ahead of the owner's nickname
	Initials string
}

// ResolveResult is an image, a redirect, or not found.
type ResolveResult struct {
	Image        []byte
	LastModified time.Time
	From         string
	Redirect     string
	NotFound     bool
}

// ValidHash reports whether hash is a lowercase hex SHA256 or MD5 digest.
func ValidHash(hash string) bool {
	return hashPattern.MatchString(hash)
}

// Resolve tries a WeAvatar upload (app override first), Gravatar, then QQ.
// WeAvatar and Gravatar images audited as banned are swapped for the ban
// image, and queued for audit when not audited yet. With nothing found, or
// Force, Default decides. Images are re-encoded as Ext at Size.
func (uc *AvatarUsecase) Resolve(ctx context.Context, req ResolveRequest) (ResolveResult, error) {
	req = normalize(req)

	var (
		res   ResolveResult
		found bool
		owner string // user ID of the matching upload, for initials
	)

	// a forced default still looks the owner up when it needs their name
	needOwner := req.wantsInitials() && req.Initials == "" && req.Name == ""
	if ValidHash(req.Hash) && (!req.Force || needOwner) {
		avatar, err := uc.repo.FindForServe(ctx, req.Hash, req.AppID)
		switch {
		case err == nil:
			owner = avatar.UserID
			if !req.Force {
				res, found = uc.fromUpload(ctx, avatar, req.AppID)
			}
		case !errors.Is(err, rio.ErrNotFound):
			uc.log.WarnContext(ctx, "find avatar failed", slog.String("hash", req.Hash), slog.Any("err", err))
		}
	}
	if !found && !req.Force {
		res, found = uc.fromGravatar(ctx, req.Hash)
	}
	if !found && !req.Force {
		res, found = uc.fromQQ(ctx, req.Hash)
	}

	if found && res.From != FromQQ && uc.banned(ctx, req.Hash, req.AppID, res.Image) {
		img, err := embed.DefaultFS.ReadFile("default/ban.png")
		if err != nil {
			return ResolveResult{}, err
		}
		res.Image, res.LastModified = img, time.Now()
	}

	if !found || req.Force {
		return uc.fallback(ctx, req, owner)
	}
	return encode(res, req)
}

func (uc *AvatarUsecase) fromUpload(ctx context.Context, avatar *Avatar, appID string) (ResolveResult, bool) {
	img, modTime, err := uc.readUpload(avatar, appID)
	if err != nil || len(img) == 0 {
		uc.log.WarnContext(ctx, "read avatar failed", slog.String("hash", avatar.SHA256), slog.Any("err", err))
		return ResolveResult{}, false
	}

	return ResolveResult{Image: img, LastModified: modTime, From: FromWeAvatar}, true
}

func (uc *AvatarUsecase) readUpload(avatar *Avatar, appID string) ([]byte, time.Time, error) {
	if app := appOverride(avatar); app != nil {
		img, err := uc.store.ReadAppAvatar(appID, avatar.SHA256)
		return img, app.UpdatedAt, err
	}

	img, err := uc.store.ReadAvatar(avatar.SHA256)
	return img, avatar.UpdatedAt, err
}

func (uc *AvatarUsecase) fromGravatar(ctx context.Context, hash string) (ResolveResult, bool) {
	img, modTime, err := uc.cached(ctx, kindGravatar, hash, uc.fetcher.Gravatar)
	if err != nil || len(img) == 0 {
		return ResolveResult{}, false
	}

	return ResolveResult{Image: img, LastModified: modTime, From: FromGravatar}, true
}

func (uc *AvatarUsecase) fromQQ(ctx context.Context, hash string) (ResolveResult, bool) {
	qq, ok := uc.qq.Lookup(hash)
	if !ok {
		return ResolveResult{}, false
	}

	img, modTime, err := uc.cached(ctx, kindQQ, strconv.FormatUint(uint64(qq), 10), uc.fetcher.QQ)
	if err != nil || len(img) == 0 {
		return ResolveResult{}, false
	}

	return ResolveResult{Image: img, LastModified: modTime, From: FromQQ}, true
}

// cached serves a fetch from disk while younger than CacheTTL; a failed cache
// write must not cost the client the image it was fetched for.
func (uc *AvatarUsecase) cached(
	ctx context.Context,
	kind, key string,
	fetch func(ctx context.Context, key string) ([]byte, error),
) ([]byte, time.Time, error) {
	if img, modTime, ok := uc.store.ReadCache(kind, key); ok && time.Since(modTime) < CacheTTL {
		return img, modTime, nil
	}

	img, err := fetch(ctx, key)
	if err != nil {
		return nil, time.Time{}, err
	}
	if err = uc.store.WriteCache(kind, key, img); err != nil {
		uc.log.WarnContext(ctx, "cache avatar failed",
			slog.String("kind", kind),
			slog.String("key", key),
			slog.Any("err", err),
		)
	}

	return img, time.Now(), nil
}

// banned reports the audit verdict of img; without one it queues an audit and
// serves the image meanwhile.
func (uc *AvatarUsecase) banned(ctx context.Context, hash, appID string, img []byte) bool {
	image, err := uc.images.Find(ctx, digest(img))
	switch {
	case err == nil:
		return image.Banned
	case errors.Is(err, rio.ErrNotFound):
		uc.enqueueAudit(ctx, hash, appID)
	default:
		uc.log.WarnContext(ctx, "find image verdict failed", slog.String("hash", hash), slog.Any("err", err))
	}
	return false
}

func (uc *AvatarUsecase) fallback(ctx context.Context, req ResolveRequest, owner string) (ResolveResult, error) {
	switch {
	case req.Default == "404":
		return ResolveResult{NotFound: true}, nil
	case str.IsURL(req.Default):
		return ResolveResult{Redirect: req.Default}, nil
	}

	seed := req.Hash
	if seed == "" {
		seed = defaultSeed
	}
	text := ""
	if req.wantsInitials() {
		text = req.Initials
		if text == "" {
			// as on Gravatar, name gives the initial; the owner's nickname next
			name := req.Name
			if name == "" && owner != "" {
				name = uc.nickname(ctx, owner)
			}
			text = initial(name)
		}
	}

	img, err := uc.gen.Generate(req.Default, seed, req.Size, text)
	if err != nil {
		return ResolveResult{}, err
	}

	return encode(ResolveResult{Image: img, LastModified: time.Now(), From: FromWeAvatar}, req)
}

func (uc *AvatarUsecase) nickname(ctx context.Context, userID string) string {
	name, err := uc.users.Nickname(ctx, userID)
	if err != nil {
		uc.log.WarnContext(ctx, "find avatar owner failed", slog.String("user_id", userID), slog.Any("err", err))
		return ""
	}
	return name
}

// letter is the deprecated alias of initials.
func (r ResolveRequest) wantsInitials() bool {
	return r.Default == "initials" || r.Default == "letter"
}

func normalize(req ResolveRequest) ResolveRequest {
	if req.Ext == "" {
		req.Ext = defaultExt
	}
	req.Size = min(max(req.Size, 1), maxSize)
	if !ValidHash(req.Hash) {
		req.Force = true
	}
	return req
}

func appOverride(avatar *Avatar) *AppAvatar {
	if avatar.AppSHA256.Loaded() && avatar.AppSHA256.Row() != nil {
		return avatar.AppSHA256.Row()
	}
	if avatar.AppMD5.Loaded() && avatar.AppMD5.Row() != nil {
		return avatar.AppMD5.Row()
	}
	return nil
}

func initial(name string) string {
	for _, r := range name {
		return string(r)
	}
	return zodiac[rand.IntN(len(zodiac))] //nolint:gosec // cosmetic pick, not security
}

func encode(res ResolveResult, req ResolveRequest) (ResolveResult, error) {
	img, _, err := imaging.Decode(res.Image)
	if err != nil {
		return ResolveResult{}, err
	}
	if res.Image, err = imaging.Encode(imaging.Resize(img, req.Size, req.Size), req.Ext); err != nil {
		return ResolveResult{}, err
	}

	return res, nil
}
