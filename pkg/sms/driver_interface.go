package sms

type Driver interface {
	Send(phone string, message Message) error
}
