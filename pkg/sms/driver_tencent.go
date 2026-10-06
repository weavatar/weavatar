package sms

import (
	"context"
	"errors"
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	sdkerror "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Tencent struct {
	secretId, secretKey, signName, templateId, sdkAppId, expireTime string
}

func (r *Tencent) Send(ctx context.Context, phone string, message Message) error {
	credential := common.NewCredential(r.secretId, r.secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	client, err := tencentsms.NewClient(credential, "ap-beijing", cpf)
	if err != nil {
		return fmt.Errorf("sms: failed to create tencent client: %w", err)
	}

	request := tencentsms.NewSendSmsRequest()
	request.PhoneNumberSet = common.StringPtrs([]string{phone})
	request.SignName = new(r.signName)
	request.TemplateId = new(r.templateId)
	request.TemplateParamSet = common.StringPtrs([]string{message.Data["code"], r.expireTime})
	request.SmsSdkAppId = new(r.sdkAppId)

	response, err := client.SendSmsWithContext(ctx, request)

	if sdkError, ok := errors.AsType[*sdkerror.TencentCloudSDKError](err); ok {
		return fmt.Errorf("sms: failed to send sms, code: %s, message: %s, requestId: %s", sdkError.Code, sdkError.Message, sdkError.RequestId)
	}
	if err != nil {
		return err
	}

	if response.Response == nil || len(response.Response.SendStatusSet) == 0 || response.Response.SendStatusSet[0] == nil {
		return errors.New("sms: tencent returned empty send status")
	}

	status := response.Response.SendStatusSet[0]
	if code := deref(status.Code); code != "Ok" {
		return fmt.Errorf("sms: failed to send sms, code: %s, sn: %s, message: %s", code, deref(status.SerialNo), deref(status.Message))
	}

	return nil
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}
