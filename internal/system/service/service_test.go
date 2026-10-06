package service_test

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"
	"github.com/libtnb/validator"
	"github.com/libtnb/validator/contrib/openapi"

	mocksbiz "github.com/weavatar/weavatar/internal/mocks/system/biz"
	"github.com/weavatar/weavatar/internal/system/biz"
	"github.com/weavatar/weavatar/internal/system/service"
)

func TestCount(t *testing.T) {
	usage := &mocksbiz.Usage{
		FetchFunc: func(context.Context, string, time.Time, time.Time) (uint, error) { return 7, nil },
	}
	app := newTestApp(t, usage, &mocksbiz.Avatars{})

	check.Equal(t, get(t, app, "/api/system/count"), `{"msg":"success","data":{"usage":7}}`)
}

func TestRandomAvatars_EmptyListOnFailure(t *testing.T) {
	avatars := &mocksbiz.Avatars{
		RandomHashesFunc: func(context.Context, int) ([]string, error) { return nil, errors.New("db down") },
	}
	app := newTestApp(t, &mocksbiz.Usage{}, avatars)

	check.Equal(t, get(t, app, "/api/system/random_avatars"), `{"msg":"success","data":{"avatars":[]}}`)
}

// TestSystemRoutesDocument guards router startup, which fails on a bad document.
func TestSystemRoutesDocument(t *testing.T) {
	g, err := openapi.New("weavatar", "dev", openapi.WithValidator(validator.MustNew()))
	must.NoError(t, err)

	params := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	for _, e := range service.SystemRoutes(service.NewSystemService(nil)) {
		check.NoError(t, e.Document(g, e.Method, params.ReplaceAllString(e.Path, "{$1}"), e.Summary, e.Tags))
	}
	_, err = g.JSON()
	must.NoError(t, err)
}

func newTestApp(t *testing.T, usage *mocksbiz.Usage, avatars *mocksbiz.Avatars) *fiber.App {
	t.Helper()

	uc := biz.NewSystemUsecase(usage, avatars, cache.NewCache(), "weavatar.com", slog.New(slog.DiscardHandler))
	app := fiber.New()
	for _, e := range service.SystemRoutes(service.NewSystemService(uc)) {
		app.Add([]string{e.Method}, e.Path, e.Handler)
	}
	return app
}

func get(t *testing.T, app *fiber.App, target string) string {
	t.Helper()
	resp, err := app.Test(httptest.NewRequest(fiber.MethodGet, target, nil))
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	must.Equal(t, resp.StatusCode, fiber.StatusOK)
	body, err := io.ReadAll(resp.Body)
	must.NoError(t, err)
	return string(body)
}
