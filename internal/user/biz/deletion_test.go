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
	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/user/biz"
)

// inTx marks the ctx the fake Transactor hands to the work it runs.
type inTx struct{}

type deletionFixture struct {
	repo    *mocksbiz.UserRepo
	oauth   *mocksbiz.OAuthProvider
	tx      *mocksbiz.Transactor
	cache   *ttlCache
	cleaned []string // userIDs the cleanups received, in order
	uc      *biz.DeletionUsecase
}

func TestDeletionURLCarriesAUserBoundState(t *testing.T) {
	f := newDeletionFixture(t, nil)

	raw, err := f.uc.DeletionURL(t.Context(), "u1")
	must.NoError(t, err)

	u, err := url.Parse(raw)
	must.NoError(t, err)
	check.Equal(t, u.Scheme+"://"+u.Host+u.Path, "https://account.haozi.net/oauth/authorize")
	q := u.Query()
	check.Equal(t, q.Get("client_id"), "client-1")
	check.Equal(t, q.Get("redirect_uri"), redirectURI) // the login callback is reused
	check.Equal(t, q.Get("response_type"), "code")
	check.Equal(t, q.Get("scope"), "basic")

	state := q.Get("state")
	check.True(t, strings.HasPrefix(state, "delete-"))
	check.Len(t, state, len("delete-")+16)
	check.Equal(t, f.cache.Get(state), any("u1"))
	check.Equal(t, f.cache.ttls[state], 5*time.Minute)
}

func TestConfirmDeletionRejectsExpiredState(t *testing.T) {
	f := newDeletionFixture(t, nil)

	err := f.uc.ConfirmDeletion(t.Context(), "u1", "code", "delete-missing")

	must.Error(t, err)
	check.Equal(t, apperr.KindOf(err), apperr.KindInvalid)
	check.Equal(t, apperr.CodeOf(err), "user.state_expired")
	check.Len(t, f.oauth.ExchangeCalls(), 0)
}

func TestConfirmDeletionRejectsLoginState(t *testing.T) {
	f := newDeletionFixture(t, nil)
	must.NoError(t, f.cache.Put("login-abc", true, time.Minute))

	err := f.uc.ConfirmDeletion(t.Context(), "u1", "code", "login-abc")

	check.Equal(t, apperr.CodeOf(err), "user.state_expired")
	check.True(t, f.cache.Has("login-abc")) // another feature's state is left alone
}

func TestConfirmDeletionRejectsAnotherUsersState(t *testing.T) {
	f := newDeletionFixture(t, nil)
	state := f.deletionState(t)

	err := f.uc.ConfirmDeletion(t.Context(), "u2", "code", state)

	check.Equal(t, apperr.CodeOf(err), "user.state_expired")
	check.Len(t, f.oauth.ExchangeCalls(), 0)
}

func TestConfirmDeletionStateIsSingleUse(t *testing.T) {
	f := newDeletionFixture(t, nil)
	state := f.deletionState(t)
	f.findReturns(&biz.User{ID: "u1", UnionID: "union-1"})
	f.exchangeReturns(biz.Identity{UnionID: "union-1"})
	f.repo.DeleteFunc = func(context.Context, *biz.User) error { return nil }
	must.NoError(t, f.uc.ConfirmDeletion(t.Context(), "u1", "code", state))

	err := f.uc.ConfirmDeletion(t.Context(), "u1", "code", state)

	check.Equal(t, apperr.CodeOf(err), "user.state_expired")
	check.Len(t, f.oauth.ExchangeCalls(), 1)
}

func TestConfirmDeletionOfUnknownUserIsNotFound(t *testing.T) {
	f := newDeletionFixture(t, nil)
	state := f.deletionState(t)
	f.repo.FindFunc = func(context.Context, string) (*biz.User, error) { return nil, notFound() }

	err := f.uc.ConfirmDeletion(t.Context(), "u1", "code", state)

	must.ErrorIs(t, err, rio.ErrNotFound)
	check.Len(t, f.oauth.ExchangeCalls(), 0)
}

func TestConfirmDeletionRejectsAnotherIdentity(t *testing.T) {
	f := newDeletionFixture(t, nil)
	state := f.deletionState(t)
	f.findReturns(&biz.User{ID: "u1", UnionID: "union-1"})
	f.exchangeReturns(biz.Identity{UnionID: "union-2"})

	err := f.uc.ConfirmDeletion(t.Context(), "u1", "code-1", state)

	must.Error(t, err)
	check.Equal(t, apperr.KindOf(err), apperr.KindForbidden)
	check.Equal(t, apperr.CodeOf(err), "user.deletion_identity_mismatch")
	check.Equal(t, oops.GetPublic(err, ""), "授权的账号与当前账号不一致")
	exchanged := f.oauth.ExchangeCalls()
	must.Len(t, exchanged, 1)
	check.Equal(t, exchanged[0].Code, "code-1")
	check.Equal(t, exchanged[0].RedirectURI, redirectURI)
	check.Len(t, f.tx.RunCalls(), 0)
	check.Len(t, f.cleaned, 0)
}

