// Package avatars fetches avatars from Gravatar and QQ.
package avatars

import (
	"sync"
	"time"

	"github.com/imroc/req/v3"
)

// maxResponseSize bounds memory per fetch; real avatars are far smaller.
const maxResponseSize = 8 << 20

var client = sync.OnceValue(func() *req.Client {
	c := req.C()
	c.SetTimeout(5 * time.Second)
	c.SetCommonRetryCount(2)
	c.SetMaxResponseSize(maxResponseSize)
	c.ImpersonateSafari()
	return c
})
