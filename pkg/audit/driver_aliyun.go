package audit

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	green20220302 "github.com/alibabacloud-go/green-20220302/v3/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type Aliyun struct {
	accessKeyId     string
	accessKeySecret string
}

// NewAliyun 创建阿里云图片审核实例
func NewAliyun(accessKeyId, accessKeySecret string) Driver {
	return &Aliyun{
		accessKeyId:     accessKeyId,
		accessKeySecret: accessKeySecret,
	}
}

// Check 检查图片是否违规 true: 违规 false: 未违规
func (r *Aliyun) Check(url string) (bool, string, error) {
	client, err := r.createClient("beijing")
	if err != nil {
		return false, "", err
	}

	parameters, err := json.Marshal(map[string]string{
		"imageUrl": url,
	})
	if err != nil {
		return false, "", err
	}

	imageModerationRequest := &green20220302.ImageModerationRequest{
		Service:           tea.String("baselineCheck"),
		ServiceParameters: tea.String(string(parameters)),
	}
	runtime := &util.RuntimeOptions{
		Autoretry:   tea.Bool(true),
		MaxAttempts: tea.Int(3),
	}
	response, _err := client.ImageModerationWithOptions(imageModerationRequest, runtime)

	// 自动切换地域
	flag := false
	if _err != nil {
		var err *tea.SDKError
		if errors.As(_err, &err) {
			// 系统异常，切换到下个地域调用。
			if *err.StatusCode == 500 {
				flag = true
			}
		}
	}
	if response == nil || *response.StatusCode == 500 || *response.Body.Code == 500 {
		flag = true
	}
	if flag {
		client, err := r.createClient("shanghai")
		if err != nil {
			return false, "", err
		}
		response, _err = client.ImageModerationWithOptions(imageModerationRequest, runtime)
		if _err != nil {
			return false, "", _err
		}
	}

	if response != nil {
		statusCode := tea.IntValue(tea.ToInt(response.StatusCode))
		body := response.Body
		imageModerationResponseData := body.Data
		if statusCode == http.StatusOK {
			if tea.IntValue(tea.ToInt(body.Code)) == 200 {
				result := imageModerationResponseData.Result
				remark := ""
				flag := false
				for i := 0; i < len(result); i++ {
					if tea.Float32Value(result[i].Confidence) > 80 {
						flag = true
					}
					remark += fmt.Sprintf("%f-%s(%s), ", tea.Float32Value(result[i].Confidence), tea.StringValue(result[i].Description), tea.StringValue(result[i].Label))
				}
				return flag, remark, nil
			} else {
				return false, "", fmt.Errorf("aliyun audit failed, url:%s, httpCode:%d, requestId:%s, msg:%s", url, statusCode, tea.StringValue(body.RequestId), tea.StringValue(body.Msg))
			}
		} else {
			return false, "", fmt.Errorf("aliyun audit request failed, url:%s, httpCode:%d, requestId:%s, msg:%s", url, statusCode, tea.StringValue(body.RequestId), tea.StringValue(body.Msg))
		}
	}

	return false, "", nil
}

func (r *Aliyun) createClient(endpoint string) (*green20220302.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     &r.accessKeyId,
		AccessKeySecret: &r.accessKeySecret,
	}
	if endpoint == "shanghai" {
		config.RegionId = tea.String("cn-shanghai")
		config.Endpoint = tea.String("green-cip.cn-shanghai.aliyuncs.com")
	}
	if endpoint == "beijing" {
		config.RegionId = tea.String("cn-beijing")
		config.Endpoint = tea.String("green-cip.cn-beijing.aliyuncs.com")
	}

	return green20220302.NewClient(config)
}