func TestConfirmDeletionRollsBackWhenACleanupFails(t *testing.T) {
	boom := errors.New("avatar store down")
	f := newDeletionFixture(t, []registry.UserCleanup{
		{Name: "avatar", Run: func(context.Context, string) error { return boom }},
		{Name: "later", Run: func(context.Context, string) error {
			t.Fatal("a cleanup after the failed one must not run")
			return nil
		}},
	})
	state := f.deletionState(t)
	f.findReturns(&biz.User{ID: "u1", UnionID: "union-1"})
	f.exchangeReturns(biz.Identity{UnionID: "union-1"})

	err := f.uc.ConfirmDeletion(t.Context(), "u1", "code", state)

	must.ErrorIs(t, err, boom)
	check.ErrorContains(t, err, "clean up avatar of user u1")
	check.Equal(t, apperr.KindOf(err), apperr.Kind(""))
	check.Len(t, f.tx.RunCalls(), 1) // the runner saw the error and rolled back
	check.Len(t, f.repo.DeleteCalls(), 0)
}

func TestConfirmDeletionRunsCleanupsThenSoftDeletes(t *testing.T) {
	f := newDeletionFixture(t, nil)
	state := f.deletionState(t)
	found := &biz.User{ID: "u1", UnionID: "union-1", Nickname: "alice", RealName: true}
	f.findReturns(found)
	f.exchangeReturns(biz.Identity{UnionID: "union-1"})
	f.repo.DeleteFunc = func(ctx context.Context, _ *biz.User) error {
		check.True(t, ctx.Value(inTx{}) == true, "delete must join the transaction")
		check.DeepEqual(t, f.cleaned, []string{"u1", "u1"}) // cleanups ran first
		return nil
	}

	must.NoError(t, f.uc.ConfirmDeletion(t.Context(), "u1", "code", state))

	check.Len(t, f.tx.RunCalls(), 1)
	deleted := f.repo.DeleteCalls()
	must.Len(t, deleted, 1)
	// the row goes to rio untouched: soft delete only stamps deleted_at
	check.True(t, deleted[0].User == found)
	check.Equal(t, deleted[0].User.Nickname, "alice")
	check.True(t, deleted[0].User.RealName)
}

// newDeletionFixture defaults cleanups to two recorders that require the
// transaction ctx.
func newDeletionFixture(t *testing.T, cleanups registry.UserCleanups) *deletionFixture {
	t.Helper()

	f := &deletionFixture{
		repo:  &mocksbiz.UserRepo{},
		oauth: &mocksbiz.OAuthProvider{},
		tx: &mocksbiz.Transactor{RunFunc: func(ctx context.Context, fn func(context.Context) error) error {
			return fn(context.WithValue(ctx, inTx{}, true))
		}},
		cache: &ttlCache{Cache: cache.NewCache(cache.WithCleanupInterval(0)), ttls: map[string]time.Duration{}},
	}
	if cleanups == nil {
		record := func(ctx context.Context, userID string) error {
			check.True(t, ctx.Value(inTx{}) == true, "cleanup must join the transaction")
			f.cleaned = append(f.cleaned, userID)
			return nil
		}
		cleanups = registry.UserCleanups{{Name: "first", Run: record}, {Name: "second", Run: record}}
	}
	f.uc = biz.NewDeletionUsecase(f.repo, f.oauth, f.cache, f.tx, cleanups,
		appinfo.Domain("weavatar.com"),
		appinfo.OAuthClient{BaseURL: "https://account.haozi.net", ClientID: "client-1"},
	)
	return f
}

// deletionState starts a deletion for u1 and returns its state.
func (f *deletionFixture) deletionState(t *testing.T) string {
	t.Helper()

	raw, err := f.uc.DeletionURL(t.Context(), "u1")
	must.NoError(t, err)
	u, err := url.Parse(raw)
	must.NoError(t, err)
	return u.Query().Get("state")
}

func (f *deletionFixture) findReturns(user *biz.User) {
	f.repo.FindFunc = func(context.Context, string) (*biz.User, error) { return user, nil }
}

func (f *deletionFixture) exchangeReturns(identity biz.Identity) {
	f.oauth.ExchangeFunc = func(context.Context, string, string) (biz.Identity, error) {
		return identity, nil
	}
}
