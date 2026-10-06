package cdn

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
)

type fakeDriver struct {
	usage    uint
	err      error
	refreshs [][]string
}

func (f *fakeDriver) RefreshUrl(_ context.Context, urls []string) error {
	f.refreshs = append(f.refreshs, urls)
	return f.err
}

func (f *fakeDriver) RefreshPath(_ context.Context, paths []string) error {
	f.refreshs = append(f.refreshs, paths)
	return f.err
}

func (f *fakeDriver) GetUsage(context.Context, string, time.Time, time.Time) (uint, error) {
	return f.usage, f.err
}

func TestNew(t *testing.T) {
	c, err := New(Config{Drivers: []string{
		"baishan", "cloudflare", "ctyun", "huawei", "edgeone",
		"starshield", "upyun", "wafpro", "wjdun", "yundun",
	}})
	must.NoError(t, err)
	must.Len(t, c.drivers, 10)
	wjdun, ok := c.drivers[8].(*WafPro)
	must.True(t, ok)
	check.Equal(t, wjdun.endpoint, "https://user.wjdun.cn") // same platform, own host

	c, err = New(Config{})
	must.NoError(t, err)
	must.Len(t, c.drivers, 0)
	must.NoError(t, c.RefreshUrl(t.Context(), []string{"https://example.com/a"}))

	_, err = New(Config{Drivers: []string{"baishan", "nope"}})
	must.ErrorContains(t, err, `unsupported driver "nope"`)
}

func TestRefreshContinuesAfterFailure(t *testing.T) {
	boom := errors.New("boom")
	a := &fakeDriver{err: boom}
	b := &fakeDriver{}
	c := &Cdn{drivers: []Driver{a, b}}

	err := c.RefreshUrl(t.Context(), []string{"u"})
	must.ErrorIs(t, err, boom)
	check.Len(t, b.refreshs, 1)

	err = c.RefreshPath(t.Context(), []string{"p"})
	must.ErrorIs(t, err, boom)
	check.Len(t, b.refreshs, 2)
}

func TestRefreshStopsOnCanceledContext(t *testing.T) {
	a := &fakeDriver{}
	c := &Cdn{drivers: []Driver{a}}

	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	must.ErrorIs(t, c.RefreshUrl(ctx, []string{"u"}), context.Canceled)
	check.Len(t, a.refreshs, 0)
}

func TestGetUsage(t *testing.T) {
	now := time.Now()
	c := &Cdn{drivers: []Driver{&fakeDriver{usage: 3}, &fakeDriver{usage: 4}}}
	total, err := c.GetUsage(t.Context(), "example.com", now.AddDate(0, 0, -1), now)
	must.NoError(t, err)
	check.Equal(t, total, uint(7))

	boom := errors.New("boom")
	c = &Cdn{drivers: []Driver{&fakeDriver{usage: 3}, &fakeDriver{err: boom}}}
	total, err = c.GetUsage(t.Context(), "example.com", now.AddDate(0, 0, -1), now)
	must.ErrorIs(t, err, boom)
	check.Equal(t, total, uint(0))
}

func TestTimeLayouts(t *testing.T) {
	loc := time.FixedZone("CST", 8*3600)
	ts := time.Date(2026, 10, 6, 0, 0, 0, 0, loc)
	check.Equal(t, ts.Format(iso8601Layout), "2026-10-06T00:00:00+08:00")
	check.Equal(t, ts.Format(iso8601MilliLayout), "2026-10-06T00:00:00+08:00")
	check.Equal(t, ts.Add(120*time.Millisecond).Format(iso8601MilliLayout), "2026-10-06T00:00:00.12+08:00")
	check.Equal(t, ts.UTC().Format(iso8601Layout), "2026-10-05T16:00:00+00:00")
}
