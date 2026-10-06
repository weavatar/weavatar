package cdn

import (
	"context"
	"errors"
	"fmt"
	"time"
)

// Config enables Drivers in the listed order.
type Config struct {
	Drivers    []string         `koanf:"drivers"`
	BaiShan    BaiShanConfig    `koanf:"baishan"`
	Cloudflare CloudflareConfig `koanf:"cloudflare"`
	CTYun      CTYunConfig      `koanf:"ctyun"`
	HuaWei     HuaWeiConfig     `koanf:"huawei"`
	EdgeOne    EdgeOneConfig    `koanf:"edgeone"`
	StarShield StarShieldConfig `koanf:"starshield"`
	UpYun      UpYunConfig      `koanf:"upyun"`
	WafPro     WafProConfig     `koanf:"wafpro"`
	WjDun      WjDunConfig      `koanf:"wjdun"`
	YunDun     YunDunConfig     `koanf:"yundun"`
}

type BaiShanConfig struct {
	Token string `koanf:"token"`
}

type CloudflareConfig struct {
	APIKey   string `koanf:"api_key"`
	APIEmail string `koanf:"api_email"`
	ZoneID   string `koanf:"zone_id"`
}

type CTYunConfig struct {
	AppID     string `koanf:"app_id"`
	AppSecret string `koanf:"app_secret"`
}

type HuaWeiConfig struct {
	AccessKey string `koanf:"access_key"`
	SecretKey string `koanf:"secret_key"`
}

type EdgeOneConfig struct {
	SecretID  string `koanf:"secret_id"`
	SecretKey string `koanf:"secret_key"`
}

type StarShieldConfig struct {
	AccessKey  string `koanf:"access_key"`
	SecretKey  string `koanf:"secret_key"`
	InstanceID string `koanf:"instance_id"`
	ZoneID     string `koanf:"zone_id"`
}

type UpYunConfig struct {
	Token string `koanf:"token"`
}

type WafProConfig struct {
	APIKey    string `koanf:"api_key"`
	APISecret string `koanf:"api_secret"`
}

type WjDunConfig struct {
	APIKey    string `koanf:"api_key"`
	APISecret string `koanf:"api_secret"`
}

type YunDunConfig struct {
	Username string `koanf:"username"`
	Password string `koanf:"password"`
}

type Cdn struct {
	drivers []Driver
}

func New(config Config) (*Cdn, error) {
	drivers := make([]Driver, 0, len(config.Drivers))
	for _, name := range config.Drivers {
		var driver Driver
		switch name {
		case "baishan":
			driver = newBaiShan(config.BaiShan)
		case "cloudflare":
			driver = newCloudFlare(config.Cloudflare)
		case "ctyun":
			driver = newCTYun(config.CTYun)
		case "huawei":
			driver = &HuaWei{accessKey: config.HuaWei.AccessKey, secretKey: config.HuaWei.SecretKey}
		case "edgeone":
			driver = &EdgeOne{secretId: config.EdgeOne.SecretID, secretKey: config.EdgeOne.SecretKey}
		case "starshield":
			driver = &StarShield{
				accessKey:  config.StarShield.AccessKey,
				secretKey:  config.StarShield.SecretKey,
				instanceID: config.StarShield.InstanceID,
				zoneID:     config.StarShield.ZoneID,
			}
		case "upyun":
			driver = newUpYun(config.UpYun)
		case "wafpro":
			driver = newWafPro("wafpro", "https://scdn.console.waf.pro", config.WafPro.APIKey, config.WafPro.APISecret)
		case "wjdun":
			driver = newWafPro("wjdun", "https://user.wjdun.cn", config.WjDun.APIKey, config.WjDun.APISecret)
		case "yundun":
			driver = newYunDun(config.YunDun)
		default:
			return nil, fmt.Errorf("cdn: unsupported driver %q", name)
		}
		drivers = append(drivers, driver)
	}

	return &Cdn{drivers: drivers}, nil
}

// RefreshUrl keeps going past a failing driver and joins the errors.
func (c *Cdn) RefreshUrl(ctx context.Context, urls []string) error {
	return c.each(ctx, func(d Driver) error { return d.RefreshUrl(ctx, urls) })
}

// RefreshPath keeps going past a failing driver and joins the errors.
func (c *Cdn) RefreshPath(ctx context.Context, paths []string) error {
	return c.each(ctx, func(d Driver) error { return d.RefreshPath(ctx, paths) })
}

// GetUsage sums every driver's count, stopping at the first error; drivers
// format dates in start's location, so pass the zone the days belong to.
func (c *Cdn) GetUsage(ctx context.Context, domain string, start, end time.Time) (uint, error) {
	var total uint
	for _, driver := range c.drivers {
		usage, err := driver.GetUsage(ctx, domain, start, end)
		if err != nil {
			return 0, err
		}
		total += usage
	}
	return total, nil
}

func (c *Cdn) each(ctx context.Context, fn func(Driver) error) error {
	var errs []error
	for _, driver := range c.drivers {
		if err := ctx.Err(); err != nil {
			return errors.Join(append(errs, err)...)
		}
		if err := fn(driver); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
