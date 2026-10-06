package cdn

import (
	"context"
	"fmt"
	"time"

	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/starshield/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/starshield/client"
)

// StarShield is JD Cloud StarShield; its SDK takes no context, so ctx is only
// checked before each call.
type StarShield struct {
	accessKey, secretKey string
	instanceID           string
	zoneID               string
}

func (s *StarShield) RefreshUrl(ctx context.Context, urls []string) error {
	return s.purge(ctx, "url", urls)
}

func (s *StarShield) RefreshPath(ctx context.Context, paths []string) error {
	return s.purge(ctx, "path", paths)
}

func (s *StarShield) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	request := apis.NewZoneRequestSumRequest(s.zoneID, "all", domain, startTime.Format(time.DateOnly)+"T00:00:00.000Z", endTime.Format(time.DateOnly)+"T00:00:00.000Z")
	request.AddHeader("x-jdcloud-account-id", s.instanceID)

	resp, err := s.client().ZoneRequestSum(request)
	if err != nil {
		return 0, err
	}
	if resp.Error.Code != 0 {
		return 0, fmt.Errorf("cdn: fail to get starshield usage, code: %d, status: %s, message: %s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return uint(resp.Result.Value), nil
}

func (s *StarShield) purge(ctx context.Context, kind string, prefixes []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	request := apis.NewPurgeFilesByCache_TagsAndHostOrPrefixRequest(s.zoneID)
	request.AddHeader("x-jdcloud-account-id", s.instanceID)
	request.SetPrefixes(prefixes)

	resp, err := s.client().PurgeFilesByCache_TagsAndHostOrPrefix(request)
	if err != nil {
		return err
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("cdn: fail to refresh starshield %s, code: %d, status: %s, message: %s", kind, resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}

func (s *StarShield) client() *client.StarshieldClient {
	c := client.NewStarshieldClient(core.NewCredentials(s.accessKey, s.secretKey))
	c.DisableLogger()
	return c
}
