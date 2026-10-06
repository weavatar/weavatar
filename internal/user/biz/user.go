// Package biz holds the user module's business logic.
package biz

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/go-rio/rio"
	"github.com/libtnb/cache"
	"github.com/libtnb/utils/str"
	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/shared/apperr"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/pkg/id"
)

const (
	// statePrefix keeps a client from passing another cache key off as a state.
	statePrefix = "login-"
	stateTTL    = 5 * time.Minute

	defaultNickname = "新用户"
)

type User struct {
	ID        string     `rio:",pk" json:"id"`
	OpenID    string     `json:"-"`
	UnionID   string     `json:"-"`
	Nickname  string     `json:"nickname"`
	Avatar    string     `json:"avatar"`
	RealName  bool       `json:"real_name"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `rio:",softdelete" json:"-"`
}

// UserRepo reports a miss as rio.ErrNotFound and a taken union_id as
// rio.ErrDuplicateKey; calls join the transaction a Transactor put in ctx.
type UserRepo interface {
	FindByUnionID(ctx context.Context, unionID string) (*User, error)
	Find(ctx context.Context, id string) (*User, error)
	Create(ctx context.Context, user *User) error
	Update(ctx context.Context, user *User) error
	// Delete soft-deletes: the row keeps every column and gains deleted_at.
	Delete(ctx context.Context, user *User) error
}

// Transactor runs fn in one database transaction; repo and cleanup calls
// made with fn's ctx join it.
type Transactor interface {
	Run(ctx context.Context, fn func(ctx context.Context) error) error
}

// Identity is the account the OAuth server vouches for.
type Identity struct {
	OpenID   string
	UnionID  string
	Nickname string
	RealName bool
}

// OAuthProvider trades an authorization code for the user's identity.
type OAuthProvider interface {
	Exchange(ctx context.Context, code, redirectURI string) (Identity, error)
}

// Tokens issues login tokens.
type Tokens interface {
	Issue(userID string) (string, error)
}

// UserUsecase is shared by HTTP and other modules.
type UserUsecase struct {
	repo   UserRepo
	oauth  OAuthProvider
	tokens Tokens
	cache  cache.Cache
	domain string
	client appinfo.OAuthClient
}

func NewUserUsecase(
	repo UserRepo,
	oauth OAuthProvider,
	tokens Tokens,
	c cache.Cache,
	domain appinfo.Domain,
	client appinfo.OAuthClient,
) *UserUsecase {
	return &UserUsecase{
		repo:   repo,
		oauth:  oauth,
		tokens: tokens,
		cache:  c,
		domain: string(domain),
		client: client,
	}
}

// LoginURL returns the authorization URL for a fresh login state.
func (uc *UserUsecase) LoginURL(_ context.Context) (string, error) {
	state := statePrefix + str.Random(16)
	if err := uc.cache.Put(state, true, stateTTL); err != nil {
		return "", oops.In("user").Wrapf(err, "store login state")
	}

	return authorizeURL(uc.client, redirectURI(uc.domain), state), nil
}

// Callback redeems a one-shot state and code for a login token, creating the
// user on first login and refreshing the real-name flag afterwards.
func (uc *UserUsecase) Callback(ctx context.Context, code, state string) (string, error) {
	if !strings.HasPrefix(state, statePrefix) || uc.cache.Pull(state) == nil {
		return "", errStateExpired()
	}

	identity, err := uc.oauth.Exchange(ctx, code, redirectURI(uc.domain))
	if err != nil {
		return "", err
	}

	user, err := uc.login(ctx, identity)
	if err != nil {
		return "", err
	}

	token, err := uc.tokens.Issue(user.ID)
	if err != nil {
		return "", oops.In("user").Wrapf(err, "issue token for user %s", user.ID)
	}

	return token, nil
}

func (uc *UserUsecase) Get(ctx context.Context, id string) (*User, error) {
	return uc.repo.Find(ctx, id)
}

// Update changes the profile fields a user may edit.
func (uc *UserUsecase) Update(ctx context.Context, id, nickname, avatar string) (*User, error) {
	user, err := uc.repo.Find(ctx, id)
	if err != nil {
		return nil, err
	}

	user.Nickname = nickname
	user.Avatar = avatar
	if err = uc.repo.Update(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// login finds the user behind identity, creating it on first sight.
func (uc *UserUsecase) login(ctx context.Context, identity Identity) (*User, error) {
	user, err := uc.repo.FindByUnionID(ctx, identity.UnionID)
	if err == nil {
		if user.RealName != identity.RealName {
			user.RealName = identity.RealName
			if err = uc.repo.Update(ctx, user); err != nil {
				return nil, err
			}
		}
		return user, nil
	}
	if !errors.Is(err, rio.ErrNotFound) {
		return nil, err
	}

	user = &User{
		ID:       id.Generate(),
		OpenID:   identity.OpenID,
		UnionID:  identity.UnionID,
		Nickname: defaultNickname,
		Avatar:   "https://" + uc.domain + "/avatar/?d=mp",
		RealName: identity.RealName,
	}
	if err = uc.repo.Create(ctx, user); err != nil {
		// a concurrent first login won the insert: use its row
		if errors.Is(err, rio.ErrDuplicateKey) {
			return uc.repo.FindByUnionID(ctx, identity.UnionID)
		}
		return nil, err
	}

	return user, nil
}

func errStateExpired() error {
	return apperr.Invalid("user.state_expired", "状态已过期").In("user").New("login state expired")
}

// authorizeURL is the OAuth server's authorization page for one state.
func authorizeURL(client appinfo.OAuthClient, redirectURI, state string) string {
	return client.BaseURL + "/oauth/authorize?client_id=" + client.ClientID +
		"&redirect_uri=" + url.QueryEscape(redirectURI) +
		"&response_type=code&scope=basic&state=" + state
}

// redirectURI is the one callback registered with the OAuth server; logins
// and deletion confirmations share it.
func redirectURI(domain string) string {
	return "https://" + domain + "/oauth/callback"
}
