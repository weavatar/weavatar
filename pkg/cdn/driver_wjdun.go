package cdn

import (
	"fmt"

	"github.com/dromara/carbon/v2"
	"github.com/imroc/req/v3"
	"github.com/spf13/cast"
)

type WjDun struct {
	apiKey, apiSecret string
}

type WjDunClean struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

type WjDunRefreshResponse struct {
	Code    any    `json:"code"`
	Message string `json:"msg"`
}

type WjDunUsageResponse struct {
	Code    any       `json:"code"`
	Data    [][2]uint `json:"data"`
	Message string    `json:"msg"`
}

// RefreshUrl 刷新URL
func (d *WjDun) RefreshUrl(urls []string) error {
	client := req.C()

	data := make([]WjDunClean, len(urls))
	for i, url := range urls {
		data[i] = WjDunClean{
			Type: "clean_url",
			Data: map[string]string{"url": url + "*"},
		}
	}

	var resp WjDunRefreshResponse
	_, err := client.R().SetBody(data).SetSuccessResult(&resp).SetErrorResult(&resp).SetHeaders(map[string]string{
		"api-key":    d.apiKey,
		"api-secret": d.apiSecret,
	}).Post("https://user.wjdun.cn/v1/jobs")
	if err != nil {
		return err
	}

	if cast.ToString(resp.Code) != "0" {
		return fmt.Errorf("cdn: failed to refresh wjdun url, code: %s, message: %s", cast.ToString(resp.Code), resp.Message)
	}

	return nil
}

// RefreshPath 刷新路径
func (d *WjDun) RefreshPath(paths []string) error {
	client := req.C()

	data := make([]WjDunClean, len(paths))
	for i, url := range paths {
		data[i] = WjDunClean{
			Type: "clean_dir",
			Data: map[string]string{"url": url},
		}
	}

	var resp WjDunRefreshResponse
	_, err := client.R().SetBody(data).SetSuccessResult(&resp).SetErrorResult(&resp).SetHeaders(map[string]string{
		"api-key":    d.apiKey,
		"api-secret": d.apiSecret,
	}).Post("https://user.wjdun.cn/v1/jobs")
	if err != nil {
		return err
	}

	if cast.ToString(resp.Code) != "0" {
		return fmt.Errorf("cdn: failed to refresh wjdun path, code: %s, message: %s", cast.ToString(resp.Code), resp.Message)
	}

	return nil
}

// GetUsage 获取用量
func (d *WjDun) GetUsage(domain string, startTime, endTime *carbon.Carbon) (uint, error) {
	client := req.C()

	var resp WjDunUsageResponse
	_, err := client.R().SetSuccessResult(&resp).SetErrorResult(&resp).SetHeaders(map[string]string{
		"api-key":    d.apiKey,
		"api-secret": d.apiSecret,
	}).Get("https://user.wjdun.cn/v1/monitor/site/realtime?type=req&start=" + startTime.ToDateString() + "%2000:00:00" + "&end=" + endTime.ToDateString() + "%2000:00:00" + "&domain=" + domain + "&server_post=")

	if err != nil {
		return 0, err
	}

	if cast.ToString(resp.Code) != "0" {
		return 0, fmt.Errorf("cdn: failed to get wjdun usage, code: %s, message: %s", cast.ToString(resp.Code), resp.Message)
	}

	sum := uint(0)
	for _, data := range resp.Data {
		sum += data[1]
	}

	return sum, nil
}
