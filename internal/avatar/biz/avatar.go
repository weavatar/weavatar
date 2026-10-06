// Package biz holds the avatar module's business logic.
package biz

import (
	"context"
	"time"

	"github.com/go-rio/rio"
)

// Avatar binds an email or phone (Raw) to an uploaded image, addressed by the
// SHA256 or MD5 of Raw.
type Avatar struct {
	SHA256    string    `rio:",pk" json:"sha256"`
	MD5       string    `json:"md5"`
	Raw       string    `json:"raw"`
	UserID    string    `json:"user_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// per-app overrides; only loaded by AvatarRepo.FindForServe
	AppSHA256 rio.HasOne[AppAvatar] `rio:",fk:avatar_sha256,ref:sha256" json:"-"`
	AppMD5    rio.HasOne[AppAvatar] `rio:",fk:avatar_md5,ref:md5" json:"-"`
}

// App is a third-party application that may serve its own avatars.
type App struct {
	ID        string    `rio:",pk" json:"id"`
	UserID    string    `json:"user_id"`
	Name      string    `json:"name"`
	Secret    string    `json:"-"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AppAvatar marks an avatar as overridden for one app.
type AppAvatar struct {
	AvatarSHA256 string    `rio:",pk" json:"avatar_sha256"`
	AvatarMD5    string    `rio:",pk" json:"avatar_md5"`
	AppID        string    `rio:",pk" json:"app_id"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Image is the audit verdict for one image, keyed by the SHA256 of its bytes.
type Image struct {
	Hash      string    `rio:",pk" json:"hash"`
	Banned    bool      `json:"banned"`
	Remark    string    `json:"remark"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AvatarRepo is the avatar persistence boundary; a missing row is
// rio.ErrNotFound. Calls join the transaction a Transactor put in ctx.
type AvatarRepo interface {
	List(ctx context.Context, userID string, page, limit int) ([]*Avatar, int64, error)
	ListAllByUser(ctx context.Context, userID string) ([]*Avatar, error)
	// Find matches hash against sha256 or md5 among userID's avatars.
	Find(ctx context.Context, userID, hash string) (*Avatar, error)
	ExistsByRaw(ctx context.Context, raw string) (bool, error)
	// FindForServe matches hash against sha256 or md5 and loads AppSHA256 and
	// AppMD5 filtered to appID (loaded empty when appID is "").
	FindForServe(ctx context.Context, hash, appID string) (*Avatar, error)
	// Create reports a taken sha256 as ErrAvatarExists.
	Create(ctx context.Context, avatar *Avatar) error
	// Touch bumps UpdatedAt.
	Touch(ctx context.Context, avatar *Avatar) error
	Delete(ctx context.Context, avatar *Avatar) error
	// RandomHashes returns the sha256 of up to n random avatars.
	RandomHashes(ctx context.Context, n int) ([]string, error)
}

// ImageRepo stores audit verdicts; a missing row is rio.ErrNotFound.
type ImageRepo interface {
	Find(ctx context.Context, hash string) (*Image, error)
	Create(ctx context.Context, image *Image) error
}

// Transactor runs fn in one database transaction; repo and port calls made
// with fn's ctx join it.
type Transactor interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

// Users is the slice of the user module avatars need.
type Users interface {
	// Nickname returns "" for an unknown user.
	Nickname(ctx context.Context, userID string) (string, error)
}

// Store keeps image files: uploads, fetched-avatar caches and audit copies.
type Store interface {
	WriteAvatar(sha256 string, img []byte) error
	ReadAvatar(sha256 string) ([]byte, error)
	// RemoveAvatar succeeds when the file is already gone.
	RemoveAvatar(sha256 string) error
	ReadAppAvatar(appID, sha256 string) ([]byte, error)
	// ReadCache returns a cached fetch (kind "gravatar" or "qq") and its
	// modification time; freshness is the caller's call.
	ReadCache(kind, key string) (img []byte, modTime time.Time, ok bool)
	WriteCache(kind, key string, img []byte) error
	// WriteChecker keeps a copy of an image sent to audit.
	WriteChecker(imgHash string, img []byte) error
	// PurgeCache removes cached fetches not modified within olderThan.
	PurgeCache(ctx context.Context, olderThan time.Duration) error
}

// Fetcher downloads avatars from upstream services.
type Fetcher interface {
	Gravatar(ctx context.Context, hash string) ([]byte, error)
	QQ(ctx context.Context, qq string) ([]byte, error)
}

// QQHashes maps a QQ mailbox digest (md5 or sha256) to the QQ number.
type QQHashes interface {
	Lookup(hash string) (uint32, bool)
}

// Generator draws a default avatar of a Gravatar-style d kind, unknown kinds
// the WeAvatar default; size applies to identicon and color, text to initials.
type Generator interface {
	Generate(kind, seed string, size int, text string) ([]byte, error)
}

// Purger invalidates CDN caches.
type Purger interface {
	Refresh(ctx context.Context, urls []string) error
}

// Auditor runs AI moderation on a public image URL.
type Auditor interface {
	Check(ctx context.Context, url string) (banned bool, remark string, err error)
}
