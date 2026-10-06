package service_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
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
	"github.com/weavatar/weavatar/internal/shared/transport"
	"github.com/weavatar/weavatar/internal/user/biz"
	"github.com/weavatar/weavatar/internal/user/service"
)

type testApp struct {
	app    *fiber.App
	repo   *mocksbiz.UserRepo
	oauth  *mocksbiz.OAuthProvider
	tokens *mocksbiz.Tokens
	cache  cache.Cache
	parser *jwt.JWT
}

func TestLogin(t *testing.T) {
	a := newTestApp(t)

	status, env := do[service.LoginURL](t, a, httptest.NewRequest(fiber.MethodGet, "/api/user/login", nil), "")

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, env.Msg, "success")
	u, err := url.Parse(env.Data.URL)
	must.NoError(t, err)
	check.Equal(t, u.Host, "account.haozi.net")
	check.True(t, a.cache.Has(u.Query().Get("state")))
}

func TestCallback(t *testing.T) {
	a := newTestApp(t)
	must.NoError(t, a.cache.Put("login-abc", true, time.Minute))
	a.oauth.ExchangeFunc = func(context.Context, string, string) (biz.Identity, error) {
		return biz.Identity{UnionID: "union-1"}, nil
	}
	a.repo.FindByUnionIDFunc = func(context.Context, string) (*biz.User, error) {
		return &biz.User{ID: "u1", UnionID: "union-1"}, nil
	}
	a.tokens.IssueFunc = func(string) (string, error) { return "token-1", nil }

	status, env := do[service.LoginToken](t, a,
		jsonRequest(fiber.MethodPost, "/api/user/callback", `{"code":"c","state":"login-abc"}`), "")

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, env.Data.Token, "token-1")
	check.Equal(t, a.oauth.ExchangeCalls()[0].Code, "c")
}

func TestCallback_StateExpiredIs400(t *testing.T) {
	a := newTestApp(t) // no mock funcs: the usecase must stop at the state

	status, env := do[any](t, a,
		jsonRequest(fiber.MethodPost, "/api/user/callback", `{"code":"c","state":"login-gone"}`), "")

	check.Equal(t, status, fiber.StatusBadRequest)
	check.Equal(t, env.Msg, "状态已过期")
}

func TestCallback_MissingCodeIs422(t *testing.T) {
	a := newTestApp(t)

	status, _ := do[any](t, a,
		jsonRequest(fiber.MethodPost, "/api/user/callback", `{"state":"login-abc"}`), "")

	check.Equal(t, status, fiber.StatusUnprocessableEntity)
}

func TestLoginRequired(t *testing.T) {
	a := newTestApp(t)

	for _, req := range []*http.Request{
		httptest.NewRequest(fiber.MethodGet, "/api/user/info", nil),
		jsonRequest(fiber.MethodPut, "/api/user/info", `{"nickname":"a","avatar":"https://a/b.png"}`),
		httptest.NewRequest(fiber.MethodPost, "/api/user/logout", nil),
	} {
		status, env := do[any](t, a, req, "")
		check.Equal(t, status, fiber.StatusUnauthorized, req.Method+" "+req.URL.Path)
		check.Equal(t, env.Msg, "未登录")
	}
}

func TestInfo(t *testing.T) {
	a := newTestApp(t)
	a.repo.FindFunc = func(_ context.Context, id string) (*biz.User, error) {
		return &biz.User{ID: id, OpenID: "secret-open-id", Nickname: "alice", RealName: true}, nil
	}

	status, env := do[map[string]any](t, a, httptest.NewRequest(fiber.MethodGet, "/api/user/info", nil), "u1")

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, env.Data["id"], any("u1"))
	check.Equal(t, env.Data["nickname"], any("alice"))
	check.Equal(t, env.Data["real_name"], any(true))
	for _, key := range []string{"avatar", "created_at", "updated_at"} {
		check.MapContainsKey(t, env.Data, key)
	}
	// OAuth identifiers stay server-side
	check.MapNotContainsKey(t, env.Data, "open_id")
	check.MapNotContainsKey(t, env.Data, "union_id")
	check.Equal(t, a.repo.FindCalls()[0].ID, "u1")
}

