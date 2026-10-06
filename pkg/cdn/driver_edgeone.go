package cdn

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cast"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sdkerror "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	teo "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/teo/v20220901"
)

type EdgeOne struct {
	secretId, secretKey string
}

func (r *EdgeOne) RefreshUrl(ctx context.Context, urls []string) error {
	// copied: the other drivers receive the same slice
	targets := make([]string, len(urls))
	for i, url := range urls {
		targets[i] = strings.TrimSuffix(url, "*")
	}

	return r.purge(ctx, "purge_url", "url", targets)
}

func (r *EdgeOne) RefreshPath(ctx context.Context, paths []string) error {
	return r.purge(ctx, "purge_prefix", "path", paths)
}

func (r *EdgeOne) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
	client, err := r.client()
	if err != nil {
		return 0, err
	}

	request := teo.NewDescribeTimingL7AnalysisDataRequest()
	request.StartTime = new(startTime.Format(iso8601Layout))
	request.EndTime = new(endTime.Format(iso8601Layout))
	request.MetricNames = common.StringPtrs([]string{"l7Flow_request"})
	request.ZoneIds = common.StringPtrs([]string{"*"})
	request.Interval = new("day")

	response, err := client.DescribeTimingL7AnalysisDataWithContext(ctx, request)
	if sdkError, ok := errors.AsType[*sdkerror.TencentCloudSDKError](err); ok {
		return 0, fmt.Errorf("cdn: failed to get edgeone usage, code: %s, message: %s, requestId: %s", sdkError.Code, sdkError.Message, sdkError.RequestId)
	}
	if err != nil {
		return 0, fmt.Errorf("cdn: failed to get edgeone usage: %w", err)
	}

	// no traffic in the window comes back as an empty data set
	res := response.Response
	if res == nil || res.TotalCount == nil || *res.TotalCount == 0 || len(res.Data) == 0 || res.Data[0] == nil ||
		len(res.Data[0].TypeValue) == 0 || res.Data[0].TypeValue[0] == nil || res.Data[0].TypeValue[0].Sum == nil {
		return 0, nil
	}

	return cast.ToUint(*res.Data[0].TypeValue[0].Sum), nil
}

func (r *EdgeOne) purge(ctx context.Context, purgeType, kind string, targets []string) error {
	client, err := r.client()
	if err != nil {
		return err
	}

	request := teo.NewCreatePurgeTaskRequest()
	request.ZoneId = new("*")
	request.Type = new(purgeType)
	request.Targets = common.StringPtrs(targets)

	_, err = client.CreatePurgeTaskWithContext(ctx, request)
	if sdkError, ok := errors.AsType[*sdkerror.TencentCloudSDKError](err); ok {
		return fmt.Errorf("cdn: failed to refresh edgeone %s, code: %s, message: %s, requestId: %s", kind, sdkError.Code, sdkError.Message, sdkError.RequestId)
	}
	if err != nil {
		return fmt.Errorf("cdn: failed to refresh edgeone %s: %w", kind, err)
	}

	return nil
}

func (r *EdgeOne) client() (*teo.Client, error) {
	credential := common.NewCredential(r.secretId, r.secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "teo.tencentcloudapi.com"

	client, err := teo.NewClient(credential, "ap-chongqing", cpf)
	if err != nil {
		return nil, fmt.Errorf("cdn: failed to create edgeone client: %w", err)
	}

	return client, nil
}
