package service_test

import (
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/cron"

	"github.com/weavatar/weavatar/internal/avatar/service"
)

func TestPurgeExpiredCacheJob_Registers(t *testing.T) {
	c, err := cron.New(cron.WithSecondsField())
	must.NoError(t, err)

	must.NoError(t, service.PurgeExpiredCacheJob(nil)(c))

	var names []string
	for e := range c.Entries() {
		names = append(names, e.Name)
	}
	check.DeepEqual(t, names, []string{"avatar.purge_expired_cache"})
}
