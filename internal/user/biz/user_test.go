package biz_test

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-rio/rio"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"
	"github.com/samber/oops"

	mocksbiz "github.com/weavatar/weavatar/internal/mocks/user/biz"
	"github.com/weavatar/weavatar/internal/shared/apperr"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/internal/user/biz"
)

const (
	redirectURI = "https://weavatar.com/oauth/callback"
	issuedToken = "token-1"
)

// ttlCache records the lifetime each key is written with.
type ttlCache struct {
	cache.Cache
	ttls map[string]time.Duration
}

func (c *ttlCache) Put(key string, value any, ttl time.Duration) error {
	c.ttls[key] = ttl
	return c.Cache.Put(key, value, ttl)
}

type fixture struct {
	repo   *mocksbiz.UserRepo
	oauth  *mocksbiz.OAuthProvider
	tokens *mocksbiz.Tokens
	cache  *ttlCache
	uc     *biz.UserUsecase
}

func TestLoginURLCarriesAFreshState(t *testing.T) {
	f := newFixture()

	raw, err := f.uc.LoginURL(t.Context())
	must.NoError(t, err)

	u, err := url.Parse(raw)
	must.NoError(t, err)
	check.Equal(t, u.Scheme+"://"+u.Host+u.Path, "https://account.haozi.net/oauth/authorize")
	q := u.Query()
	check.Equal(t, q.Get("client_id"), "client-1")
	check.Equal(t, q.Get("redirect_uri"), redirectURI)
	check.Equal(t, q.Get("response_type"), "code")
	check.Equal(t, q.Get("scope"), "basic")
	check.Contains(t, raw, "redirect_uri="+url.QueryEscape(redirectURI))

	state := q.Get("state")
	check.True(t, strings.HasPrefix(state, "login-"))
	check.Len(t, state, len("login-")+16)
	check.True(t, f.cache.Has(state))
	check.Equal(t, f.cache.ttls[state], 5*time.Minute)
}

func TestCallbackRejectsExpiredState(t *testing.T) {
	f := newFixture()

	_, err := f.uc.Callback(t.Context(), "code", "login-missing")

	must.Error(t, err)
	check.Equal(t, apperr.KindOf(err), apperr.KindInvalid)
	check.Equal(t, apperr.CodeOf(err), "user.state_expired")
	check.Equal(t, oops.GetPublic(err, ""), "状态已过期")
}

func TestCallbackRejectsForeignCacheKey(t *testing.T) {
	f := newFixture()
	must.NoError(t, f.cache.Put("code:avatar:a@b.c", "123456", time.Minute))

	_, err := f.uc.Callback(t.Context(), "code", "code:avatar:a@b.c")

	check.Equal(t, apperr.CodeOf(err), "user.state_expired")
}

func TestCallbackStateIsSingleUse(t *testing.T) {
	f := newFixture()
	state := f.loginState(t)
	f.exchangeReturns(biz.Identity{UnionID: "union-1"})
	f.repo.FindByUnionIDFunc = func(context.Context, string) (*biz.User, error) {
		return &biz.User{ID: "u1", UnionID: "union-1"}, nil
	}
	f.issueTokens()
	_, err := f.uc.Callback(t.Context(), "code", state)
	must.NoError(t, err)

	_, err = f.uc.Callback(t.Context(), "code", state)

	check.Equal(t, apperr.CodeOf(err), "user.state_expired")
	check.Len(t, f.oauth.ExchangeCalls(), 1)
}

