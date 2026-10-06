package service_test

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"image"
	"image/png"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/go-rio/rio"
	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"
	"github.com/libtnb/utils/jwt"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/avatar/service"
	mocksbiz "github.com/weavatar/weavatar/internal/mocks/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/rule"
	"github.com/weavatar/weavatar/pkg/queue"
)

const (
	md5Hash  = "0bc83cb571cd1c50ba6f3e8a78ef1346"
	testKey  = "a-long-string-with-32-characters"
	captcha  = `{"lot_number":"l","captcha_output":"o","pass_token":"p","gen_time":"1"}`
	takenRaw = "taken@example.com"
)

type env struct {
	app     *fiber.App
	token   string
	cache   cache.Cache
	repo    *mocksbiz.AvatarRepo
	images  *mocksbiz.ImageRepo
	store   *mocksbiz.Store
	fetcher *mocksbiz.Fetcher
	qq      *mocksbiz.QQHashes
	gen     *mocksbiz.Generator
}

// notExists stands in for the database-backed rule.
type notExists struct{}

func (notExists) Signature() string { return "not_exists" }

func (notExists) Message() string { return "{field} 已存在" }

func (notExists) Validate(f *validator.Field) (bool, error) {
	v, _ := f.Value[string]()
	return v != takenRaw, nil
}

type reply struct {
	StatusCode int
	Header     http.Header
}

// newEnv mounts the real routes, login middleware included, over mocked ports
// and a validator configured like production minus the database.
func newEnv(t *testing.T) *env {
	t.Helper()

	e := &env{
		cache:   cache.NewCache(),
		repo:    &mocksbiz.AvatarRepo{},
		images:  &mocksbiz.ImageRepo{},
		store:   &mocksbiz.Store{},
		fetcher: &mocksbiz.Fetcher{},
		qq:      &mocksbiz.QQHashes{},
		gen:     &mocksbiz.Generator{},
	}
	tx := &mocksbiz.TxRunner{RunFunc: func(ctx context.Context, fn func(context.Context) error) error { return fn(ctx) }}
	log := slog.New(slog.DiscardHandler)
	// the queue never starts and the purger has no func: purges stay queued
	uc := biz.NewAvatarUsecase(e.repo, e.images, tx, &mocksbiz.Users{}, e.store, e.fetcher, e.qq, e.gen,
		&mocksbiz.Purger{}, &mocksbiz.Auditor{}, queue.New(10, log), e.cache, "weavatar.com", log)

	v, err := validator.New(
		validator.WithTagNameFunc(fieldName),
		validator.WithStrictRequired(),
		validator.WithRules(rule.NewVerifyCode(e.cache)),
		validator.WithFallibleRules(notExists{}, rule.NewGeetest(nil, true)),
	)
	must.NoError(t, err)

	auth := jwt.NewJWT(testKey, time.Hour)
	e.token, err = auth.Generate(&jwt.Claims{Subject: "u1"})
	must.NoError(t, err)

	e.app = fiber.New()
	for _, ep := range service.AvatarRoutes(service.NewAvatarService(uc, v), auth) {
		handlers := make([]any, 0, len(ep.Middlewares)+1)
		for _, m := range ep.Middlewares {
			handlers = append(handlers, m)
		}
		handlers = append(handlers, ep.Handler)
		e.app.Add([]string{ep.Method}, ep.Path, handlers[0], handlers[1:]...)
	}

	return e
}

func TestAvatar_ServesImageWithCacheHeaders(t *testing.T) {
	e := newEnv(t)
	e.repo.FindForServeFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return &biz.Avatar{SHA256: md5Hash, UpdatedAt: time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)}, nil
	}
	e.store.ReadAvatarFunc = func(string) ([]byte, error) { return pngOf(t, 100), nil }
	e.images.FindFunc = func(context.Context, string) (*biz.Image, error) { return &biz.Image{}, nil }

	resp, body := e.do(t, httptest.NewRequest(fiber.MethodGet, "/api/avatar/"+strings.ToUpper(md5Hash)+".png?s=50", nil))

	must.Equal(t, resp.StatusCode, fiber.StatusOK)
	check.Equal(t, resp.Header.Get(fiber.HeaderContentType), "image/png")
	check.Equal(t, resp.Header.Get("X-Avatar-By"), "weavatar.com")
	check.Equal(t, resp.Header.Get("X-Avatar-From"), "weavatar")
	check.Equal(t, resp.Header.Get(fiber.HeaderCacheControl), "public, max-age=300")
	check.Equal(t, resp.Header.Get(fiber.HeaderLastModified), "Fri, 02 Jan 2026 03:04:05 GMT")
	check.NotEqual(t, resp.Header.Get(fiber.HeaderExpires), "")
	check.Equal(t, resp.Header.Get(fiber.HeaderVary), "Accept-Encoding, Accept")
	cfg, err := png.DecodeConfig(strings.NewReader(body))
	must.NoError(t, err)
	check.Equal(t, cfg.Width, 50)
	check.Equal(t, e.repo.FindForServeCalls()[0].Hash, md5Hash)
}

