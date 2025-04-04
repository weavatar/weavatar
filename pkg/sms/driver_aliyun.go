package sms

import (
	"encoding/json"
	"fmt"

	openapi "github.com/alibabacloud-go/darabonba-openapi/v2/client"
	dysmsapi20170525 "github.com/alibabacloud-go/dysmsapi-20170525/v4/client"
	util "github.com/alibabacloud-go/tea-utils/v2/service"
	"github.com/alibabacloud-go/tea/tea"
)

type Aliyun struct {
	accessKeyId, accessKeySecret, signName, templateCode, expireTime string
}

func (r *Aliyun) Send(phone string, message Message) error {
	client, err := r.createClient()
	if err != nil {
		return err
	}

	param, err := json.Marshal(message.Data)
	if err != nil {
		return err
	}

	sendSmsRequest := &dysmsapi20170525.SendSmsRequest{
		SignName:      &r.signName,
		TemplateCode:  &r.templateCode,
		PhoneNumbers:  tea.String(phone),
		TemplateParam: tea.String(string(param)),
	}

	result, err := client.SendSmsWithOptions(sendSmsRequest, &util.RuntimeOptions{
		Autoretry:   tea.Bool(true),
		MaxAttempts: tea.Int(3),
	})
	if err != nil {
		return err
	}

	if tea.StringValue(result.Body.Message) != "OK" {
		return fmt.Errorf("sms send failed: %s, code: %s, request id: %s", tea.StringValue(result.Body.Message), tea.StringValue(result.Body.Code), tea.StringValue(result.Body.RequestId))
	}

	return nil
}

func (r *Aliyun) createClient() (*dysmsapi20170525.Client, error) {
	config := &openapi.Config{
		AccessKeyId:     &r.accessKeyId,
		AccessKeySecret: &r.accessKeySecret,
		Endpoint:        tea.String("dysmsapi.aliyuncs.com"),
	}

	return dysmsapi20170525.NewClient(config)
}
