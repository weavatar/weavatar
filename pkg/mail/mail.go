package mail

import (
	"crypto/tls"

	"github.com/wneessen/go-mail"
)

type Mail struct {
	host           string
	port           int
	user, password string
}

func New(host string, port int, user, password string) *Mail {
	return &Mail{
		host:     host,
		port:     port,
		user:     user,
		password: password,
	}
}

func (r *Mail) Send(to, subject, content string) error {
	message := mail.NewMsg()
	if err := message.From(r.user); err != nil {
		return err
	}
	if err := message.To(to); err != nil {
		return err
	}

	message.Subject(subject)
	message.SetBodyString(mail.TypeTextHTML, content)

	client, err := mail.NewClient(r.host,
		mail.WithPort(r.port),
		mail.WithSMTPAuth(mail.SMTPAuthPlain),
		mail.WithTLSPolicy(mail.TLSOpportunistic),
		mail.WithTLSPortPolicy(mail.TLSOpportunistic),
		mail.WithTLSConfig(&tls.Config{InsecureSkipVerify: true}),
		mail.WithUsername(r.user), mail.WithPassword(r.password))
	if err != nil {
		return err
	}
	return client.DialAndSend(message)
}
