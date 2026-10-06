package sms

import "context"

type Driver interface {
	Send(ctx context.Context, phone string, message Message) error
}
