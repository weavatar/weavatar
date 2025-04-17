package cdn

import (
	"sync"

	"github.com/dromara/carbon/v2"
	"github.com/knadh/koanf/v2"
)

var (
	instance *Cdn
	once     sync.Once
)

type Cdn struct {
	drivers []Driver
}

func NewCdn(conf *koanf.Koanf) *Cdn {
	once.Do(func() {
		names := conf.MustStrings("cdn.driver")
		var drivers []Driver
		for _, driver := range names {
			switch driver {
			case "baishan":
				drivers = append(drivers, &BaiShan{
					token: conf.MustString("cdn.baishan.token"),
				})
			case "cloudflare":
				drivers = append(drivers, &CloudFlare{
					apiKey:   conf.MustString("cdn.cloudflare.apiKey"),
					apiEmail: conf.MustString("cdn.cloudflare.apiEmail"),
					zoneID:   conf.MustString("cdn.cloudflare.zoneID"),
				})
			case "ctyun":
				drivers = append(drivers, &CTYun{
					appID:       conf.MustString("cdn.ctyun.appID"),
					appSecret:   conf.MustString("cdn.ctyun.appSecret"),
					apiEndpoint: "https://open.ctcdn.cn",
				})
			case "huawei":
				drivers = append(drivers, &HuaWei{
					accessKey: conf.MustString("cdn.huawei.accessKey"),
					secretKey: conf.MustString("cdn.huawei.secretKey"),
				})
			case "starshield":
				drivers = append(drivers, &StarShield{
					accessKey:  conf.MustString("cdn.starshield.accessKey"),
					secretKey:  conf.MustString("cdn.starshield.secretKey"),
					instanceID: conf.MustString("cdn.starshield.instanceID"),
					zoneID:     conf.MustString("cdn.starshield.zoneID"),
				})
			case "upyun":
				drivers = append(drivers, &UpYun{
					token: conf.MustString("cdn.upyun.token"),
				})
			case "wafpro":
				drivers = append(drivers, &WafPro{
					apiKey:    conf.MustString("cdn.wafpro.apiKey"),
					apiSecret: conf.MustString("cdn.wafpro.apiSecret"),
				})
			case "yundun":
				drivers = append(drivers, &YunDun{
					username: conf.MustString("cdn.yundun.username"),
					password: conf.MustString("cdn.yundun.password"),
				})
			}
		}

		instance = &Cdn{
			drivers: drivers,
		}
	})

	return instance
}

func (c *Cdn) RefreshUrl(urls []string) error {
	for _, driver := range c.drivers {
		if err := driver.RefreshUrl(urls); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cdn) RefreshPath(paths []string) error {
	for _, driver := range c.drivers {
		if err := driver.RefreshPath(paths); err != nil {
			return err
		}
	}
	return nil
}

func (c *Cdn) GetUsage(domain string, start, end *carbon.Carbon) (uint, error) {
	var total uint
	for _, driver := range c.drivers {
		usage, err := driver.GetUsage(domain, start, end)
		if err != nil {
			return 0, err
		}
		total += usage
	}
	return total, nil
}
