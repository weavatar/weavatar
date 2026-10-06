package transport_test

import (
	"io"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/utils/jwt"

	"github.com/weavatar/weavatar/internal/shared/transport"
)

const testKey = "a-long-string-with-32-characters"

func TestMustLoginAdmitsValidToken(t *testing.T) {
	parser := jwt.NewJWT(testKey, time.Hour)

	status, body := authRequest(t, parser, "Bearer "+issue(t, parser, "u123456789"))
	must.Equal(t, status, fiber.StatusOK)
	must.Equal(t, body, "u123456789")
}

func TestMustLoginRejects(t *testing.T) {
	parser := jwt.NewJWT(testKey, time.Hour)
	expired := jwt.NewJWT(testKey, -time.Hour)
	foreign := jwt.NewJWT("another-key-with-32-characters!!", time.Hour)

	for name, tc := range map[string]struct {
		header string
		want   string
	}{
		"missing header":   {header: "", want: "未登录"},
		"wrong scheme":     {header: "Basic " + issue(t, parser, "u1"), want: "未登录"},
		"empty token":      {header: "Bearer ", want: "未登录"},
		"garbage token":    {header: "Bearer not.a.jwt", want: "未登录"},
		"foreign key":      {header: "Bearer " + issue(t, foreign, "u1"), want: "未登录"},
		"empty subject":    {header: "Bearer " + issue(t, parser, ""), want: "未登录"},
		"expired token":    {header: "Bearer " + issue(t, expired, "u1"), want: "登录已过期"},
		"lowercase scheme": {header: "bearer " + issue(t, expired, "u1"), want: "登录已过期"},
	} {
		t.Run(name, func(t *testing.T) {
			status, body := authRequest(t, parser, tc.header)
			must.Equal(t, status, fiber.StatusUnauthorized)
			must.Contains(t, body, tc.want)
		})
	}
}

func TestBearerToken(t *testing.T) {
	for header, want := range map[string]string{
		"Bearer abc":   "abc",
		"bearer abc":   "abc",
		"Bearer  abc ": "abc",
		"Bearer ":      "",
		"Basic abc":    "",
		"abc":          "",
		"":             "",
	} {
		got, ok := transport.BearerToken(header)
		must.Equal(t, got, want, must.Msgf("header %q", header))
		must.Equal(t, ok, want != "", must.Msgf("header %q", header))
	}
}

func authRequest(t *testing.T, parser *jwt.JWT, header string) (int, string) {
	t.Helper()
	app := fiber.New()
	app.Get("/me", transport.MustLogin(parser), func(c fiber.Ctx) error {
		return c.SendString(transport.UserID(c))
	})

	req := httptest.NewRequest(fiber.MethodGet, "/me", nil)
	if header != "" {
		req.Header.Set(fiber.HeaderAuthorization, header)
	}
	resp, err := app.Test(req)
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	must.NoError(t, err)
	return resp.StatusCode, string(body)
}

func issue(t *testing.T, parser *jwt.JWT, subject string) string {
	t.Helper()
	token, err := parser.Generate(&jwt.Claims{Subject: subject})
	must.NoError(t, err)
	return token
}
