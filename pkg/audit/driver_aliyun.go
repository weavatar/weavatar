package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green20220302 "github.com/alibabacloud-go/green-20220302/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type Aliyun struct {
	accessKeyId     string
	accessKeySecret string
}

func NewAliyun(accessKeyId, accessKeySecret string) Driver {
	return &Aliyun{
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
	}
}

func (r *Aliyun) Check(ctx context.Context, url string) (bool, string, error) {
	parameters, err := json.Marshal(map[string]string{
		"imageUrl": url,
	})
	if err != nil {
		return false, "", err
	}

	request := &green20220302.ImageModerationRequest{
		Service:           new("baselineCheck"),
		ServiceParameters: new(string(parameters)),
	}

	// Shanghai backs up Beijing when it errors
	response, err := r.moderate(ctx, "beijing", request)
	if ctx.Err() == nil && (err != nil || response == nil || response.Body == nil ||
		tea.Int32Value(response.StatusCode) == http.StatusInternalServerError ||
		tea.Int32Value(response.Body.Code) == http.StatusInternalServerError) {
		response, err = r.moderate(ctx, "shanghai", request)
	}
	if err != nil {
		return false, "", err
	}
	if response == nil || response.Body == nil {
		return false, "", fmt.Errorf("aliyun audit returned empty response, url:%s", url)
	}

	statusCode := tea.Int32Value(response.StatusCode)
	body := response.Body
	if statusCode != http.StatusOK {
		return false, "", fmt.Errorf("aliyun audit request failed, url:%s, httpCode:%d, requestId:%s, msg:%s", url, statusCode, tea.StringValue(body.RequestId), tea.StringValue(body.Msg))
	}
	if tea.Int32Value(body.Code) != http.StatusOK {
		return false, "", fmt.Errorf("aliyun audit failed, url:%s, httpCode:%d, requestId:%s, msg:%s", url, statusCode, tea.StringValue(body.RequestId), tea.StringValue(body.Msg))
	}
	if body.Data == nil {
		return false, "", nil
	}

	var remark strings.Builder
	banned := false
	for _, result := range body.Data.Result {
		if result == nil {
			continue
		}
		if tea.Float32Value(result.Confidence) > 80 {
			banned = true
		}
		_, _ = fmt.Fprintf(&remark, "%f-%s(%s), ", tea.Float32Value(result.Confidence), tea.StringValue(result.Description), tea.StringValue(result.Label))
	}

	return banned, remark.String(), nil
}

func (r *Aliyun) moderate(ctx context.Context, region string, request *green20220302.ImageModerationRequest) (*green20220302.ImageModerationResponse, error) {
	client, err := r.createClient(region)
	if err != nil {
		return nil, err
	}

	return client.ImageModerationWithContext(ctx, request, &util.RuntimeOptions{
		Autoretry:   new(true),
		MaxAttempts: new(3),
	})
}

func (r *Aliyun) createClient(region string) (*green20220302.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     &r.accessKeyId,
		AccessKeySecret: &r.accessKeySecret,
		RegionId:        new("cn-" + region),
		Endpoint:        new("green-cip.cn-" + region + ".aliyuncs.com"),
	}

	return green20220302.NewClient(config)
}
