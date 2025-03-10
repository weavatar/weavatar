package avatars

import (
	"sync"
	"time"

	"github.com/imroc/req/v3"
)

var client = sync.OnceValue(func() *req.Client {
	c := req.C()
	c.SetTimeout(5 * time.Second)
	c.SetCommonRetryCount(2)
	c.ImpersonateSafari()
	return c
})()
