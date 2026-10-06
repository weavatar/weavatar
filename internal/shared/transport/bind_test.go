package transport_test

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/shared/transport"
)

type createReq struct {
	Name string `json:"name" validate:"required && min:3 && max:10"`
}

func bindOn[T any](t *testing.T, method, target, body, contentType string) (*T, int) {
	t.Helper()

	var bound *T
	app := fiber.New()
	app.All("/bind/:id?", func(c fiber.Ctx) error {
		req, err := transport.Bind[T](c, validator.MustNew())
		if err != nil {
			return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
		}
		bound = req
		return transport.Success[any](c, nil)
	})

	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, target, reader)
	if contentType != "" {
		req.Header.Set(fiber.HeaderContentType, contentType)
	}
	resp, err := app.Test(req)
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	return bound, resp.StatusCode
}

func TestBindBodyAndValidate(t *testing.T) {
	got, status := bindOn[createReq](t, fiber.MethodPost, "/bind", `{"name":"alice"}`, fiber.MIMEApplicationJSON)
	must.Equal(t, status, fiber.StatusOK)
	must.Equal(t, got.Name, "alice")
}

func TestBindRejectsInvalid(t *testing.T) {
	_, status := bindOn[createReq](t, fiber.MethodPost, "/bind", `{"name":"ab"}`, fiber.MIMEApplicationJSON)
	must.Equal(t, status, fiber.StatusUnprocessableEntity)
}

func TestBindRunsPrepareHook(t *testing.T) {
	got, status := bindOn[transport.Paginate](t, fiber.MethodGet, "/bind", "", "")
	must.Equal(t, status, fiber.StatusOK)
	must.Equal(t, got.Page, 1)
	must.Equal(t, got.Limit, 10)

	got, status = bindOn[transport.Paginate](t, fiber.MethodGet, "/bind?page=3&limit=50", "", "")
	must.Equal(t, status, fiber.StatusOK)
	must.Equal(t, got.Page, 3)
	must.Equal(t, got.Limit, 50)
}

func TestBindQueryOverLimitFailsValidation(t *testing.T) {
	_, status := bindOn[transport.Paginate](t, fiber.MethodGet, "/bind?limit=5000", "", "")
	must.Equal(t, status, fiber.StatusUnprocessableEntity)
}

type uriReq struct {
	ID uint `uri:"id" validate:"required && number"`
}

func TestBindURI(t *testing.T) {
	got, status := bindOn[uriReq](t, fiber.MethodGet, "/bind/42", "", "")
	must.Equal(t, status, fiber.StatusOK)
	must.Equal(t, got.ID, 42)
}

// layeredReq carries fields every binding source can set.
type layeredReq struct {
	Code string `json:"code" form:"code" query:"code"`
	ID   string `json:"id" query:"id" uri:"id"`
}

func TestBindLetsBodyOverrideQueryAndPathOverrideBoth(t *testing.T) {
	got, status := bindOn[layeredReq](t, fiber.MethodPost, "/bind/from-path?code=evil&id=from-query",
		`{"code":"good","id":"from-body"}`, fiber.MIMEApplicationJSON)
	must.Equal(t, status, fiber.StatusOK)
	must.Equal(t, got.Code, "good", must.Msgf("a query parameter cannot replace a body field"))
	must.Equal(t, got.ID, "from-path", must.Msgf("the path is authoritative"))

	got, status = bindOn[layeredReq](t, fiber.MethodPost, "/bind?Code=evil", "code=good", fiber.MIMEApplicationForm)
	must.Equal(t, status, fiber.StatusOK)
	must.Equal(t, got.Code, "good", must.Msgf("query keys match case-insensitively but still lose to the form"))

	got, status = bindOn[layeredReq](t, fiber.MethodPost, "/bind?code=from-query", `{}`, fiber.MIMEApplicationJSON)
	must.Equal(t, status, fiber.StatusOK)
	must.Equal(t, got.Code, "from-query", must.Msgf("the query still fills what the body leaves out"))
}
