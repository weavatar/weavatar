package cdn

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/imroc/req/v3"
)

type CTYun struct {
	appID       string
	appSecret   string
	apiEndpoint string
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

// RefreshUrl 刷新URL
func (c *CTYun) RefreshUrl(urls []string) error {
	api := "/api/v1/refreshmanage/create"

	timestamp, signature, err := c.getSignature(api)
	if err != nil {
		return err
	}

	client := req.C()
	client.SetTimeout(60 * time.Second)

	client.SetCommonHeaders(map[string]string{
		"x-alogic-now":       timestamp,
		"x-alogic-app":       c.appID,
		"x-alogic-ac":        "app",
		"x-alogic-signature": signature,
	})

	for i, url := range urls {
		urls[i] = "https://" + url
	}

	data := map[string]any{
		"values":    urls,
		"task_type": 1,
	}

	var resp CTYunRefreshResponse
	_, err = client.R().SetBody(data).SetSuccessResult(&resp).SetErrorResult(&resp).Post(c.apiEndpoint + api)
	if err != nil {
		return err
	}

	if resp.Code != 100000 {
		return fmt.Errorf("cdn: refresh ctyun url failed, code: %d, message: %s", resp.Code, resp.Message)
	}

	return nil
}

// RefreshPath 刷新路径
func (c *CTYun) RefreshPath(paths []string) error {
	api := "/api/v1/refreshmanage/create"

	timestamp, signature, err := c.getSignature(api)
	if err != nil {
		return err
	}

	client := req.C()
	client.SetTimeout(60 * time.Second)

	client.SetCommonHeaders(map[string]string{
		"x-alogic-now":       timestamp,
		"x-alogic-app":       c.appID,
		"x-alogic-ac":        "app",
		"x-alogic-signature": signature,
	})

	// 天翼云文档要求统一使用 http 协议
	for i, path := range paths {
		paths[i] = "http://" + path
	}

	data := map[string]any{
		"values":    paths,
		"task_type": 2,
	}

	var resp CTYunRefreshResponse
	_, err = client.R().SetBody(data).SetSuccessResult(&resp).SetErrorResult(&resp).Post(c.apiEndpoint + api)
	if err != nil {
		return err
	}

	if resp.Code != 100000 {
		return fmt.Errorf("cdn: refresh ctyun path failed, code: %d, message: %s", resp.Code, resp.Message)
	}

	return nil
}

// GetUsage 获取使用量
func (c *CTYun) GetUsage(domain string, startTime, endTime *carbon.Carbon) (uint, error) {
	api := "/api/v2/statisticsanalysis/query_request_num_data"

	timestamp, signature, err := c.getSignature(api)
	if err != nil {
		return 0, err
	}

	client := req.C()
	client.SetTimeout(60 * time.Second)

	client.SetCommonHeaders(map[string]string{
		"x-alogic-now":       timestamp,
		"x-alogic-app":       c.appID,
		"x-alogic-ac":        "app",
		"x-alogic-signature": signature,
	})

	var usage CTYunUsageResponse
	_, err = client.R().SetBodyJsonMarshal(map[string]any{
		"interval":   "24h",
		"domain":     []string{domain},
		"start_time": startTime.Timestamp(),
		"end_time":   endTime.Timestamp(),
	}).SetSuccessResult(&usage).Post(c.apiEndpoint + api)
	if err != nil {
		return 0, err
	}

	if usage.Code != 100000 {
		return 0, fmt.Errorf("cdn: get ctyun usage failed, code: %d, message: %s", usage.Code, usage.Message)
	}

	sum := uint(0)
	for _, data := range usage.ReqRequestNumDataInterval {
		sum += uint(data.RequestNum)
	}

	return sum, nil
}

func (c *CTYun) hmacSha256Byte(target, key string) []byte {
	h := hmac.New(sha256.New, []byte(key))
	h.Write([]byte(target))
	hashBytes := h.Sum(nil)

	return hashBytes
}

func (c *CTYun) encrypt(content, key string) (signature string, err error) {
	// 替换空格为+
	key = strings.ReplaceAll(key, " ", "+")
	// 替换-为+号
	key = strings.ReplaceAll(key, "-", "+")
	// 替换_为/号
	key = strings.ReplaceAll(key, "_", "/")
	// 填充=，字节为4的倍数
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