func TestCallbackCreatesUserOnFirstLogin(t *testing.T) {
	f := newFixture()
	state := f.loginState(t)
	f.exchangeReturns(biz.Identity{OpenID: "open-1", UnionID: "union-1", Nickname: "alice", RealName: true})
	f.repo.FindByUnionIDFunc = func(context.Context, string) (*biz.User, error) { return nil, notFound() }
	f.repo.CreateFunc = func(context.Context, *biz.User) error { return nil }
	f.issueTokens()

	token, err := f.uc.Callback(t.Context(), "code-1", state)

	must.NoError(t, err)
	check.Equal(t, token, issuedToken)

	exchanged := f.oauth.ExchangeCalls()
	must.Len(t, exchanged, 1)
	check.Equal(t, exchanged[0].Code, "code-1")
	check.Equal(t, exchanged[0].RedirectURI, redirectURI)

	created := f.repo.CreateCalls()
	must.Len(t, created, 1)
	user := created[0].User
	check.Len(t, user.ID, 10)
	check.Equal(t, user.OpenID, "open-1")
	check.Equal(t, user.UnionID, "union-1")
	check.Equal(t, user.Nickname, "新用户")
	check.Equal(t, user.Avatar, "https://weavatar.com/avatar/?d=mp")
	check.True(t, user.RealName)

	issued := f.tokens.IssueCalls()
	must.Len(t, issued, 1)
	check.Equal(t, issued[0].UserID, user.ID)
}

func TestCallbackSavesRealNameChange(t *testing.T) {
	f := newFixture()
	state := f.loginState(t)
	f.exchangeReturns(biz.Identity{OpenID: "open-1", UnionID: "union-1", RealName: true})
	f.repo.FindByUnionIDFunc = func(context.Context, string) (*biz.User, error) {
		return &biz.User{ID: "u1", UnionID: "union-1", Nickname: "alice", RealName: false}, nil
	}
	f.repo.UpdateFunc = func(context.Context, *biz.User) error { return nil }
	f.issueTokens()

	_, err := f.uc.Callback(t.Context(), "code", state)

	must.NoError(t, err)
	updated := f.repo.UpdateCalls()
	must.Len(t, updated, 1)
	check.Equal(t, updated[0].User.ID, "u1")
	check.True(t, updated[0].User.RealName)
	check.Equal(t, updated[0].User.Nickname, "alice")
	check.Equal(t, f.tokens.IssueCalls()[0].UserID, "u1")
}

func TestCallbackLeavesUnchangedUserUnwritten(t *testing.T) {
	f := newFixture()
	state := f.loginState(t)
	f.exchangeReturns(biz.Identity{UnionID: "union-1", RealName: true})
	f.repo.FindByUnionIDFunc = func(context.Context, string) (*biz.User, error) {
		return &biz.User{ID: "u1", UnionID: "union-1", RealName: true}, nil
	}
	f.issueTokens()

	token, err := f.uc.Callback(t.Context(), "code", state)

	must.NoError(t, err)
	check.Equal(t, token, issuedToken)
}

func TestCallbackUsesTheWinnerOfAFirstLoginRace(t *testing.T) {
	f := newFixture()
	state := f.loginState(t)
	f.exchangeReturns(biz.Identity{UnionID: "union-1"})
	lookups := 0
	f.repo.FindByUnionIDFunc = func(context.Context, string) (*biz.User, error) {
		lookups++
		if lookups == 1 {
			return nil, notFound()
		}
		return &biz.User{ID: "winner", UnionID: "union-1"}, nil
	}
	f.repo.CreateFunc = func(context.Context, *biz.User) error {
		return oops.In("user").Wrapf(rio.ErrDuplicateKey, "create user")
	}
	f.issueTokens()

	_, err := f.uc.Callback(t.Context(), "code", state)

	must.NoError(t, err)
	check.Equal(t, lookups, 2)
	check.Equal(t, f.tokens.IssueCalls()[0].UserID, "winner")
}

func TestCallbackPassesExchangeFailureThrough(t *testing.T) {
	f := newFixture()
	state := f.loginState(t)
	boom := errors.New("oauth down")
	f.oauth.ExchangeFunc = func(context.Context, string, string) (biz.Identity, error) {
		return biz.Identity{}, boom
	}

	_, err := f.uc.Callback(t.Context(), "code", state)

	must.ErrorIs(t, err, boom)
	check.Equal(t, apperr.KindOf(err), apperr.Kind(""))
}

