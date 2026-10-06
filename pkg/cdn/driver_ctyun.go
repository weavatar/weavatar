package cdn

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

type CTYun struct {
	appID       string
	appSecret   string
	apiEndpoint string
	client      *req.Client
}

type CTYunRefreshResponse struct {
	Code     int    `json:"code"`
	Message  string `json:"message"`
	SubmitID string `json:"submit_id"`
	Result   []struct {
		TaskID string `json:"task_id"`
		URL    string `json:"url"`
	} `json:"result"`
}

type CTYunUsageResponse struct {
	StartTime                 int64  `json:"start_time"`
	Code                      int    `json:"code"`
	EndTime                   int64  `json:"end_time"`
	Interval                  string `json:"interval"`
	Message                   string `json:"message"`
	ReqRequestNumDataInterval []struct {
		HitRequestRate           float64 `json:"hit_request_rate"`
		TimeStamp                int64   `json:"time_stamp"`
		MissRequestNum           int     `json:"miss_request_num"`
		RequestNum               int     `json:"request_num"`
		ApplicationLayerProtocol string  `json:"application_layer_protocol"`
	} `json:"req_request_num_data_interval"`
}

func newCTYun(cfg CTYunConfig) *CTYun {
	return &CTYun{
		appID:       cfg.AppID,
		appSecret:   cfg.AppSecret,
		apiEndpoint: "https://open.ctcdn.cn",
		client:      newClient(),
	}
}

func (c *CTYun) RefreshUrl(ctx context.Context, urls []string) error {
	return c.refresh(ctx, 1, "url", urls)
}

// RefreshPath sends http:// paths, which CTYun requires; paths is copied
// because the other drivers receive the same slice.
func (c *CTYun) RefreshPath(ctx context.Context, paths []string) error {
	values := make([]string, len(paths))
	for i, path := range paths {
		values[i] = strings.ReplaceAll(path, "https://", "http://")
	}
	return c.refresh(ctx, 2, "path", values)
}

func (c *CTYun) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
	const api = "/api/v2/statisticsanalysis/query_request_num_data"
	request, err := c.request(ctx, api)
	if err != nil {
		return 0, err
	}

	var usage CTYunUsageResponse
	_, err = request.SetBodyJsonMarshal(map[string]any{
		"interval":   "24h",
		"domain":     []string{domain},
		"start_time": startTime.Unix(),
		"end_time":   endTime.Unix(),
	}).SetSuccessResult(&usage).Post(c.apiEndpoint + api)
	if err != nil {
		return 0, err
	}

	if usage.Code != 100000 {
		return 0, fmt.Errorf("cdn: get ctyun usage failed, code: %d, message: %s", usage.Code, usage.Message)
	}

	var sum uint
	for _, data := range usage.ReqRequestNumDataInterval {
		sum += uint(data.RequestNum)
	}

	return sum, nil
}

func (c *CTYun) refresh(ctx context.Context, taskType int, kind string, values []string) error {
	const api = "/api/v1/refreshmanage/create"
	request, err := c.request(ctx, api)
	if err != nil {
		return err
	}

	var resp CTYunRefreshResponse
	_, err = request.SetBody(map[string]any{
		"values":    values,
		"task_type": taskType,
	}).SetSuccessResult(&resp).SetErrorResult(&resp).Post(c.apiEndpoint + api)
	if err != nil {
		return err
	}

	if resp.Code != 100000 {
		return fmt.Errorf("cdn: refresh ctyun %s failed, code: %d, message: %s", kind, resp.Code, resp.Message)
	}

	return nil
}

// request signs for one api path; the signature embeds the current time, so
// it goes on the request rather than the shared client.
func (c *CTYun) request(ctx context.Context, api string) (*req.Request, error) {
	timestamp, signature, err := c.getSignature(api)
	if err != nil {
		return nil, err
	}

	return c.client.R().SetContext(ctx).SetHeaders(map[string]string{
		"x-alogic-now":       timestamp,
		"x-alogic-app":       c.appID,
		"x-alogic-ac":        "app",
		"x-alogic-signature": signature,
	}), nil
}

func (c *CTYun) hmacSha256Byte(target, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(target))
	hashBytes := h.Sum(nil)

	return hashBytes
}

func (c *CTYun) encrypt(content, key string) (signature string, err error) {
	// the secret is unpadded URL-safe base64, sometimes with spaces for '+'
	key = strings.ReplaceAll(key, " ", "+")
	key = strings.ReplaceAll(key, "-", "+")
	key = strings.ReplaceAll(key, "_", "/")
	for len(key)%4 != 0 {
		key += "="
	}
	b64Code, err := base64.StdEncoding.DecodeString(key)
	if err != nil {
		return "", err
	}

	signedByte := c.hmacSha256Byte(content, string(b64Code))
	signedStr := base64.URLEncoding.EncodeToString(signedByte)
	signature = strings.ReplaceAll(signedStr, "=", "")

	return signature, nil
}

func (c *CTYun) getSignature(url string) (string, string, error) {
	timestampMs := time.Now().Unix() * 1000
	timestampDay := timestampMs / 86400000
	timestampMsStr := strconv.FormatInt(timestampMs, 10)

	signStr := fmt.Sprintf("%s\n%v\n%s", c.appID, timestampMs, url)
	identity := fmt.Sprintf("%s:%v", c.appID, timestampDay)

	tmpSignature, err := c.encrypt(identity, c.appSecret)
	if err != nil {
		return "", "", err
	}

	signature, err := c.encrypt(signStr, tmpSignature)
	if err != nil {
		return "", "", err
	}

	return timestampMsStr, signature, nil
}
