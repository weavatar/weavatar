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
					Token: conf.MustString("cdn.baishan.token"),
				})
			case "cloudflare":
				drivers = append(drivers, &CloudFlare{
					APIKey:   conf.MustString("cdn.cloudflare.apiKey"),
					APIEmail: conf.MustString("cdn.cloudflare.apiEmail"),
					ZoneID:   conf.MustString("cdn.cloudflare.zoneID"),
				})
			case "huawei":
				drivers = append(drivers, &HuaWei{
					AccessKey: conf.MustString("cdn.huawei.accessKey"),
					SecretKey: conf.MustString("cdn.huawei.secretKey"),
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

func (c *Cdn) GetUsage(domain string, start, end carbon.Carbon) (uint, error) {
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
