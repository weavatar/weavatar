package service_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/go-rio/rio"
	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"
	"github.com/libtnb/utils/jwt"

	mocksbiz "github.com/weavatar/weavatar/internal/mocks/user/biz"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/shared/transport"
	"github.com/weavatar/weavatar/internal/user/biz"
	"github.com/weavatar/weavatar/internal/user/service"
)

// envelope decodes the response wrapper; Code is the apperr key of an error.
type envelope struct {
	Msg  string          `json:"msg"`
	Code string          `json:"code"`
	Data json.RawMessage `json:"data"`
}

// harness serves the user routes, MustLogin included, against mocked ports;
// cleaned records the users handed to the account-deletion cleanup.
type harness struct {
	app     *fiber.App
	repo    *mocksbiz.UserRepo
	oauth   *mocksbiz.OAuthProvider
	tokens  *mocksbiz.Tokens
	cache   cache.Cache
	signer  *jwt.JWT
	cleaned []string
}

func TestLoginReturnsTheAuthorizationURL(t *testing.T) {
	h := newHarness(t)

	status, body := h.do(t, fiber.MethodGet, "/api/user/login", "", "")

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, body.Msg, "success")
	var login service.LoginURL
	must.NoError(t, json.Unmarshal(body.Data, &login))
	u, err := url.Parse(login.URL)
	must.NoError(t, err)
	check.Equal(t, u.Host, "account.haozi.net")
	check.True(t, h.cache.Has(u.Query().Get("state")))
}

func TestCallbackReturnsAToken(t *testing.T) {
	h := newHarness(t)
	must.NoError(t, h.cache.Put("login-abc", true, time.Minute))
	h.oauth.ExchangeFunc = func(context.Context, string, string) (biz.Identity, error) {
		return biz.Identity{UnionID: "union-1"}, nil
	}
	h.repo.FindByUnionIDFunc = func(context.Context, string) (*biz.User, error) {
		return &biz.User{ID: "u1", UnionID: "union-1"}, nil
	}
	h.tokens.IssueFunc = func(string) (string, error) { return "token-1", nil }

	status, body := h.do(t, fiber.MethodPost, "/api/user/callback", "", `{"code":"c","state":"login-abc"}`)

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, string(body.Data), `{"token":"token-1"}`)
	check.Equal(t, h.oauth.ExchangeCalls()[0].Code, "c")
}

func TestCallbackWithExpiredStateIsBadRequest(t *testing.T) {
	h := newHarness(t) // no mock funcs: the usecase must stop at the state

	status, body := h.do(t, fiber.MethodPost, "/api/user/callback", "", `{"code":"c","state":"login-gone"}`)

	check.Equal(t, status, fiber.StatusBadRequest)
	check.Equal(t, body.Msg, "状态已过期")
}

func TestCallbackWithoutCodeIsUnprocessable(t *testing.T) {
	h := newHarness(t)

	status, _ := h.do(t, fiber.MethodPost, "/api/user/callback", "", `{"state":"login-abc"}`)

	check.Equal(t, status, fiber.StatusUnprocessableEntity)
}

func TestSignedInRoutesRequireLogin(t *testing.T) {
	h := newHarness(t)

	for _, req := range []struct{ method, path, payload string }{
		{fiber.MethodGet, "/api/user/info", ""},
		{fiber.MethodPut, "/api/user/info", `{"nickname":"a","avatar":"https://a/b.png"}`},
		{fiber.MethodPost, "/api/user/logout", ""},
		{fiber.MethodGet, "/api/user/deletion/login", ""},
		{fiber.MethodPost, "/api/user/deletion/confirm", `{"code":"c","state":"delete-abc"}`},
	} {
		status, body := h.do(t, req.method, req.path, "", req.payload)
		check.Equal(t, status, fiber.StatusUnauthorized, req.method+" "+req.path)
		check.Equal(t, body.Msg, "未登录")
	}
}

