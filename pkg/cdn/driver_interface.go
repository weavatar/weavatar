package cdn

import "github.com/dromara/carbon/v2"

type Driver interface {
	// RefreshUrl 通过URL刷新缓存
	RefreshUrl(urls []string) error
	// RefreshPath 通过路径刷新缓存
	RefreshPath(paths []string) error
	// GetUsage 获取域名请求量
	GetUsage(domain string, startTime, endTime carbon.Carbon) (uint, error)
}
