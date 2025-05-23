package cdn

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/cloudflare/cloudflare-go/v4"
	"github.com/cloudflare/cloudflare-go/v4/cache"
	"github.com/cloudflare/cloudflare-go/v4/option"
	"github.com/dromara/carbon/v2"
	"github.com/imroc/req/v3"
)

type CloudFlare struct {
	apiKey, apiEmail string // 密钥
	zoneID           string // 域名标识
}

// CloudFlareGraphQLQuery 结构体用于构造 GraphQL 查询
type CloudFlareGraphQLQuery struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

// CloudFlareHttpRequests 结构体用于解析 GraphQL 查询结果
type CloudFlareHttpRequests struct {
	Data struct {
		Viewer struct {
			Zones []struct {
				HttpRequests1DGroups []struct {
					Sum struct {
						Requests int `json:"requests"`
					} `json:"sum"`
				} `json:"httpRequests1dGroups"`
			} `json:"zones"`
		} `json:"viewer"`
	} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// RefreshUrl 刷新URL
func (s *CloudFlare) RefreshUrl(urls []string) error {
	client := cloudflare.NewClient(
		option.WithAPIKey(s.apiKey),
		option.WithAPIEmail(s.apiEmail),
	)

	for i, url := range urls {
		urls[i] = strings.TrimPrefix(url, "https://")
		urls[i] = strings.TrimPrefix(url, "http://")
	}

	var newUrls cache.CachePurgeParamsBodyCachePurgeSingleFile
	newUrls.Files = cloudflare.F(urls)

	resp, err := client.Cache.Purge(context.Background(), cache.CachePurgeParams{
		ZoneID: cloudflare.F(s.zoneID),
		Body:   newUrls,
	})
	if err != nil {
		return err
	}
	if resp.ID == "" {
		return fmt.Errorf("cdn: fail to refresh cloudflare url: %s", resp.JSON.RawJSON())
	}

	return nil
}

// RefreshPath 刷新路径
func (s *CloudFlare) RefreshPath(paths []string) error {
	return s.RefreshUrl(paths)
}

// GetUsage 获取用量
func (s *CloudFlare) GetUsage(domain string, startTime, endTime *carbon.Carbon) (uint, error) {
	client := req.C()
	client.SetBaseURL("https://api.cloudflare.com/client/v4")
	client.SetTimeout(10 * time.Second)
	client.SetCommonRetryCount(2)
	client.SetCommonHeaders(map[string]string{
		"X-Auth-Email": s.apiEmail,
		"X-Auth-Key":   s.apiKey,
	})

	query := CloudFlareGraphQLQuery{
		Query: `
		{
		  viewer {
			zones(filter: {zoneTag: $zoneTag}) {
			  httpRequests1dGroups(limit: 1, filter: {date_gt: $start, date_lt: $end}) {
				sum {
				  requests
				}
			  }
			}
		  }
		}
        `,
		Variables: map[string]any{
			"zoneTag": s.zoneID,
			// CloudFlare 不这样写的话取不到数据
			"start": startTime.SubDay().ToDateString(),
			"end":   endTime.ToDateString(),
		},
	}

	var resp CloudFlareHttpRequests
	_, err := client.R().SetBodyJsonMarshal(query).SetSuccessResult(&resp).SetErrorResult(&resp).Post("/graphql")
	if err != nil {
		return 0, err
	}

	// 数据可能为空，需要判断
	if len(resp.Data.Viewer.Zones) == 0 || len(resp.Data.Viewer.Zones[0].HttpRequests1DGroups) == 0 {
		return 0, fmt.Errorf("cdn: fail to get cloudflare usage: %v", resp.Errors)
	}

	return uint(resp.Data.Viewer.Zones[0].HttpRequests1DGroups[0].Sum.Requests), nil
}
