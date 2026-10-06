package cdn

import (
	"context"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
	"github.com/spf13/cast"
)

// WafPro drives the SCDN console API that WafPro and WjDun each host.
type WafPro struct {
	name     string
	endpoint string
	client   *req.Client
}

type WafProClean struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

type WafProRefreshResponse struct {
	Code    any    `json:"code"`
	Message string `json:"msg"`
}

type WafProUsageResponse struct {
	Code    any       `json:"code"`
	Data    [][2]uint `json:"data"`
	Message string    `json:"msg"`
}

func newWafPro(name, endpoint, apiKey, apiSecret string) *WafPro {
	return &WafPro{
		name:     name,
		endpoint: endpoint,
		client: newClient().SetCommonHeaders(map[string]string{
			"api-key":    apiKey,
			"api-secret": apiSecret,
		}),
	}
}

// RefreshUrl appends "*" so query-string variants are purged too.
func (d *WafPro) RefreshUrl(ctx context.Context, urls []string) error {
	jobs := make([]WafProClean, len(urls))
	for i, url := range urls {
		jobs[i] = WafProClean{Type: "clean_url", Data: map[string]string{"url": url + "*"}}
	}
	return d.clean(ctx, "url", jobs)
}

func (d *WafPro) RefreshPath(ctx context.Context, paths []string) error {
	jobs := make([]WafProClean, len(paths))
	for i, path := range paths {
		jobs[i] = WafProClean{Type: "clean_dir", Data: map[string]string{"url": path}}
	}
	return d.clean(ctx, "path", jobs)
}

func (d *WafPro) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
	var resp WafProUsageResponse
	_, err := d.client.R().SetContext(ctx).SetSuccessResult(&resp).SetErrorResult(&resp).SetQueryParams(map[string]string{
		"type":        "req",
		"start":       startTime.Format(time.DateTime),
		"end":         endTime.Format(time.DateTime),
		"domain":      domain,
		"server_post": "",
	}).Get(d.endpoint + "/v1/monitor/site/realtime")
	if err != nil {
		return 0, err
	}

	if code := cast.ToString(resp.Code); code != "0" {
		return 0, fmt.Errorf("cdn: failed to get %s usage, code: %s, message: %s", d.name, code, resp.Message)
	}

	var sum uint
	for _, point := range resp.Data {
		sum += point[1]
	}

	return sum, nil
}

func (d *WafPro) clean(ctx context.Context, kind string, jobs []WafProClean) error {
	var resp WafProRefreshResponse
	_, err := d.client.R().SetContext(ctx).SetBody(jobs).SetSuccessResult(&resp).SetErrorResult(&resp).Post(d.endpoint + "/v1/jobs")
	if err != nil {
		return err
	}

	if code := cast.ToString(resp.Code); code != "0" {
		return fmt.Errorf("cdn: failed to refresh %s %s, code: %s, message: %s", d.name, kind, code, resp.Message)
	}

	return nil
}