func TestInfoHidesTheOAuthIdentifiers(t *testing.T) {
	h := newHarness(t)
	h.repo.FindFunc = func(_ context.Context, id string) (*biz.User, error) {
		return &biz.User{ID: id, OpenID: "secret-open-id", Nickname: "alice", RealName: true}, nil
	}

	status, body := h.do(t, fiber.MethodGet, "/api/user/info", h.login(t, "u1"), "")

	must.Equal(t, status, fiber.StatusOK)
	var info map[string]any
	must.NoError(t, json.Unmarshal(body.Data, &info))
	check.Equal(t, info["id"], any("u1"))
	check.Equal(t, info["nickname"], any("alice"))
	check.Equal(t, info["real_name"], any(true))
	for _, key := range []string{"avatar", "created_at", "updated_at"} {
		check.MapContainsKey(t, info, key)
	}
	// OAuth identifiers stay server-side
	check.MapNotContainsKey(t, info, "open_id")
	check.MapNotContainsKey(t, info, "union_id")
	check.Equal(t, h.repo.FindCalls()[0].ID, "u1")
}

func TestInfoOfDeletedUserIsNotFound(t *testing.T) {
	h := newHarness(t)
	h.repo.FindFunc = func(context.Context, string) (*biz.User, error) { return nil, rio.ErrNotFound }

	status, _ := h.do(t, fiber.MethodGet, "/api/user/info", h.login(t, "u1"), "")

	check.Equal(t, status, fiber.StatusNotFound)
}

func TestUpdateInfoSavesNicknameAndAvatar(t *testing.T) {
	h := newHarness(t)
	h.repo.FindFunc = func(_ context.Context, id string) (*biz.User, error) {
		return &biz.User{ID: id, Nickname: "old"}, nil
	}
	h.repo.UpdateFunc = func(context.Context, *biz.User) error { return nil }

	status, body := h.do(t, fiber.MethodPut, "/api/user/info", h.login(t, "u1"),
		`{"nickname":"alice","avatar":"https://weavatar.com/avatar/?d=mp"}`)

	must.Equal(t, status, fiber.StatusOK)
	var info map[string]any
	must.NoError(t, json.Unmarshal(body.Data, &info))
	check.Equal(t, info["nickname"], any("alice"))
	updated := h.repo.UpdateCalls()
	must.Len(t, updated, 1)
	check.Equal(t, updated[0].User.ID, "u1")
	check.Equal(t, updated[0].User.Nickname, "alice")
	check.Equal(t, updated[0].User.Avatar, "https://weavatar.com/avatar/?d=mp")
}

func TestUpdateInfoWithInvalidAvatarIsUnprocessable(t *testing.T) {
	h := newHarness(t) // no repo funcs: validation must fail first

	status, _ := h.do(t, fiber.MethodPut, "/api/user/info", h.login(t, "u1"),
		`{"nickname":"alice","avatar":"not a url"}`)

	check.Equal(t, status, fiber.StatusUnprocessableEntity)
}

func TestLogoutSucceeds(t *testing.T) {
	h := newHarness(t)

	status, body := h.do(t, fiber.MethodPost, "/api/user/logout", h.login(t, "u1"), "")

	check.Equal(t, status, fiber.StatusOK)
	check.Equal(t, body.Msg, "success")
}

func TestDeletionLoginReturnsAUserBoundState(t *testing.T) {
	h := newHarness(t)

	status, body := h.do(t, fiber.MethodGet, "/api/user/deletion/login", h.login(t, "u1"), "")

	must.Equal(t, status, fiber.StatusOK)
	var login service.LoginURL
	must.NoError(t, json.Unmarshal(body.Data, &login))
	u, err := url.Parse(login.URL)
	must.NoError(t, err)
	check.Equal(t, u.Host, "account.haozi.net")
	state := u.Query().Get("state")
	check.True(t, strings.HasPrefix(state, "delete-"))
	check.Equal(t, h.cache.Get(state), any("u1"))
}

func TestDeletionConfirmDeletesTheAccount(t *testing.T) {
	h := newHarness(t)
	must.NoError(t, h.cache.Put("delete-abc", "u1", time.Minute))
	h.repo.FindFunc = func(_ context.Context, id string) (*biz.User, error) {
		return &biz.User{ID: id, UnionID: "union-1"}, nil
	}
	h.oauth.ExchangeFunc = func(context.Context, string, string) (biz.Identity, error) {
		return biz.Identity{UnionID: "union-1"}, nil
	}
	h.repo.DeleteFunc = func(context.Context, *biz.User) error { return nil }

	status, body := h.do(t, fiber.MethodPost, "/api/user/deletion/confirm", h.login(t, "u1"), `{"code":"c","state":"delete-abc"}`)

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, body.Msg, "success")
	check.Equal(t, string(body.Data), "null")
	check.Equal(t, h.oauth.ExchangeCalls()[0].Code, "c")
	check.DeepEqual(t, h.cleaned, []string{"u1"})
	deleted := h.repo.DeleteCalls()
	must.Len(t, deleted, 1)
	check.Equal(t, deleted[0].User.ID, "u1")
}

