// Package data adapts the SMS and mail clients to the verifycode ports.
package data

import (
	"context"

	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/verifycode/biz"
	"github.com/weavatar/weavatar/pkg/mail"
	"github.com/weavatar/weavatar/pkg/sms"
)

const brand = "WeAvatar"

type smsSender struct {
	client *sms.SMS
}

func NewSMSSender(client *sms.SMS) biz.SMSSender {
	return &smsSender{client: client}
}

func (s *smsSender) Send(ctx context.Context, phone, code string) error {
	if err := s.client.Send(ctx, phone, sms.Message{Data: map[string]string{"code": code}}); err != nil {
		return oops.In("verifycode").Wrapf(err, "send sms code")
	}

	return nil
}

type mailSender struct {
	client *mail.Mail
}

func NewMailSender(client *mail.Mail) biz.MailSender {
	return &mailSender{client: client}
}

func (m *mailSender) Send(ctx context.Context, to, code string) error {
	if err := m.client.Send(ctx, to, "验证码", mail.CodeTmpl(brand, code)); err != nil {
		return oops.In("verifycode").Wrapf(err, "send mail code")
	}

	return nil
}
