package cdn

import (
	"fmt"

	"github.com/dromara/carbon/v2"
	"github.com/imroc/req/v3"
)

type WafPro struct {
	apiKey, apiSecret string
}

type WafProClean struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

type WafProRefreshResponse struct {
	Code    uint   `json:"code"`
	Message string `json:"msg"`
}

type WafProUsageResponse struct {
	Code    uint      `json:"code"`
	Data    [][2]uint `json:"data"`
	Message string    `json:"msg"`
}

// RefreshUrl 刷新URL
func (d *WafPro) RefreshUrl(urls []string) error {
	client := req.C()

	data := make([]WafProClean, len(urls))
	for i, url := range urls {
		data[i] = WafProClean{
			Type: "clean_url",
			Data: map[string]string{"url": "https://" + url + "*"},
		}
	}

	var resp WafProRefreshResponse
	_, err := client.R().SetBody(data).SetSuccessResult(&resp).SetErrorResult(&resp).SetHeaders(map[string]string{
		"api-key":    d.apiKey,
		"api-secret": d.apiSecret,
	}).Post("https://scdn.console.waf.pro/v1/jobs")
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("cdn: failed to refresh wafpro url, code: %d, message: %s", resp.Code, resp.Message)
	}

	return nil
}

// RefreshPath 刷新路径
func (d *WafPro) RefreshPath(paths []string) error {
	client := req.C()

	data := make([]WafProClean, len(paths))
	for i, url := range paths {
		data[i] = WafProClean{
			Type: "clean_dir",
			Data: map[string]string{"url": "https://" + url},
		}
	}

	var resp WafProRefreshResponse
	_, err := client.R().SetBody(data).SetSuccessResult(&resp).SetErrorResult(&resp).SetHeaders(map[string]string{
		"api-key":    d.apiKey,
		"api-secret": d.apiSecret,
	}).Post("https://scdn.console.waf.pro/v1/jobs")
	if err != nil {
		return err
	}

	if resp.Code != 0 {
		return fmt.Errorf("cdn: failed to refresh wafpro path, code: %d, message: %s", resp.Code, resp.Message)
	}

	return nil
}

// GetUsage 获取用量
func (d *WafPro) GetUsage(domain string, startTime, endTime *carbon.Carbon) (uint, error) {
	client := req.C()

	var resp WafProUsageResponse
	_, err := client.R().SetSuccessResult(&resp).SetErrorResult(&resp).SetHeaders(map[string]string{
		"api-key":    d.apiKey,
		"api-secret": d.apiSecret,
	}).Get("https://scdn.console.waf.pro/v1/monitor/site/realtime?type=req&start=" + startTime.ToDateString() + "%2000:00:00" + "&end=" + endTime.ToDateString() + "%2000:00:00" + "&domain=" + domain + "&server_post=")

	if err != nil {
		return 0, err
	}

	if resp.Code != 0 {
		return 0, fmt.Errorf("cdn: failed to get wafpro usage, code: %d, message: %s", resp.Code, resp.Message)
	}

	sum := uint(0)
	for _, data := range resp.Data {
		sum += data[1]
	}

	return sum, nil
}
