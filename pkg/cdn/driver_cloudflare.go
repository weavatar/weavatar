package cdn

import (
	"context"
	"fmt"
	"time"

	"github.com/cloudflare/cloudflare-go/v7"
	"github.com/cloudflare/cloudflare-go/v7/cache"
	"github.com/cloudflare/cloudflare-go/v7/option"
	"github.com/imroc/req/v3"
)

type CloudFlare struct {
	zoneID string
	sdk    *cloudflare.Client // purges
	client *req.Client        // GraphQL analytics, which the SDK lacks
}

type CloudFlareGraphQLQuery struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables"`
}

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

func newCloudFlare(c CloudflareConfig) *CloudFlare {
	return &CloudFlare{
		zoneID: c.ZoneID,
		sdk: cloudflare.NewClient(
			option.WithAPIKey(c.APIKey),
			option.WithAPIEmail(c.APIEmail),
			option.WithRequestTimeout(requestTimeout),
		),
		client: newClient().
			SetBaseURL("https://api.cloudflare.com/client/v4").
			SetCommonRetryCount(2).
			SetCommonHeaders(map[string]string{
				"X-Auth-Email": c.APIEmail,
				"X-Auth-Key":   c.APIKey,
			}),
	}
}

func (s *CloudFlare) RefreshUrl(ctx context.Context, urls []string) error {
	// Cloudflare wants full URLs, scheme included
	resp, err := s.sdk.Cache.Purge(ctx, cache.CachePurgeParams{
		ZoneID: cloudflare.F(s.zoneID),
		Body:   cache.CachePurgeParamsBodyCachePurgeSingleFile{Files: cloudflare.F(urls)},
	})
	if err != nil {
		return err
	}
	if resp.ID == "" {
		return fmt.Errorf("cdn: fail to refresh cloudflare url: %s", resp.JSON.RawJSON())
	}

	return nil
}

func (s *CloudFlare) RefreshPath(ctx context.Context, paths []string) error {
	return s.RefreshUrl(ctx, paths)
}

func (s *CloudFlare) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
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
			// date_gt excludes the start day, so widen it by one
			"start": startTime.AddDate(0, 0, -1).Format(time.DateOnly),
			"end":   endTime.Format(time.DateOnly),
		},
	}

	var resp CloudFlareHttpRequests
	_, err := s.client.R().SetContext(ctx).SetBodyJsonMarshal(query).SetSuccessResult(&resp).SetErrorResult(&resp).Post("/graphql")
	if err != nil {
		return 0, err
	}

	if len(resp.Data.Viewer.Zones) == 0 || len(resp.Data.Viewer.Zones[0].HttpRequests1DGroups) == 0 {
		return 0, fmt.Errorf("cdn: fail to get cloudflare usage: %v", resp.Errors)
	}

	return uint(resp.Data.Viewer.Zones[0].HttpRequests1DGroups[0].Sum.Requests), nil
}
