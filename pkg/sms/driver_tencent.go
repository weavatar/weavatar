package sms

import (
	"fmt"

	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tencentsms "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/sms/v20210111"
)

type Tencent struct {
	SecretId, SecretKey, SignName, TemplateId, SdkAppId, ExpireTime string
}

func (r *Tencent) Send(phone string, message Message) error {
	credential := common.NewCredential(
		r.SecretId,
		r.SecretKey,
	)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = "sms.tencentcloudapi.com"
	client, _ := tencentsms.NewClient(credential, "ap-guangzhou", cpf)

	request := tencentsms.NewSendSmsRequest()
	request.PhoneNumberSet = common.StringPtrs([]string{phone})
	request.SignName = common.StringPtr(r.SignName)
	request.TemplateId = common.StringPtr(r.TemplateId)
	request.TemplateParamSet = common.StringPtrs([]string{message.Data["code"], r.ExpireTime})
	request.SmsSdkAppId = common.StringPtr(r.SdkAppId)

	response, err := client.SendSms(request)

	if err != nil {
		return err
	}

	statusSet := response.Response.SendStatusSet
	code := *statusSet[0].Code
	if code != "Ok" {
		return fmt.Errorf("sms send failed: %s, code: %s, sn: %s", *statusSet[0].Message, *statusSet[0].Code, *statusSet[0].SerialNo)
	}

	return nil
}
