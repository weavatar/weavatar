package audit

import (
	"context"
	"fmt"
)

// Config picks Driver: aliyun or cos.
type Config struct {
	Driver string       `koanf:"driver"`
	Aliyun AliyunConfig `koanf:"aliyun"`
	COS    COSConfig    `koanf:"cos"`
}

type AliyunConfig struct {
	AccessKeyID     string `koanf:"access_key_id"`
	AccessKeySecret string `koanf:"access_key_secret"`
}

type COSConfig struct {
	SecretID  string `koanf:"secret_id"`
	SecretKey string `koanf:"secret_key"`
	Bucket    string `koanf:"bucket"`
}

type Audit struct {
	driver Driver
}

func New(config Config) (*Audit, error) {
	switch config.Driver {
	case "aliyun":
		return &Audit{driver: NewAliyun(config.Aliyun.AccessKeyID, config.Aliyun.AccessKeySecret)}, nil
	case "cos":
		return &Audit{driver: NewCOS(config.COS.SecretID, config.COS.SecretKey, config.COS.Bucket)}, nil
	default:
		return nil, fmt.Errorf("audit: unsupported driver %q", config.Driver)
	}
}

// Check moderates the image at url; remark explains the verdict.
func (c *Audit) Check(ctx context.Context, url string) (bool, string, error) {
	return c.driver.Check(ctx, url)
}
