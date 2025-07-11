package cdn

import (
	"errors"
	"fmt"
	"strings"

	"github.com/dromara/carbon/v2"
	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sdkerror "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

type EdgeOne struct {
	secretId, secretKey string // 密钥
}

// RefreshUrl 刷新URL
func (r *EdgeOne) RefreshUrl(urls []string) error {
	for i, url := range urls {
		urls[i] = strings.TrimSuffix(url, "*")
	}

	credential := common.NewCredential(
		r.secretId,
		r.secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "teo.tencentcloudapi.com"

	client, err := teo.NewClient(credential, "ap-chongqing", cpf)
	if err != nil {
		return fmt.Errorf("cdn: failed to create edgeone client: %w", err)
	}

	request := teo.NewCreatePurgeTaskRequest()
	request.ZoneId = common.StringPtr("*")
	request.Type = common.StringPtr("purge_url")
	request.Targets = common.StringPtrs(urls)

	_, err = client.CreatePurgeTask(request)

	var sdkError *sdkerror.TencentCloudSDKError
	if errors.As(err, &sdkError) {
		return fmt.Errorf("cdn: failed to refresh edgeone url, code: %s, message: %s, requestId: %s", sdkError.Code, sdkError.Message, sdkError.RequestId)
	}
	if err != nil {
		return fmt.Errorf("cdn: failed to refresh edgeone url, err: %x", err)
	}

	return nil
}

// RefreshPath 刷新路径
func (r *EdgeOne) RefreshPath(paths []string) error {
	credential := common.NewCredential(
		r.secretId,
		r.secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "teo.tencentcloudapi.com"

	client, err := teo.NewClient(credential, "ap-chongqing", cpf)
	if err != nil {
		return fmt.Errorf("cdn: failed to create edgeone client: %w", err)
	}

	request := teo.NewCreatePurgeTaskRequest()
	request.ZoneId = common.StringPtr("*")
	request.Type = common.StringPtr("purge_prefix")
	request.Targets = common.StringPtrs(paths)

	_, err = client.CreatePurgeTask(request)

	var sdkError *sdkerror.TencentCloudSDKError
	if errors.As(err, &sdkError) {
		return fmt.Errorf("cdn: failed to refresh edgeone path, code: %s, message: %s, requestId: %s", sdkError.Code, sdkError.Message, sdkError.RequestId)
	}
	if err != nil {
		return fmt.Errorf("cdn: failed to refresh edgeone path, err: %x", err)
	}

	return nil
}

// GetUsage 获取用量
func (r *EdgeOne) GetUsage(domain string, startTime, endTime *carbon.Carbon) (uint, error) {
	credential := common.NewCredential(
		r.secretId,
		r.secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "teo.tencentcloudapi.com"

	client, err := teo.NewClient(credential, "ap-chongqing", cpf)
	if err != nil {
		return 0, fmt.Errorf("cdn: failed to create edgeone client: %w", err)
	}

	request := teo.NewDescribeTimingL7AnalysisDataRequest()
	request.StartTime = common.StringPtr(startTime.ToIso8601String())
	request.EndTime = common.StringPtr(endTime.ToIso8601String())
	request.MetricNames = common.StringPtrs([]string{"l7Flow_request"})
	request.ZoneIds = common.StringPtrs([]string{"*"})
	request.Interval = common.StringPtr("day")

	response, err := client.DescribeTimingL7AnalysisData(request)
	var sdkError *sdkerror.TencentCloudSDKError
	if errors.As(err, &sdkError) {
		return 0, fmt.Errorf("cdn: failed to get edgeone usage, code: %s, message: %s, requestId: %s", sdkError.Code, sdkError.Message, sdkError.RequestId)
	}
	if err != nil {
		return 0, fmt.Errorf("cdn: failed to get edgeone usage, err: %x", err)
	}

	if *response.Response.TotalCount == uint64(0) || len(response.Response.Data) == 0 || len(response.Response.Data[0].TypeValue) == 0 {
		return 0, nil
	}

	return cast.ToUint(*response.Response.Data[0].TypeValue[0].Sum), nil
}