func TestAvatar_Default404(t *testing.T) {
	e := newEnv(t)
	e.nothingFound()

	resp, body := e.do(t, httptest.NewRequest(fiber.MethodGet, "/api/avatar/"+md5Hash+"?d=404", nil))

	check.Equal(t, resp.StatusCode, fiber.StatusNotFound)
	check.Equal(t, body, "404 Not Found\nWeAvatar")
}

func TestAvatar_DefaultURLRedirects(t *testing.T) {
	e := newEnv(t)
	e.nothingFound()

	resp, _ := e.do(t, httptest.NewRequest(fiber.MethodGet, "/api/avatar/"+md5Hash+"?default=https%3A%2F%2Fexample.com%2Fa.png", nil))

	check.Equal(t, resp.StatusCode, fiber.StatusFound)
	check.Equal(t, resp.Header.Get(fiber.HeaderLocation), "https://example.com/a.png")
}

func TestAvatar_NormalizesQuery(t *testing.T) {
	tests := []struct {
		name   string
		target string
		kind   string
		seed   string
		size   int
		text   string
	}{
		{name: "defaults", target: "/api/avatar/" + md5Hash + "?f=y", kind: "", seed: md5Hash, size: 80},
		{name: "size clamps and aliases", target: "/api/avatar/" + md5Hash + "?size=5000&forcedefault=yes&default=identicon", kind: "identicon", seed: md5Hash, size: 2048},
		{name: "unknown default is dropped", target: "/api/avatar/" + md5Hash + "?f=y&d=bogus&s=abc", kind: "", seed: md5Hash, size: 80},
		{name: "letter alias carries initials", target: "/api/avatar/" + md5Hash + "?f=y&d=letter&letter=wa", kind: "letter", seed: md5Hash, size: 80, text: "wa"},
		{name: "no hash forces the default", target: "/api/avatar?d=initials&name=Bob", kind: "initials", seed: "weavatar", size: 80, text: "B"},
		{name: "invalid hash forces the default", target: "/api/avatar/zz" + md5Hash + "?d=mp", kind: "mp", seed: "zz" + md5Hash, size: 80},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t) // lookup funcs stay nil: every case must be forced
			e.gen.GenerateFunc = func(string, string, int, string) ([]byte, error) { return pngOf(t, 40), nil }

			resp, _ := e.do(t, httptest.NewRequest(fiber.MethodGet, tt.target, nil))

			must.Equal(t, resp.StatusCode, fiber.StatusOK)
			check.Equal(t, resp.Header.Get(fiber.HeaderContentType), "image/webp")
			generated := e.gen.GenerateCalls()
			must.Len(t, generated, 1)
			check.Equal(t, generated[0].Kind, tt.kind)
			check.Equal(t, generated[0].Seed, tt.seed)
			check.Equal(t, generated[0].Size, tt.size)
			check.Equal(t, generated[0].Text, tt.text)
		})
	}
}

func TestAvatar_Head(t *testing.T) {
	e := newEnv(t)
	e.gen.GenerateFunc = func(string, string, int, string) ([]byte, error) { return pngOf(t, 40), nil }

	resp, body := e.do(t, httptest.NewRequest(fiber.MethodHead, "/api/avatar?d=mp", nil))

	check.Equal(t, resp.StatusCode, fiber.StatusOK)
	check.Equal(t, resp.Header.Get("X-Avatar-From"), "weavatar")
	check.Equal(t, body, "")
}

