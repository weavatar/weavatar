package server

import (
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/platform/conf"
	"github.com/weavatar/weavatar/internal/shared/registry"
	"github.com/weavatar/weavatar/internal/shared/transport"
)

type response struct {
	status int
	header http.Header
	body   string
}

func TestEndpointMiddlewaresRunBeforeHandler(t *testing.T) {
	var order []string
	step := func(name string) fiber.Handler {
		return func(c fiber.Ctx) error {
			order = append(order, name)
			return c.Next()
		}
	}
	deny := func(c fiber.Ctx) error {
		return transport.Error(c, fiber.StatusUnauthorized, "未登录")
	}
	routes := registry.Routes{{
		{Method: fiber.MethodGet, Path: "/open", Handler: func(c fiber.Ctx) error {
			order = append(order, "handler")
			return c.SendString("ok")
		}, Middlewares: []fiber.Handler{step("first"), step("second")}},
		{Method: fiber.MethodGet, Path: "/closed", Handler: func(c fiber.Ctx) error {
			order = append(order, "unreachable")
			return c.SendString("ok")
		}, Middlewares: []fiber.Handler{deny}},
	}}

	app, err := NewRouter(testConfig(), validator.MustNew(), "test", routes)
	must.NoError(t, err)

	resp := send(t, app, httptest.NewRequest(http.MethodGet, "/open", nil))
	must.Equal(t, resp.status, fiber.StatusOK)
	must.Equal(t, resp.body, "ok")
	must.DeepEqual(t, order, []string{"first", "second", "handler"})

	order = nil
	resp = send(t, app, httptest.NewRequest(http.MethodGet, "/closed", nil))
	must.Equal(t, resp.status, fiber.StatusUnauthorized)
	must.Contains(t, resp.body, "未登录")
	must.Len(t, order, 0)
}

func TestRootRoutesRedirectToWebsite(t *testing.T) {
	config := testConfig()
	app, err := NewRouter(config, validator.MustNew(), "test", registry.Routes{RootRoutes(config)})
	must.NoError(t, err)

	for _, path := range []string{"/", "/api", "/api/"} {
		resp := send(t, app, httptest.NewRequest(http.MethodGet, path, nil))
		must.Equal(t, resp.status, fiber.StatusFound, must.Msgf("path %s", path))
		must.Equal(t, resp.header.Get(fiber.HeaderLocation), "https://weavatar.test")
	}
}

func TestGlobalMiddlewares(t *testing.T) {
	routes := registry.Routes{{
		{Method: fiber.MethodGet, Path: "/ping", Handler: func(c fiber.Ctx) error {
			return c.SendString("pong")
		}},
	}}
	request := func(t *testing.T, origins []string, origin string) response {
		t.Helper()
		config := testConfig()
		config.HTTP.CorsOrigins = origins
		app, err := NewRouter(config, validator.MustNew(), "test", routes)
		must.NoError(t, err)

		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		req.Header.Set(fiber.HeaderOrigin, origin)
		resp := send(t, app, req)
		must.Equal(t, resp.status, fiber.StatusOK)
		must.NotEmpty(t, resp.header.Get(fiber.HeaderXRequestID))
		// no helmet: its Cross-Origin-Resource-Policy would block cross-site embedding
		must.Empty(t, resp.header.Get("Cross-Origin-Resource-Policy"))
		return resp
	}

	t.Run("same-origin by default", func(t *testing.T) {
		resp := request(t, nil, "https://evil.example")
		must.Empty(t, resp.header.Get(fiber.HeaderAccessControlAllowOrigin))
	})

	t.Run("configured origin", func(t *testing.T) {
		resp := request(t, []string{"https://blog.example"}, "https://blog.example")
		must.Equal(t, resp.header.Get(fiber.HeaderAccessControlAllowOrigin), "https://blog.example")
	})

	t.Run("any origin", func(t *testing.T) {
		resp := request(t, []string{"*"}, "https://blog.example")
		must.Equal(t, resp.header.Get(fiber.HeaderAccessControlAllowOrigin), "*")
	})
}

