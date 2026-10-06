package cdn

import (
	"context"
	"fmt"
	"time"

	"github.com/imroc/req/v3"
)

type BaiShan struct {
	token  string
	client *req.Client
}

type BaiShanRefreshResponse struct {
	Code uint `json:"code"`
	Data any  `json:"data"`
}

type BaiShanUsageResponse struct {
	Code int `json:"code"`
	Data map[string]struct {
		Domain string   `json:"domain"`
		Data   [][]uint `json:"data"`
	} `json:"data"`
}

func newBaiShan(c BaiShanConfig) *BaiShan {
	return &BaiShan{token: c.Token, client: newClient()}
}

func (b *BaiShan) RefreshUrl(ctx context.Context, urls []string) error {
	return b.refresh(ctx, "url", urls)
}

func (b *BaiShan) RefreshPath(ctx context.Context, paths []string) error {
	return b.refresh(ctx, "dir", paths)
}

func (b *BaiShan) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
	var usage BaiShanUsageResponse
	resp, err := b.client.R().SetContext(ctx).
		SetQueryParams(map[string]string{
			"token":      b.token,
			"domains":    domain,
			"start_time": startTime.Format(time.DateOnly),
			"end_time":   endTime.Format(time.DateOnly),
		}).
		SetSuccessResult(&usage).
		Get("https://cdn.api.baishan.com/v2/stat/request/eachDomain")
	if err != nil {
		return 0, err
	}
	if resp.IsErrorState() || usage.Code != 0 {
		return 0, fmt.Errorf("cdn: fail to get baishan usage, status: %d, code: %d", resp.StatusCode, usage.Code)
	}

	// each point is [timestamp, requests]
	var sum uint
	for _, item := range usage.Data {
		for _, point := range item.Data {
			if len(point) > 1 {
				sum += point[1]
			}
		}
	}

	return sum, nil
}

func (b *BaiShan) refresh(ctx context.Context, typ string, urls []string) error {
	var result BaiShanRefreshResponse
	resp, err := b.client.R().SetContext(ctx).
		SetBody(map[string]any{
			"urls": urls,
			"type": typ,
		}).
		SetSuccessResult(&result).
		SetErrorResult(&result).
		SetQueryParam("token", b.token).
		Post("https://cdn.api.baishan.com/v2/cache/refresh")
	if err != nil {
		return err
	}
	if resp.IsErrorState() || result.Code != 0 {
		return fmt.Errorf("cdn: fail to refresh baishan %s, status: %d, code: %d", typ, resp.StatusCode, result.Code)
	}

	return nil
}