func TestAvatars_RequireLogin(t *testing.T) {
	e := newEnv(t)
	e.token = ""

	resp, body := e.do(t, httptest.NewRequest(fiber.MethodGet, "/api/avatars", nil))

	check.Equal(t, resp.StatusCode, fiber.StatusUnauthorized)
	check.Contains(t, body, "未登录")
}

func TestAvatarList(t *testing.T) {
	e := newEnv(t)
	e.repo.ListFunc = func(context.Context, string, int, int) ([]*biz.Avatar, int64, error) {
		return []*biz.Avatar{{SHA256: md5Hash, Raw: "a@example.com"}}, 1, nil
	}

	resp, body := e.do(t, httptest.NewRequest(fiber.MethodGet, "/api/avatars?page=2", nil))

	must.Equal(t, resp.StatusCode, fiber.StatusOK)
	check.Contains(t, body, `"total":1`)
	check.Contains(t, body, `"raw":"a@example.com"`)
	listed := e.repo.ListCalls()
	must.Len(t, listed, 1)
	check.Equal(t, listed[0].UserID, "u1")
	check.Equal(t, listed[0].Page, 2)
	check.Equal(t, listed[0].Limit, 10)
}

func TestAvatarCreate(t *testing.T) {
	e := newEnv(t)
	must.NoError(t, e.cache.Put("code:avatar:a@example.com", "123456", time.Minute))
	e.repo.CreateFunc = func(context.Context, *biz.Avatar) error { return nil }
	e.store.WriteAvatarFunc = func(string, []byte) error { return nil }

	resp, body := e.do(t, multipartRequest(t, fiber.MethodPost, "/api/avatars", map[string]string{
		"raw": "a@example.com", "verify_code": "123456", "captcha": captcha,
	}, pngOf(t, 100)))

	must.Equal(t, resp.StatusCode, fiber.StatusOK)
	check.Contains(t, body, `"raw":"a@example.com"`)
	created := e.repo.CreateCalls()
	must.Len(t, created, 1)
	check.Equal(t, created[0].Avatar.UserID, "u1")
	check.Equal(t, created[0].Avatar.Raw, "a@example.com")
}

func TestAvatarCreate_RejectsInvalidForm(t *testing.T) {
	tests := []struct {
		name   string
		fields map[string]string
		noFile bool
	}{
		{name: "wrong code", fields: map[string]string{"raw": "a@example.com", "verify_code": "000000", "captcha": captcha}},
		{name: "taken raw", fields: map[string]string{"raw": takenRaw, "verify_code": "123456", "captcha": captcha}},
		{name: "no captcha", fields: map[string]string{"raw": "a@example.com", "verify_code": "123456"}},
		{name: "no file", fields: map[string]string{"raw": "a@example.com", "verify_code": "123456", "captcha": captcha}, noFile: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv(t) // repo funcs stay nil: nothing may be created
			must.NoError(t, e.cache.Put("code:avatar:a@example.com", "123456", time.Minute))
			must.NoError(t, e.cache.Put("code:avatar:"+takenRaw, "123456", time.Minute))
			img := pngOf(t, 100)
			if tt.noFile {
				img = nil
			}

			resp, _ := e.do(t, multipartRequest(t, fiber.MethodPost, "/api/avatars", tt.fields, img))

			check.Equal(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
		})
	}
}

func TestAvatarUpdate(t *testing.T) {
	e := newEnv(t)
	e.repo.FindFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return &biz.Avatar{SHA256: md5Hash + md5Hash, MD5: md5Hash, UserID: "u1"}, nil
	}
	e.repo.TouchFunc = func(context.Context, *biz.Avatar) error { return nil }
	e.store.WriteAvatarFunc = func(string, []byte) error { return nil }

	resp, _ := e.do(t, multipartRequest(t, fiber.MethodPut, "/api/avatars/"+md5Hash, map[string]string{"captcha": captcha}, pngOf(t, 100)))

	must.Equal(t, resp.StatusCode, fiber.StatusOK)
	found := e.repo.FindCalls()
	must.Len(t, found, 1)
	check.Equal(t, found[0].UserID, "u1")
	check.Equal(t, found[0].Hash, md5Hash)
}

func TestAvatarUpdate_BadImageIsRejected(t *testing.T) {
	e := newEnv(t)
	e.repo.FindFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return &biz.Avatar{SHA256: md5Hash + md5Hash, MD5: md5Hash, UserID: "u1"}, nil
	}

	resp, body := e.do(t, multipartRequest(t, fiber.MethodPut, "/api/avatars/"+md5Hash, map[string]string{"captcha": captcha}, pngOf(t, 20)))

	check.Equal(t, resp.StatusCode, fiber.StatusBadRequest)
	check.Contains(t, body, "头像必须大于 40px")
}

