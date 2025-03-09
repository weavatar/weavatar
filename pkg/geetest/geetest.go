package geetest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/go-resty/resty/v2"
)

type Geetest struct {
	client     *resty.Client
	CaptchaID  string
	CaptchaKey string
}

func NewGeetest(CaptchaID, CaptchaKey string) *Geetest {
	client := resty.New()
	client.SetBaseURL("https://gcaptcha4.geetest.com")
	client.SetTimeout(5 * time.Second)
	client.SetJSONMarshaler(json.Marshal)
	client.SetJSONUnmarshaler(json.Unmarshal)

	return &Geetest{
		client:     client,
		CaptchaID:  CaptchaID,
		CaptchaKey: CaptchaKey,
	}
}

func (r *Geetest) Verify(ticket Ticket) (bool, error) {
	resp, err := r.client.R().
		SetFormData(map[string]string{
			"lot_number":     ticket.LotNumber,
			"captcha_output": ticket.CaptchaOutput,
			"pass_token":     ticket.PassToken,
			"gen_time":       ticket.GenTime,
			"sign_token":     hmacEncode(r.CaptchaKey, ticket.LotNumber),
		}).
		SetQueryParam("captcha_id", r.CaptchaID).
		ForceContentType("application/json").
		SetResult(&Response{}).
		Post("/validate")

	if err != nil {
		return false, err
	}
	if !resp.IsSuccess() {
		return false, fmt.Errorf("%s %s", resp.Status(), resp.String())
	}

	res := resp.Result().(*Response)
	if res.Status != "success" {
		return false, fmt.Errorf("%s %s", res.Code, res.Msg)
	}

	if res.Result != "success" {
		return false, fmt.Errorf("%s", res.Reason)
	}

	return true, nil
}

func hmacEncode(key string, data string) string {
	hasher := hmac.New(sha256.New, []byte(key))
	hasher.Write([]byte(data))
	return hex.EncodeToString(hasher.Sum(nil))
}