func TestInfo_DeletedUserIs404(t *testing.T) {
	a := newTestApp(t)
	a.repo.FindFunc = func(context.Context, string) (*biz.User, error) { return nil, rio.ErrNotFound }

	status, _ := do[any](t, a, httptest.NewRequest(fiber.MethodGet, "/api/user/info", nil), "u1")

	check.Equal(t, status, fiber.StatusNotFound)
}

func TestUpdateInfo(t *testing.T) {
	a := newTestApp(t)
	a.repo.FindFunc = func(_ context.Context, id string) (*biz.User, error) {
		return &biz.User{ID: id, Nickname: "old"}, nil
	}
	a.repo.UpdateFunc = func(context.Context, *biz.User) error { return nil }

	status, env := do[map[string]any](t, a, jsonRequest(fiber.MethodPut, "/api/user/info",
		`{"nickname":"alice","avatar":"https://weavatar.com/avatar/?d=mp"}`), "u1")

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, env.Data["nickname"], any("alice"))
	updated := a.repo.UpdateCalls()
	must.Len(t, updated, 1)
	check.Equal(t, updated[0].User.ID, "u1")
	check.Equal(t, updated[0].User.Nickname, "alice")
	check.Equal(t, updated[0].User.Avatar, "https://weavatar.com/avatar/?d=mp")
}

func TestUpdateInfo_InvalidAvatarIs422(t *testing.T) {
	a := newTestApp(t) // no repo funcs: validation must fail first

	status, _ := do[any](t, a, jsonRequest(fiber.MethodPut, "/api/user/info",
		`{"nickname":"alice","avatar":"not a url"}`), "u1")

	check.Equal(t, status, fiber.StatusUnprocessableEntity)
}

func TestLogout(t *testing.T) {
	a := newTestApp(t)

	status, env := do[any](t, a, httptest.NewRequest(fiber.MethodPost, "/api/user/logout", nil), "u1")

	check.Equal(t, status, fiber.StatusOK)
	check.Equal(t, env.Msg, "success")
}

// newTestApp mounts the real route table, MustLogin included, over mocked ports.
func newTestApp(t *testing.T) *testApp {
	t.Helper()

	a := &testApp{
		repo:   &mocksbiz.UserRepo{},
		oauth:  &mocksbiz.OAuthProvider{},
		tokens: &mocksbiz.Tokens{},
		cache:  cache.NewCache(cache.WithCleanupInterval(0)),
		parser: jwt.NewJWT("a-long-string-with-32-characters", time.Hour),
	}
	uc := biz.NewUserUsecase(a.repo, a.oauth, a.tokens, a.cache,
		appinfo.Domain("weavatar.com"),
		appinfo.OAuthClient{BaseURL: "https://account.haozi.net", ClientID: "client-1"},
	)
	user := service.NewUserService(uc, newValidator(t))

	a.app = fiber.New()
	for _, e := range service.UserRoutes(user, a.parser) {
		handlers := make([]any, 0, len(e.Middlewares)+1)
		for _, m := range e.Middlewares {
			handlers = append(handlers, m)
		}
		handlers = append(handlers, e.Handler)
		a.app.Add([]string{e.Method}, e.Path, handlers[0], handlers[1:]...)
	}

	return a
}

// do sends req, optionally as user, and decodes the envelope into data.
func do[T any](t *testing.T, a *testApp, req *http.Request, userID string) (int, transport.Envelope[T]) {
	t.Helper()

	if userID != "" {
		token, err := a.parser.Generate(&jwt.Claims{Subject: userID})
		must.NoError(t, err)
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+token)
	}
	resp, err := a.app.Test(req)
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	must.NoError(t, err)

	var env transport.Envelope[T]
	must.NoError(t, json.Unmarshal(body, &env), string(body))
	return resp.StatusCode, env
}

func jsonRequest(method, path, body string) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	return req
}
