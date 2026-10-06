package service_test

import (
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/avatar/service"
	"github.com/weavatar/weavatar/internal/shared/rule"
	"github.com/weavatar/weavatar/internal/shared/transport"
)

func TestCheckRules(t *testing.T) {
	opts := append([]validator.Option{validator.WithStrictRequired()}, rule.Options(nil, nil, nil, true)...)
	v, err := validator.New(opts...)
	must.NoError(t, err)

	check.NoError(t, v.Check[transport.Paginate]())
	check.NoError(t, v.Check[service.Avatar]())
	check.NoError(t, v.Check[service.AvatarCreate]())
	check.NoError(t, v.Check[service.AvatarUpdate]())
	check.NoError(t, v.Check[service.AvatarDelete]())
	check.NoError(t, v.Check[service.AvatarCheck]())
	check.NoError(t, v.Check[service.AvatarQq]())
}
