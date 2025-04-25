package cdn

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/dromara/carbon/v2"
	"github.com/go-rat/utils/str"
	"github.com/imroc/req/v3"
)

type YunDun struct {
	username, password string
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
		Wholesite  []interface{} `json:"wholesite"`
		Specialurl []string      `json:"specialurl"`
		Specialdir []interface{} `json:"specialdir"`
		RequestId  string        `json:"request_id"`
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

// RefreshUrl 刷新URL
func (y *YunDun) RefreshUrl(urls []string) error {
	client, err := y.login()
	if err != nil {
		return err
	}

	// 提交刷新请求
	refreshURL := "https://www.yundun.com/api/V4/Web.Domain.DashBoard.saveCache"
	data := map[string][]string{
		"specialurl": urls,
	}

	var refreshResponse YunDunRefreshResponse
	var errorResponse YunDunErrorResponse
	_, err = client.R().SetBody(data).SetSuccessResult(&refreshResponse).SetErrorResult(&errorResponse).Put(refreshURL)
	if err != nil {
		return err
	}

	if refreshResponse.Status.Code != 1 {
		return fmt.Errorf("cdn: failed to refresh yundun url, code: %d, message: %s", errorResponse.Status.Code, errorResponse.Status.Message)
	}

	return nil
}

// RefreshPath 刷新路径
func (y *YunDun) RefreshPath(paths []string) error {
	client, err := y.login()
	if err != nil {
		return err
	}

	// 提交刷新请求
	refreshURL := "https://www.yundun.com/api/V4/Web.Domain.DashBoard.saveCache"
	data := map[string][]string{
		"specialdir": paths,
	}

	var refreshResponse YunDunRefreshResponse
	var errorResponse YunDunErrorResponse
	_, err = client.R().SetBody(data).SetSuccessResult(&refreshResponse).SetErrorResult(&errorResponse).Put(refreshURL)
	if err != nil {
		return err
	}

	if refreshResponse.Status.Code != 1 {
		return fmt.Errorf("cdn: failed to refresh yundun path, code: %d, message: %s", errorResponse.Status.Code, errorResponse.Status.Message)
	}

	return nil
}

// GetUsage 获取使用量
func (y *YunDun) GetUsage(domain string, startTime, endTime *carbon.Carbon) (uint, error) {

	client, err := y.login()
	if err != nil {
		return 0, err
	}

	var request = YunDunUsageRequest{
		Router:    "cdn.domain.times",
		StartTime: startTime.ToDateTimeString(),
		EndTime:   endTime.ToDateTimeString(),
		Nodes:     []string{},
		GroupId:   []string{},
		SubDomain: []string{domain},
		Interval:  "1d",
	}
	var usageResponse YunDunUsageResponse
	var errorResponse YunDunErrorResponse

	_, err = client.R().SetBodyJsonMarshal(request).SetSuccessResult(&usageResponse).SetErrorResult(&errorResponse).Post("https://www.yundun.com/api/V4/stati.data.get")
	if err != nil {
		return 0, err
	}

	if usageResponse.Status.Code != 1 {
		return 0, fmt.Errorf("cdn: failed to get yundun usage, code: %d, message: %s", errorResponse.Status.Code, errorResponse.Status.Message)
	}

	return uint(usageResponse.Data.TotalTimes.Total.Total), nil
}

// login 登录平台
func (y *YunDun) login() (*req.Client, error) {
	timeStamp := strconv.Itoa(int(carbon.Now(carbon.PRC).TimestampMilli()))
	rand.NewSource(time.Now().UnixNano())
	random := str.RandomN(16)
	callback := "jsonp_" + timeStamp + "_" + random
	attachURL := fmt.Sprintf("https://www.yundun.com/api/sso/V4/attach?callback=%s&_time=%s", callback, timeStamp)

	client := req.C()
	client.ImpersonateSafari()

	// 先获取登录 Token
	_, err := client.R().Get(attachURL)
	if err != nil {
		return nil, err
	}

	// 提交登录请求
	loginURL := "https://www.yundun.com/api/sso/V4/login?sso_version=2"
	loginParams := map[string]string{
		"username": y.username,
		"password": y.password,
	}
	_, err = client.R().SetFormData(loginParams).Post(loginURL)
	if err != nil {
		return nil, err
	}

	return client, nil
}
