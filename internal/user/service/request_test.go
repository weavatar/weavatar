package service_test

import (
	"strings"
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/shared/rule"
	"github.com/weavatar/weavatar/internal/user/service"
)

// TestCheckRules catches invalid validate tags at test time.
func TestCheckRules(t *testing.T) {
	v := newValidator(t)

	check.NoError(t, v.Check[service.UserCallback]())
	check.NoError(t, v.Check[service.UserUpdate]())
}

func TestUserCallbackRules(t *testing.T) {
	v := newValidator(t)

	cases := []struct {
		name string
		req  service.UserCallback
		want bool
	}{
		{"valid", service.UserCallback{Code: "c", State: "login-x"}, true},
		{"missing code", service.UserCallback{State: "login-x"}, false},
		{"missing state", service.UserCallback{Code: "c"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := v.Valid(t.Context(), &tc.req)
			must.NoError(t, err)
			check.Equal(t, ok, tc.want)
		})
	}
}

func TestUserUpdateRules(t *testing.T) {
	v := newValidator(t)
	long := strings.Repeat("a", 256)

	cases := []struct {
		name string
		req  service.UserUpdate
		want bool
	}{
		{"valid", service.UserUpdate{Nickname: "alice", Avatar: "https://weavatar.com/avatar/?d=mp"}, true},
		{"missing nickname", service.UserUpdate{Avatar: "https://weavatar.com/a.png"}, false},
		{"missing avatar", service.UserUpdate{Nickname: "alice"}, false},
		{"avatar without scheme", service.UserUpdate{Nickname: "alice", Avatar: "weavatar.com/a.png"}, false},
		{"plain http avatar", service.UserUpdate{Nickname: "alice", Avatar: "http://weavatar.com/a.png"}, true},
		{"javascript avatar", service.UserUpdate{Nickname: "alice", Avatar: "javascript://weavatar.com/%0aalert(1)"}, false},
		{"nickname too long", service.UserUpdate{Nickname: long, Avatar: "https://weavatar.com/a.png"}, false},
		{"avatar too long", service.UserUpdate{Nickname: "alice", Avatar: "https://weavatar.com/" + long}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, err := v.Valid(t.Context(), &tc.req)
			must.NoError(t, err)
			check.Equal(t, ok, tc.want)
		})
	}
}

// newValidator mirrors the production validator's rule set and strictness.
func newValidator(t *testing.T) *validator.Validator {
	t.Helper()

	opts := append([]validator.Option{validator.WithStrictRequired()}, rule.Options(nil, nil, nil, true)...)
	v, err := validator.New(opts...)
	must.NoError(t, err)
	return v
}
