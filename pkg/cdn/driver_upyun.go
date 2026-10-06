package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/imroc/req/v3"
)

type UpYun struct {
	token  string
	client *req.Client
}

type UpYunPurgeBatch struct {
	NoIf      uint   `json:"noif"`
	SourceUrl string `json:"source_url"`
}

type UpYunPurgeBatchSuccessResponse struct {
	Code   uint   `json:"code"`
	Status string `json:"status"`
}

type UpYunUsageSuccessResponse struct {
	Rps         float64     `json:"rps"`
	Reqs        float64     `json:"reqs"`
	Bytes       float64     `json:"bytes"`
	Bandwidth   float64     `json:"bandwidth"`
	Treqs       float64     `json:"treqs"`
	Tbytes      float64     `json:"tbytes"`
	Tbandwidth  float64     `json:"tbandwidth"`
	Hreqs       float64     `json:"hreqs"`
	Hbytes      float64     `json:"hbytes"`
	Hbandwidth  float64     `json:"hbandwidth"`
	Dreqs       float64     `json:"dreqs"`
	Dbytes      float64     `json:"dbytes"`
	Dbandwidth  float64     `json:"dbandwidth"`
	Wsbytes     float64     `json:"wsbytes"`
	Wsbandwidth float64     `json:"wsbandwidth"`
	Time        json.Number `json:"time"`
}

type UpYunErrorResponse struct {
	ErrorCode uint   `json:"error_code"`
	Request   string `json:"request"`
	Message   string `json:"message"`
}

func newUpYun(c UpYunConfig) *UpYun {
	return &UpYun{token: c.Token, client: newClient()}
}

func (u *UpYun) RefreshUrl(ctx context.Context, urls []string) error {
	return u.refresh(ctx, "url", urls)
}

func (u *UpYun) RefreshPath(ctx context.Context, paths []string) error {
	return u.refresh(ctx, "path", paths)
}

func (u *UpYun) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
	var successResp []UpYunUsageSuccessResponse
	var errorResp UpYunErrorResponse
	resp, err := u.client.R().SetContext(ctx).SetSuccessResult(&successResp).SetErrorResult(&errorResp).SetBearerAuthToken(u.token).SetQueryParams(map[string]string{
		"start_time":  startTime.Format(iso8601MilliLayout),
		"end_time":    endTime.Format(iso8601MilliLayout),
		"query_type":  "domain",
		"query_value": domain,
		"flow_type":   "cdn",
		"flow_source": "cdn",
	}).Get("https://api.upyun.com/flow/common_data")
	if err != nil {
		return 0, err
	}
	if resp.IsErrorState() {
		return 0, fmt.Errorf("cdn: failed to get upyun usage, status: %d, code: %d, message: %s, request: %s", resp.StatusCode, errorResp.ErrorCode, errorResp.Message, errorResp.Request)
	}

	var sum uint
	for _, data := range successResp {
		sum += uint(data.Reqs)
	}

	return sum, nil
}

func (u *UpYun) refresh(ctx context.Context, kind string, urls []string) error {
	// source_url is newline-terminated entries, the last one included
	var source strings.Builder
	for _, url := range urls {
		source.WriteString(url)
		source.WriteByte('\n')
	}

	var successResp []UpYunPurgeBatchSuccessResponse
	var errorResp UpYunErrorResponse
	resp, err := u.client.R().SetContext(ctx).
		SetBody(UpYunPurgeBatch{NoIf: 1, SourceUrl: source.String()}).
		SetSuccessResult(&successResp).
		SetErrorResult(&errorResp).
		SetBearerAuthToken(u.token).
		Post("https://api.upyun.com/buckets/purge/batch")
	if err != nil {
		return err
	}
	if resp.IsErrorState() {
		return fmt.Errorf("cdn: failed to refresh upyun %s, status: %d, code: %d, message: %s, request: %s", kind, resp.StatusCode, errorResp.ErrorCode, errorResp.Message, errorResp.Request)
	}

	for _, result := range successResp {
		if result.Code != 1 {
			return fmt.Errorf("cdn: failed to refresh upyun %s, code: %d, status: %s", kind, result.Code, result.Status)
		}
	}

	return nil
}
