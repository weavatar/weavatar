package audit

import (
	"sync"

	"github.com/knadh/koanf/v2"
)

var (
	instance *Audit
	once     sync.Once
)

type Audit struct {
	driver Driver
}

func NewAudit(conf *koanf.Koanf) *Audit {
	once.Do(func() {
		switch conf.MustString("audit.driver") {
		case "aliyun":
			instance = &Audit{
				driver: NewAliyun(conf.MustString("audit.aliyun.accessKeyId"), conf.MustString("audit.aliyun.accessKeySecret")),
			}
		case "cos":
			instance = &Audit{
				driver: NewCOS(conf.MustString("audit.cos.secretId"), conf.MustString("audit.cos.secretKey"), conf.MustString("audit.cos.bucket")),
			}
		}
		panic("failed to initialize image audit, unsupported driver: " + conf.MustString("audit.driver"))
	})

	return instance
}

// Check 检查图片是否违规 true: 违规 false: 未违规
func (c *Audit) Check(url string) (bool, error) {
	return c.driver.Check(url)
}
