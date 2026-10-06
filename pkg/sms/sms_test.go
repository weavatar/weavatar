package sms

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
)

type fakeDriver struct {
	err   error
	calls []string
}

func (f *fakeDriver) Send(_ context.Context, phone string, _ Message) error {
	f.calls = append(f.calls, phone)
	return f.err
}

func newTestSMS(aliyun, tencent Driver, clock *time.Time) *SMS {
	return &SMS{
		aliyun:  aliyun,
		tencent: tencent,
		now:     func() time.Time { return *clock },
		sent:    make(map[string]time.Time),
	}
}

func TestNew(t *testing.T) {
	s := New(Config{
		Aliyun:  AliyunConfig{AccessKeyID: "ak", SignName: "sign"},
		Tencent: TencentConfig{SecretID: "id", SDKAppID: "app"},
	}, 5*time.Minute)

	ali, ok := s.aliyun.(*Aliyun)
	must.True(t, ok)
	check.Equal(t, ali.accessKeyId, "ak")
	check.Equal(t, ali.signName, "sign")

	tc, ok := s.tencent.(*Tencent)
	must.True(t, ok)
	check.Equal(t, tc.secretId, "id")
	check.Equal(t, tc.sdkAppId, "app")
	check.Equal(t, tc.expireTime, "5")
}

func TestSendPrefersAliyunThenResendsViaTencent(t *testing.T) {
	clock := time.Unix(1000, 0)
	ali, tc := &fakeDriver{}, &fakeDriver{}
	s := newTestSMS(ali, tc, &clock)
	ctx := t.Context()

	must.NoError(t, s.Send(ctx, "13800000000", Message{}))
	check.Len(t, ali.calls, 1)
	check.Len(t, tc.calls, 0)

	// a resend within the window goes through Tencent and clears the entry
	clock = clock.Add(time.Minute)
	must.NoError(t, s.Send(ctx, "13800000000", Message{}))
	check.Len(t, ali.calls, 1)
	check.Len(t, tc.calls, 1)

	// with the entry cleared, the next send is back on Aliyun
	must.NoError(t, s.Send(ctx, "13800000000", Message{}))
	check.Len(t, ali.calls, 2)
	check.Len(t, tc.calls, 1)

	// past the window the entry has expired, so Aliyun again
	clock = clock.Add(resendWindow)
	must.NoError(t, s.Send(ctx, "13800000000", Message{}))
	check.Len(t, ali.calls, 3)
	check.Len(t, tc.calls, 1)
}

func TestSendFallsBackToTencent(t *testing.T) {
	clock := time.Unix(1000, 0)
	aliErr := errors.New("aliyun down")
	ali, tc := &fakeDriver{err: aliErr}, &fakeDriver{}
	s := newTestSMS(ali, tc, &clock)

	must.NoError(t, s.Send(t.Context(), "13800000000", Message{}))
	check.Len(t, tc.calls, 1)
	check.Len(t, s.sent, 1)

	tcErr := errors.New("tencent down")
	tc.err = tcErr
	err := s.Send(t.Context(), "13900000000", Message{})
	must.ErrorIs(t, err, aliErr)
	must.ErrorIs(t, err, tcErr)
}

func TestLazyCleanup(t *testing.T) {
	clock := time.Unix(1000, 0)
	s := newTestSMS(&fakeDriver{}, &fakeDriver{}, &clock)

	must.NoError(t, s.Send(t.Context(), "a", Message{}))
	must.NoError(t, s.Send(t.Context(), "b", Message{}))
	check.Len(t, s.sent, 2)

	clock = clock.Add(resendWindow)
	must.NoError(t, s.Send(t.Context(), "c", Message{}))
	check.Len(t, s.sent, 1)
	_, ok := s.sent["c"]
	check.True(t, ok)
}
