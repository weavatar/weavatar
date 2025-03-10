package audit

import (
	"sync"

	"github.com/knadh/koanf/v2"
)

type Audit struct {
	driver Driver
}

func NewAudit(conf *koanf.Koanf) *Audit {
	return sync.OnceValue(func() *Audit {
		switch conf.MustString("audit.driver") {
		case "aliyun":
			return &Audit{
				driver: NewAliyun(conf.MustString("audit.aliyun.accessKeyId"), conf.MustString("audit.aliyun.accessKeySecret")),
			}
		case "cos":
			return &Audit{
				driver: NewCOS(conf.MustString("audit.cos.secretId"), conf.MustString("audit.cos.secretKey"), conf.MustString("audit.cos.bucket")),
			}
		}
		panic("failed to initialize image audit, unsupported driver: " + conf.MustString("audit.driver"))
	})()
}

// Check 检查图片是否违规 true: 违规 false: 未违规
func (c *Audit) Check(url string) (bool, error) {
	return c.driver.Check(url)
}
