package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cache"

	mocksbiz "github.com/weavatar/weavatar/internal/mocks/verifycode/biz"
	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/internal/shared/transport"
	"github.com/weavatar/weavatar/internal/verifycode/biz"
	"github.com/weavatar/weavatar/internal/verifycode/service"
)

const captcha = `"captcha":{"lot_number":"lot","captcha_output":"out","pass_token":"pass","gen_time":"1"}`

// envelope decodes the response wrapper.
type envelope struct {
	Msg string `json:"msg"`
}

// harness serves the verify code routes, throttle included, against mocked
// senders whose funcs start nil.
type harness struct {
	app  *fiber.App
	sms  *mocksbiz.SMSSender
	mail *mocksbiz.MailSender
}

func TestSmsSendsCodeForAllowedPurpose(t *testing.T) {
	h := newHarness(t)
	h.sms.SendFunc = okSender

	status, body := h.post(t, "/api/verify_code/sms", smsBody("13800138000"))

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, body.Msg, "success")
	sent := h.sms.SendCalls()
	must.Len(t, sent, 1)
	check.Equal(t, sent[0].Phone, "13800138000")
}

func TestEmailSendsCodeByMail(t *testing.T) {
	h := newHarness(t)
	h.mail.SendFunc = okSender

	status, _ := h.post(t, "/api/verify_code/email", emailBody("a@weavatar.com"))

	must.Equal(t, status, fiber.StatusOK)
	sent := h.mail.SendCalls()
	must.Len(t, sent, 1)
	check.Equal(t, sent[0].To, "a@weavatar.com")
}

func TestInvalidRequestsAreUnprocessable(t *testing.T) {
	h := newHarness(t) // no sender funcs: validation must fail first

	for _, tc := range []struct{ path, body string }{
		{"/api/verify_code/sms", smsBody("12345")},
		{"/api/verify_code/sms", `{"phone":"13800138000","use_for":"login",` + captcha + `}`},
		{"/api/verify_code/sms", `{"phone":"13800138000","use_for":"avatar"}`},
		{"/api/verify_code/email", emailBody("not-an-email")},
	} {
		status, _ := h.post(t, tc.path, tc.body)
		check.Equal(t, status, fiber.StatusUnprocessableEntity, tc.body)
	}
}

func TestCooldownIsUnprocessable(t *testing.T) {
	h := newHarness(t)
	h.sms.SendFunc = okSender
	status, _ := h.post(t, "/api/verify_code/sms", smsBody("13800138000"))
	must.Equal(t, status, fiber.StatusOK)

	status, body := h.post(t, "/api/verify_code/sms", smsBody("13800138000"))

	check.Equal(t, status, fiber.StatusUnprocessableEntity)
	check.Equal(t, body.Msg, "请勿频繁发送验证码")
	check.Len(t, h.sms.SendCalls(), 1)
}

func TestDeliveryFailureIsInternalError(t *testing.T) {
	h := newHarness(t)
	h.mail.SendFunc = func(context.Context, string, string) error { return errors.New("smtp down") }

	status, body := h.post(t, "/api/verify_code/email", emailBody("a@weavatar.com"))

	check.Equal(t, status, fiber.StatusInternalServerError)
	check.Equal(t, body.Msg, "WeAvatar 服务出现错误")
}

func TestSmsAndEmailShareOneThrottleBudget(t *testing.T) {
	h := newHarness(t)
	h.sms.SendFunc = okSender
	h.mail.SendFunc = okSender

	// distinct targets keep the per-target cooldown out of the way
	for i := range 3 {
		status, _ := h.post(t, "/api/verify_code/sms", smsBody("1380013800"+strconv.Itoa(i)))
		must.Equal(t, status, fiber.StatusOK)
	}
	for i := range 2 {
		status, _ := h.post(t, "/api/verify_code/email", emailBody("u"+strconv.Itoa(i)+"@weavatar.com"))
		must.Equal(t, status, fiber.StatusOK)
	}

	status, _ := h.post(t, "/api/verify_code/email", emailBody("u9@weavatar.com"))
	check.Equal(t, status, http.StatusTooManyRequests)
	check.Len(t, h.mail.SendCalls(), 2)
}

func newHarness(t *testing.T) *harness {
	t.Helper()

	h := &harness{
		app:  fiber.New(),
		sms:  &mocksbiz.SMSSender{},
		mail: &mocksbiz.MailSender{},
	}
	c := cache.NewCache(cache.WithCleanupInterval(0))
	uc := biz.NewCodeUsecase(c, h.sms, h.mail, appinfo.CodeExpire(5*time.Minute))
	mount(h.app, service.VerifyCodeRoutes(service.NewVerifyCodeService(uc, newValidator(t))))

	return h
}

func (h *harness) post(t *testing.T, path, payload string) (int, envelope) {
	t.Helper()

	req := httptest.NewRequest(fiber.MethodPost, path, strings.NewReader(payload))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := h.app.Test(req)
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var body envelope
	must.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return resp.StatusCode, body
}

func smsBody(phone string) string {
	return `{"phone":"` + phone + `","use_for":"avatar",` + captcha + `}`
}

func emailBody(email string) string {
	return `{"email":"` + email + `","use_for":"avatar",` + captcha + `}`
}

func okSender(context.Context, string, string) error { return nil }

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
