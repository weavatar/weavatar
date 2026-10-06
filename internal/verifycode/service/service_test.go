package service_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
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

func TestSms(t *testing.T) {
	app, sms, _ := newTestApp(t)
	sms.SendFunc = okSender

	status, msg := post(t, app, "/api/verify_code/sms", smsBody("13800138000"))

	must.Equal(t, status, fiber.StatusOK)
	check.Equal(t, msg, "success")
	sent := sms.SendCalls()
	must.Len(t, sent, 1)
	check.Equal(t, sent[0].Phone, "13800138000")
}

func TestEmail(t *testing.T) {
	app, _, mail := newTestApp(t)
	mail.SendFunc = okSender

	status, _ := post(t, app, "/api/verify_code/email", emailBody("a@weavatar.com"))

	must.Equal(t, status, fiber.StatusOK)
	sent := mail.SendCalls()
	must.Len(t, sent, 1)
	check.Equal(t, sent[0].To, "a@weavatar.com")
}

func TestInvalidRequestsAre422(t *testing.T) {
	app, _, _ := newTestApp(t) // no sender funcs: validation must fail first

	for _, tc := range []struct{ path, body string }{
		{"/api/verify_code/sms", smsBody("12345")},
		{"/api/verify_code/sms", `{"phone":"13800138000","use_for":"login",` + captcha + `}`},
		{"/api/verify_code/sms", `{"phone":"13800138000","use_for":"avatar"}`},
		{"/api/verify_code/email", emailBody("not-an-email")},
	} {
		status, _ := post(t, app, tc.path, tc.body)
		check.Equal(t, status, fiber.StatusUnprocessableEntity, tc.body)
	}
}

func TestTooFrequentIs422(t *testing.T) {
	app, sms, _ := newTestApp(t)
	sms.SendFunc = okSender
	status, _ := post(t, app, "/api/verify_code/sms", smsBody("13800138000"))
	must.Equal(t, status, fiber.StatusOK)

	status, msg := post(t, app, "/api/verify_code/sms", smsBody("13800138000"))

	check.Equal(t, status, fiber.StatusUnprocessableEntity)
	check.Equal(t, msg, "请勿频繁发送验证码")
	check.Len(t, sms.SendCalls(), 1)
}

func TestDeliveryFailureIs500(t *testing.T) {
	app, _, mail := newTestApp(t)
	mail.SendFunc = func(context.Context, string, string) error { return errors.New("smtp down") }

	status, msg := post(t, app, "/api/verify_code/email", emailBody("a@weavatar.com"))

	check.Equal(t, status, fiber.StatusInternalServerError)
	check.Equal(t, msg, "WeAvatar 服务出现错误")
}

func TestThrottleIsSharedAcrossEndpoints(t *testing.T) {
	app, sms, mail := newTestApp(t)
	sms.SendFunc = okSender
	mail.SendFunc = okSender

	// distinct targets keep the per-target cooldown out of the way
	for i := range 3 {
		status, _ := post(t, app, "/api/verify_code/sms", smsBody("1380013800"+strconv.Itoa(i)))
		must.Equal(t, status, fiber.StatusOK)
	}
	for i := range 2 {
		status, _ := post(t, app, "/api/verify_code/email", emailBody("u"+strconv.Itoa(i)+"@weavatar.com"))
		must.Equal(t, status, fiber.StatusOK)
	}

	status, _ := post(t, app, "/api/verify_code/email", emailBody("u9@weavatar.com"))
	check.Equal(t, status, http.StatusTooManyRequests)
	check.Len(t, mail.SendCalls(), 2)
}

// newTestApp mounts the real route table, throttle included, over mocked senders.
func newTestApp(t *testing.T) (*fiber.App, *mocksbiz.SMSSender, *mocksbiz.MailSender) {
	t.Helper()

	sms := &mocksbiz.SMSSender{}
	mail := &mocksbiz.MailSender{}
	c := cache.NewCache(cache.WithCleanupInterval(0))
	code := service.NewVerifyCodeService(
		biz.NewCodeUsecase(c, sms, mail, appinfo.CodeExpire(5*time.Minute)),
		newValidator(t),
	)

	app := fiber.New()
	for _, e := range service.VerifyCodeRoutes(code) {
		handlers := make([]any, 0, len(e.Middlewares)+1)
		for _, m := range e.Middlewares {
			handlers = append(handlers, m)
		}
		handlers = append(handlers, e.Handler)
		app.Add([]string{e.Method}, e.Path, handlers[0], handlers[1:]...)
	}

	return app, sms, mail
}

func post(t *testing.T, app *fiber.App, path, body string) (int, string) {
	t.Helper()

	req := httptest.NewRequest(fiber.MethodPost, path, strings.NewReader(body))
	req.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	resp, err := app.Test(req)
	must.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	must.NoError(t, err)

	var env transport.Envelope[any]
	must.NoError(t, json.Unmarshal(raw, &env), string(raw))
	return resp.StatusCode, env.Msg
}

func smsBody(phone string) string {
	return `{"phone":"` + phone + `","use_for":"avatar",` + captcha + `}`
}

func emailBody(email string) string {
	return `{"email":"` + email + `","use_for":"avatar",` + captcha + `}`
}

func okSender(context.Context, string, string) error { return nil }