func TestDocsServeOpenAPI(t *testing.T) {
	config := testConfig()
	config.HTTP.Docs = true
	app, err := NewRouter(config, validator.MustNew(), "v1.2.3", nil)
	must.NoError(t, err)

	resp := send(t, app, httptest.NewRequest(http.MethodGet, "/openapi.json", nil))
	must.Equal(t, resp.status, fiber.StatusOK)
	must.Contains(t, resp.body, `"title": "WeAvatar"`)
	must.Contains(t, resp.body, `"version": "v1.2.3"`)

	resp = send(t, app, httptest.NewRequest(http.MethodGet, "/docs", nil))
	must.Equal(t, resp.status, fiber.StatusOK)
}

func TestFrameworkErrorsAnswerInEnvelope(t *testing.T) {
	routes := registry.Routes{{
		{Method: fiber.MethodGet, Path: "/panic", Handler: func(fiber.Ctx) error {
			panic("secret backend detail")
		}},
	}}
	app, err := NewRouter(testConfig(), validator.MustNew(), "test", routes)
	must.NoError(t, err)

	for _, tc := range []struct {
		method, path string
		status       int
		body         string
	}{
		{http.MethodGet, "/missing", fiber.StatusNotFound, `{"msg":"Not Found","data":null}`},
		{http.MethodPost, "/panic", fiber.StatusMethodNotAllowed, `{"msg":"Method Not Allowed","data":null}`},
		{http.MethodGet, "/panic", fiber.StatusInternalServerError, `{"msg":"Internal Server Error","data":null}`},
	} {
		resp := send(t, app, httptest.NewRequest(tc.method, tc.path, nil))
		must.Equal(t, resp.status, tc.status, must.Msgf("%s %s", tc.method, tc.path))
		must.Equal(t, resp.body, tc.body)
	}
}

func TestClientIPFollowsTheProxyHeader(t *testing.T) {
	for name, tc := range map[string]struct {
		proxyHeader string
		realIP      string
		want        string
	}{
		"header configured":         {"X-Real-IP", "203.0.113.7", "203.0.113.7"},
		"no header configured":      {"", "203.0.113.7", "127.0.0.1"},
		"invalid header falls back": {"X-Real-IP", "not-an-ip", "127.0.0.1"},
	} {
		t.Run(name, func(t *testing.T) {
			must.Equal(t, clientIP(t, tc.proxyHeader, tc.realIP), tc.want)
		})
	}
}

func testConfig() *conf.Config {
	return &conf.Config{
		App:  conf.App{Name: "WeAvatar"},
		HTTP: conf.HTTP{Domain: "weavatar.test", BodyLimit: 4096, HeaderLimit: 4096},
	}
}

func send(t *testing.T, app *fiber.App, req *http.Request) response {
	t.Helper()
	resp, err := app.Test(req)
	must.NoError(t, err)
	body, err := io.ReadAll(resp.Body)
	must.NoError(t, err)
	must.NoError(t, resp.Body.Close())
	return response{status: resp.StatusCode, header: resp.Header, body: string(body)}
}

// clientIP serves c.IP() on a loopback listener, so the peer is 127.0.0.1,
// and sends realIP in X-Real-IP.
func clientIP(t *testing.T, proxyHeader, realIP string) string {
	t.Helper()
	config := testConfig()
	config.HTTP.ProxyHeader = proxyHeader
	app, err := NewRouter(config, validator.MustNew(), "test", registry.Routes{{
		{Method: fiber.MethodGet, Path: "/ip", Handler: func(c fiber.Ctx) error {
			return c.SendString(c.IP())
		}},
	}})
	must.NoError(t, err)

	ln, err := new(net.ListenConfig).Listen(t.Context(), "tcp", "127.0.0.1:0")
	must.NoError(t, err)
	go func() { _ = app.Listener(ln, fiber.ListenConfig{DisableStartupMessage: true}) }()
	t.Cleanup(func() { _ = app.Shutdown() })

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, "http://"+ln.Addr().String()+"/ip", nil)
	must.NoError(t, err)
	req.Header.Set("X-Real-IP", realIP)
	resp, err := http.DefaultClient.Do(req)
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	must.NoError(t, err)
	return string(body)
}
