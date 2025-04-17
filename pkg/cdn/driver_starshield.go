package cdn

import (
	"fmt"

	"github.com/dromara/carbon/v2"
	"github.com/jdcloud-api/jdcloud-sdk-go/core"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/starshield/apis"
	"github.com/jdcloud-api/jdcloud-sdk-go/services/starshield/client"
)

type StarShield struct {
	accessKey, secretKey string // 密钥
	instanceID           string // 实例ID
	zoneID               string // 域名标识
}

type StarShieldClean struct {
	Type string            `json:"type"`
	Data map[string]string `json:"data"`
}

type StarShieldRefreshResponse struct {
	Code    uint   `json:"code"`
	Message string `json:"msg"`
}

type StarShieldUsageResponse struct {
	Code    uint      `json:"code"`
	Data    [][2]uint `json:"data"`
	Message string    `json:"msg"`
}

// RefreshUrl 刷新URL
func (s *StarShield) RefreshUrl(urls []string) error {
	jdClient := client.NewStarshieldClient(core.NewCredentials(s.accessKey, s.secretKey))
	jdClient.DisableLogger()
	request := apis.NewPurgeFilesByCache_TagsAndHostOrPrefixRequest(s.zoneID)
	request.AddHeader("x-jdcloud-account-id", s.instanceID)
	request.SetPrefixes(urls)

	resp, err := jdClient.PurgeFilesByCache_TagsAndHostOrPrefix(request)
	if err != nil {
		return err
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("fail to refresh starshield url, code: %d, status: %s, message: %s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}

// RefreshPath 刷新路径
func (s *StarShield) RefreshPath(paths []string) error {
	jdClient := client.NewStarshieldClient(core.NewCredentials(s.accessKey, s.secretKey))
	jdClient.DisableLogger()
	request := apis.NewPurgeFilesByCache_TagsAndHostOrPrefixRequest(s.zoneID)
	request.AddHeader("x-jdcloud-account-id", s.instanceID)
	request.SetPrefixes(paths)

	resp, err := jdClient.PurgeFilesByCache_TagsAndHostOrPrefix(request)
	if err != nil {
		return err
	}
	if resp.Error.Code != 0 {
		return fmt.Errorf("fail to refresh starshield path, code: %d, status: %s, message: %s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return nil
}

// GetUsage 获取用量
func (s *StarShield) GetUsage(domain string, startTime, endTime *carbon.Carbon) (uint, error) {
	jdClient := client.NewStarshieldClient(core.NewCredentials(s.accessKey, s.secretKey))
	jdClient.DisableLogger()
	request := apis.NewZoneRequestSumRequest(s.zoneID, "all", domain, startTime.ToDateString()+"T00:00:00.000Z", endTime.ToDateString()+"T00:00:00.000Z")
	request.AddHeader("x-jdcloud-account-id", s.instanceID)

	resp, err := jdClient.ZoneRequestSum(request)
	if err != nil {
		return 0, err
	}
	if resp.Error.Code != 0 {
		return 0, fmt.Errorf("fail to get starshield usage, code: %d, status: %s, message: %s", resp.Error.Code, resp.Error.Status, resp.Error.Message)
	}

	return uint(resp.Result.Value), nil
}