func TestAvatarDelete_NotFound(t *testing.T) {
	e := newEnv(t)
	e.repo.FindFunc = func(context.Context, string, string) (*biz.Avatar, error) { return nil, rio.ErrNotFound }

	resp, _ := e.do(t, httptest.NewRequest(fiber.MethodDelete, "/api/avatars/"+md5Hash, nil))

	check.Equal(t, resp.StatusCode, fiber.StatusNotFound)
}

func TestAvatarCheck(t *testing.T) {
	e := newEnv(t)
	e.repo.ExistsByRawFunc = func(context.Context, string) (bool, error) { return true, nil }

	resp, body := e.do(t, httptest.NewRequest(fiber.MethodGet, "/api/avatars/check?raw=a%40example.com", nil))

	must.Equal(t, resp.StatusCode, fiber.StatusOK)
	check.Contains(t, body, `"bind":true`)
	check.Equal(t, e.repo.ExistsByRawCalls()[0].Raw, "a@example.com")
}

func TestAvatarQq(t *testing.T) {
	e := newEnv(t)
	img := pngOf(t, 100)
	e.fetcher.QQFunc = func(context.Context, string) ([]byte, error) { return img, nil }

	resp, body := e.do(t, httptest.NewRequest(fiber.MethodGet, "/api/avatars/qq?qq=10001", nil))

	must.Equal(t, resp.StatusCode, fiber.StatusOK)
	var got struct{ Data string }
	must.NoError(t, json.Unmarshal([]byte(body), &got))
	check.Equal(t, got.Data, base64.StdEncoding.EncodeToString(img))

	resp, _ = e.do(t, httptest.NewRequest(fiber.MethodGet, "/api/avatars/qq?qq=abc", nil))
	check.Equal(t, resp.StatusCode, fiber.StatusUnprocessableEntity)
}

// do sends req as the logged-in user unless req already carries a token.
func (e *env) do(t *testing.T, req *http.Request) (reply, string) {
	t.Helper()
	if req.Header.Get(fiber.HeaderAuthorization) == "" && e.token != "" {
		req.Header.Set(fiber.HeaderAuthorization, "Bearer "+e.token)
	}
	resp, err := e.app.Test(req, fiber.TestConfig{Timeout: 0}) // 2048px encodes outlast 1s under -race
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	must.NoError(t, err)
	return reply{StatusCode: resp.StatusCode, Header: resp.Header}, string(body)
}

// nothingFound makes every lookup of the avatar endpoint miss.
func (e *env) nothingFound() {
	e.repo.FindForServeFunc = func(context.Context, string, string) (*biz.Avatar, error) { return nil, rio.ErrNotFound }
	e.store.ReadCacheFunc = func(string, string) ([]byte, time.Time, bool) { return nil, time.Time{}, false }
	e.fetcher.GravatarFunc = func(context.Context, string) ([]byte, error) { return nil, io.EOF }
	e.qq.LookupFunc = func(string) (uint32, bool) { return 0, false }
}

func fieldName(field reflect.StructField) string {
	for _, tag := range []string{"form", "json", "query", "uri"} {
		if name, _, _ := strings.Cut(field.Tag.Get(tag), ","); name != "" && name != "-" {
			return name
		}
	}
	return field.Name
}

func pngOf(t *testing.T, size int) []byte {
	t.Helper()
	var buf bytes.Buffer
	must.NoError(t, png.Encode(&buf, image.NewRGBA(image.Rect(0, 0, size, size))))
	return buf.Bytes()
}

func multipartRequest(t *testing.T, method, target string, fields map[string]string, img []byte) *http.Request {
	t.Helper()
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	for k, v := range fields {
		must.NoError(t, w.WriteField(k, v))
	}
	if img != nil {
		f, err := w.CreateFormFile("avatar", "avatar.png")
		must.NoError(t, err)
		_, err = f.Write(img)
		must.NoError(t, err)
	}
	must.NoError(t, w.Close())

	req := httptest.NewRequest(method, target, &body)
	req.Header.Set(fiber.HeaderContentType, w.FormDataContentType())
	return req
}
