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
	"github.com/weavatar/weavatar/internal/shared/transport"
	"github.com/weavatar/weavatar/internal/system/biz"
	"github.com/weavatar/weavatar/internal/system/service"
)

// harness serves the system routes against mocked ports.
type harness struct {
	app     *fiber.App
	usage   *mocksbiz.Usage
	avatars *mocksbiz.Avatars
}

func TestCountReturnsUsage(t *testing.T) {
	h := newHarness(t)
	h.usage.FetchFunc = func(context.Context, string, time.Time, time.Time) (uint, error) { return 7, nil }

	check.Equal(t, h.get(t, "/api/system/count"), `{"msg":"success","data":{"usage":7}}`)
}

func TestRandomAvatarsIsAnEmptyListOnFailure(t *testing.T) {
	h := newHarness(t)
	h.avatars.RandomHashesFunc = func(context.Context, int) ([]string, error) { return nil, errors.New("db down") }

	check.Equal(t, h.get(t, "/api/system/random_avatars"), `{"msg":"success","data":{"avatars":[]}}`)
}

// TestRoutesDocument guards router startup, which fails on a bad document.
func TestRoutesDocument(t *testing.T) {
	g, err := openapi.New("weavatar", "dev", openapi.WithValidator(validator.MustNew()))
	must.NoError(t, err)

	params := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	for _, e := range service.SystemRoutes(service.NewSystemService(nil)) {
		check.NoError(t, e.Document(g, e.Method, params.ReplaceAllString(e.Path, "{$1}"), e.Summary, e.Tags))
	}
	_, err = g.JSON()
	must.NoError(t, err)
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	h := &harness{
		app:     fiber.New(),
		usage:   &mocksbiz.Usage{},
		avatars: &mocksbiz.Avatars{},
	}
	uc := biz.NewSystemUsecase(h.usage, h.avatars, cache.NewCache(), "weavatar.com", slog.New(slog.DiscardHandler))
	mount(h.app, service.SystemRoutes(service.NewSystemService(uc)))

	return h
}

// get expects a 200 and returns the body.
func (h *harness) get(t *testing.T, target string) string {
	t.Helper()

	resp, err := h.app.Test(httptest.NewRequest(fiber.MethodGet, target, nil))
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	must.Equal(t, resp.StatusCode, fiber.StatusOK)
	body, err := io.ReadAll(resp.Body)
	must.NoError(t, err)
	return string(body)
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