func TestDeletionConfirmByAnotherAccountIsForbidden(t *testing.T) {
	h := newHarness(t)
	must.NoError(t, h.cache.Put("delete-abc", "u1", time.Minute))
	h.repo.FindFunc = func(_ context.Context, id string) (*biz.User, error) {
		return &biz.User{ID: id, UnionID: "union-1"}, nil
	}
	h.oauth.ExchangeFunc = func(context.Context, string, string) (biz.Identity, error) {
		return biz.Identity{UnionID: "union-2"}, nil
	}

	status, body := h.do(t, fiber.MethodPost, "/api/user/deletion/confirm", h.login(t, "u1"), `{"code":"c","state":"delete-abc"}`)

	check.Equal(t, status, fiber.StatusForbidden)
	check.Equal(t, body.Code, "user.deletion_identity_mismatch")
	check.Equal(t, body.Msg, "授权的账号与当前账号不一致")
	check.Len(t, h.cleaned, 0)
}

func TestDeletionConfirmWithExpiredStateIsBadRequest(t *testing.T) {
	h := newHarness(t) // no mock funcs: the usecase must stop at the state

	status, body := h.do(t, fiber.MethodPost, "/api/user/deletion/confirm", h.login(t, "u1"), `{"code":"c","state":"delete-gone"}`)

	check.Equal(t, status, fiber.StatusBadRequest)
	check.Equal(t, body.Code, "user.state_expired")
}

func TestDeletionConfirmWithoutStateIsUnprocessable(t *testing.T) {
	h := newHarness(t)

	status, _ := h.do(t, fiber.MethodPost, "/api/user/deletion/confirm", h.login(t, "u1"), `{"code":"c"}`)

	check.Equal(t, status, fiber.StatusUnprocessableEntity)
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	h := &harness{
		app:    fiber.New(),
		repo:   &mocksbiz.UserRepo{},
		oauth:  &mocksbiz.OAuthProvider{},
		tokens: &mocksbiz.Tokens{},
		cache:  cache.NewCache(cache.WithCleanupInterval(0)),
		signer: jwt.NewJWT("a-long-string-with-32-characters", time.Hour),
	}
	domain := appinfo.Domain("weavatar.com")
	client := appinfo.OAuthClient{BaseURL: "https://account.haozi.net", ClientID: "client-1"}
	tx := &mocksbiz.Transactor{RunFunc: func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }}
	cleanups := registry.UserCleanups{{Name: "test", Run: func(_ context.Context, userID string) error {
		h.cleaned = append(h.cleaned, userID)
		return nil
	}}}

	user := biz.NewUserUsecase(h.repo, h.oauth, h.tokens, h.cache, domain, client)
	deletion := biz.NewDeletionUsecase(h.repo, h.oauth, h.cache, tx, cleanups, domain, client)
	mount(h.app, service.UserRoutes(service.NewUserService(user, deletion, newValidator(t)), h.signer))

	return h
}

// do sends a JSON request, authenticated when token is set, and decodes the envelope.
func (h *harness) do(t *testing.T, method, path, token, payload string) (int, envelope) {
	t.Helper()

	var reader io.Reader
	if payload != "" {
		reader = strings.NewReader(payload)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	if token != "" {
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	}
	resp, err := h.app.Test(req)
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var body envelope
	must.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return resp.StatusCode, body
}

func (h *harness) login(t *testing.T, userID string) string {
	t.Helper()
	token, err := h.signer.Generate(&jwt.Claims{Subject: userID})
	must.NoError(t, err)
	return token
}

// mount registers endpoints the way the server does: middlewares, then handler.
func mount(app *fiber.App, endpoints transport.Endpoints) {
	for _, e := range endpoints {
		handlers := make([]any, 0, len(e.Middlewares)+1)
		for _, m := range e.Middlewares {
			handlers = append(handlers, m)
		}
		handlers = append(handlers, e.Handler)
		app.Add([]string{e.Method}, e.Path, handlers[0], handlers[1:]...)
	}
}
