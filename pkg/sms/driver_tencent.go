package sms

import (
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

func (r *Tencent) Send(phone string, message Message) error {
	credential := common.NewCredential(
		r.secretId,
		r.secretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	client, _ := tencentsms.NewClient(credential, "ap-beijing", cpf)

	request := tencentsms.NewSendSmsRequest()
	request.PhoneNumberSet = common.StringPtrs([]string{phone})
	request.SignName = common.StringPtr(r.signName)
	request.TemplateId = common.StringPtr(r.templateId)
	request.TemplateParamSet = common.StringPtrs([]string{message.Data["code"], r.expireTime})
	request.SmsSdkAppId = common.StringPtr(r.sdkAppId)

	response, err := client.SendSms(request)

	var sdkError *sdkerror.TencentCloudSDKError
	if errors.As(err, &sdkError) {
		return fmt.Errorf("sms: failed to send sms, code: %s, message: %s, requestId: %s", sdkError.Code, sdkError.Message, sdkError.RequestId)
	}
	if err != nil {
		return err
	}

	statusSet := response.Response.SendStatusSet
	code := *statusSet[0].Code
	if code != "Ok" {
		return fmt.Errorf("sms: failed to send sms, code: %s, sn: %s, message: %s", *statusSet[0].Code, *statusSet[0].SerialNo, *statusSet[0].Message)
	}

	return nil
}
