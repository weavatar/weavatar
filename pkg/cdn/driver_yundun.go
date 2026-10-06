package cdn

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/imroc/req/v3"
	"github.com/libtnb/utils/str"
)

type YunDun struct {
	username, password string
	client             *req.Client // its cookie jar holds the console session
}

type YunDunRefreshResponse struct {
	Status struct {
		Code                        int    `json:"code"`
		Message                     string `json:"message"`
		CreateAt                    string `json:"create_at"`
		ApiTimeConsuming            string `json:"api_time_consuming"`
		FunctionTimeConsuming       string `json:"function_time_consuming"`
		DispatchBeforeTimeConsuming string `json:"dispatch_before_time_consuming"`
	} `json:"status"`
	Data struct {
		Wholesite  []any    `json:"wholesite"`
		Specialurl []string `json:"specialurl"`
		Specialdir []any    `json:"specialdir"`
		RequestId  string   `json:"request_id"`
	} `json:"data"`
}

type YunDunUsageRequest struct {
	Router               string   `json:"router"`
	StartTime            string   `json:"start_time"`
	EndTime              string   `json:"end_time"`
	Nodes                []string `json:"nodes"`
	GroupId              []string `json:"group_id"`
	SubDomain            []string `json:"sub_domain"`
	SubDomainsAndNodeIps struct {
	} `json:"sub_domains_and_node_ips"`
	Interval string `json:"interval"`
}

type YunDunUsageResponse struct {
	Status struct {
		Code                        int    `json:"code"`
		Message                     string `json:"message"`
		CreateAt                    string `json:"create_at"`
		ApiTimeConsuming            string `json:"api_time_consuming"`
		FunctionTimeConsuming       string `json:"function_time_consuming"`
		DispatchBeforeTimeConsuming string `json:"dispatch_before_time_consuming"`
	} `json:"status"`
	Data struct {
		HttpsTimes struct {
			Description string `json:"description"`
			Trend       struct {
				XData []string `json:"x_data"`
				YData []int    `json:"y_data"`
			} `json:"trend"`
			Total struct {
				Unit  string `json:"unit"`
				Total int    `json:"total"`
			} `json:"total"`
		} `json:"https_times"`
		TotalTimes struct {
			Description string `json:"description"`
			Trend       struct {
				XData []string `json:"x_data"`
				YData []int    `json:"y_data"`
			} `json:"trend"`
			Total struct {
				Unit  string `json:"unit"`
				Total int    `json:"total"`
			} `json:"total"`
		} `json:"total_times"`
		HitCacheTimes struct {
			Description string `json:"description"`
			Trend       struct {
				XData []string `json:"x_data"`
				YData []int    `json:"y_data"`
			} `json:"trend"`
			Total struct {
				Unit  string `json:"unit"`
				Total int    `json:"total"`
			} `json:"total"`
		} `json:"hit_cache_times"`
	} `json:"data"`
}

type YunDunErrorResponse struct {
	Status struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	} `json:"status"`
}

func newYunDun(c YunDunConfig) *YunDun {
	client := newClient()
	client.ImpersonateSafari()
	return &YunDun{username: c.Username, password: c.Password, client: client}
}

func (y *YunDun) RefreshUrl(ctx context.Context, urls []string) error {
	return y.refresh(ctx, "specialurl", urls)
}

func (y *YunDun) RefreshPath(ctx context.Context, paths []string) error {
	return y.refresh(ctx, "specialdir", paths)
}

func (y *YunDun) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
	if err := y.login(ctx); err != nil {
		return 0, err
	}

	request := YunDunUsageRequest{
		Router:    "cdn.domain.times",
		StartTime: startTime.Format(time.DateTime),
		EndTime:   endTime.Format(time.DateTime),
		Nodes:     []string{},
		GroupId:   []string{},
		SubDomain: []string{domain},
		Interval:  "1d",
	}
	// both outcomes carry the status block, so one value serves either
	var usage YunDunUsageResponse
	_, err := y.client.R().SetContext(ctx).SetBodyJsonMarshal(request).SetSuccessResult(&usage).SetErrorResult(&usage).Post("https://www.yundun.com/api/V4/stati.data.get")
	if err != nil {
		return 0, err
	}

	if usage.Status.Code != 1 {
		return 0, fmt.Errorf("cdn: failed to get yundun usage, code: %d, message: %s", usage.Status.Code, usage.Status.Message)
	}

	return uint(usage.Data.TotalTimes.Total.Total), nil
}

func (y *YunDun) refresh(ctx context.Context, field string, urls []string) error {
	if err := y.login(ctx); err != nil {
		return err
	}

	var result YunDunRefreshResponse
	_, err := y.client.R().SetContext(ctx).
		SetBody(map[string][]string{field: urls}).
		SetSuccessResult(&result).
		SetErrorResult(&result).
		Put("https://www.yundun.com/api/V4/Web.Domain.DashBoard.saveCache")
	if err != nil {
		return err
	}

	if result.Status.Code != 1 {
		return fmt.Errorf("cdn: failed to refresh yundun %s, code: %d, message: %s", field, result.Status.Code, result.Status.Message)
	}

	return nil
}

// login runs YunDun's console SSO on every call rather than tracking when the
// session expires. The reply is assumed to carry the V4 status block like the
// other endpoints; a reply without one is let through and the API call that
// follows reports its own status.
func (y *YunDun) login(ctx context.Context) error {
	timeStamp := strconv.FormatInt(time.Now().UnixMilli(), 10)
	callback := "jsonp_" + timeStamp + "_" + str.RandomN(16)
	attachURL := fmt.Sprintf("https://www.yundun.com/api/sso/V4/attach?callback=%s&_time=%s", callback, timeStamp)

	resp, err := y.client.R().SetContext(ctx).Get(attachURL)
	if err != nil {
		return err
	}
	if resp.IsErrorState() {
		return fmt.Errorf("cdn: yundun sso attach failed, status: %d", resp.StatusCode)
	}

	resp, err = y.client.R().SetContext(ctx).SetFormData(map[string]string{
		"username": y.username,
		"password": y.password,
	}).Post("https://www.yundun.com/api/sso/V4/login?sso_version=2")
	if err != nil {
		return err
	}
	if resp.IsErrorState() {
		return fmt.Errorf("cdn: yundun login failed, status: %d", resp.StatusCode)
	}

	var result struct {
		Status *struct {
			Code    *int   `json:"code"`
			Message string `json:"message"`
		} `json:"status"`
	}
	if json.Unmarshal(resp.Bytes(), &result) == nil && result.Status != nil && result.Status.Code != nil && *result.Status.Code != 1 {
		return fmt.Errorf("cdn: yundun login failed, code: %d, message: %s", *result.Status.Code, result.Status.Message)
	}

	return nil
}