func TestCallbackDoesNotTreatRepoFailureAsNewUser(t *testing.T) {
	f := newFixture()
	state := f.loginState(t)
	boom := errors.New("db down")
	f.exchangeReturns(biz.Identity{UnionID: "union-1"})
	f.repo.FindByUnionIDFunc = func(context.Context, string) (*biz.User, error) { return nil, boom }

	_, err := f.uc.Callback(t.Context(), "code", state)

	must.ErrorIs(t, err, boom)
}

func TestGetUnknownUserIsNotFound(t *testing.T) {
	f := newFixture()
	f.repo.FindFunc = func(context.Context, string) (*biz.User, error) { return nil, notFound() }

	_, err := f.uc.Get(t.Context(), "nobody")

	must.ErrorIs(t, err, rio.ErrNotFound)
}

func TestUpdateKeepsFieldsTheUserCannotEdit(t *testing.T) {
	f := newFixture()
	created := time.Date(2024, 8, 16, 16, 25, 20, 0, time.UTC)
	f.repo.FindFunc = func(_ context.Context, id string) (*biz.User, error) {
		return &biz.User{ID: id, Nickname: "old", Avatar: "https://a/1.png", RealName: true, CreatedAt: created}, nil
	}
	f.repo.UpdateFunc = func(context.Context, *biz.User) error { return nil }

	user, err := f.uc.Update(t.Context(), "u1", "new", "https://a/2.png")

	must.NoError(t, err)
	check.Equal(t, user.Nickname, "new")
	updated := f.repo.UpdateCalls()
	must.Len(t, updated, 1)
	check.Equal(t, updated[0].User.ID, "u1")
	check.Equal(t, updated[0].User.Nickname, "new")
	check.Equal(t, updated[0].User.Avatar, "https://a/2.png")
	// read-modify-write keeps the fields the user cannot edit
	check.True(t, updated[0].User.RealName)
	check.Equal(t, updated[0].User.CreatedAt, created)
}

func TestUpdateUnknownUserIsNotFound(t *testing.T) {
	f := newFixture()
	f.repo.FindFunc = func(context.Context, string) (*biz.User, error) { return nil, notFound() }

	_, err := f.uc.Update(t.Context(), "nobody", "new", "https://a/2.png")

	must.ErrorIs(t, err, rio.ErrNotFound)
}

func newFixture() *fixture {
	f := &fixture{
		repo:   &mocksbiz.UserRepo{},
		oauth:  &mocksbiz.OAuthProvider{},
		tokens: &mocksbiz.Tokens{},
		cache:  &ttlCache{Cache: cache.NewCache(cache.WithCleanupInterval(0)), ttls: map[string]time.Duration{}},
	}
	f.uc = biz.NewUserUsecase(f.repo, f.oauth, f.tokens, f.cache,
		appinfo.Domain("weavatar.com"),
		appinfo.OAuthClient{BaseURL: "https://account.haozi.net", ClientID: "client-1"},
	)
	return f
}

func (f *fixture) loginState(t *testing.T) string {
	t.Helper()

	raw, err := f.uc.LoginURL(t.Context())
	must.NoError(t, err)
	u, err := url.Parse(raw)
	must.NoError(t, err)
	return u.Query().Get("state")
}

func (f *fixture) exchangeReturns(identity biz.Identity) {
	f.oauth.ExchangeFunc = func(context.Context, string, string) (biz.Identity, error) {
		return identity, nil
	}
}

func (f *fixture) issueTokens() {
	f.tokens.IssueFunc = func(string) (string, error) { return issuedToken, nil }
}

// notFound wraps rio.ErrNotFound the way the data layer does.
func notFound() error {
	return oops.In("user").Wrapf(rio.ErrNotFound, "find user by union id")
}
