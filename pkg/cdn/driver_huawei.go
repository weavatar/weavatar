package cdn

import (
	"context"
	"fmt"
	"time"

	"github.com/devhaozi/huaweicloud-sdk-go-v3/core/auth/global"
	cdn "github.com/devhaozi/huaweicloud-sdk-go-v3/services/cdn/v2"
	"github.com/devhaozi/huaweicloud-sdk-go-v3/services/cdn/v2/model"
	"github.com/devhaozi/huaweicloud-sdk-go-v3/services/cdn/v2/region"
	"github.com/spf13/cast"
)

// HuaWei is Huawei Cloud CDN. Its SDK takes no context, so ctx is only
// checked before each call.
type HuaWei struct {
	accessKey, secretKey string
}

func (r *HuaWei) RefreshUrl(ctx context.Context, urls []string) error {
	return r.refresh(ctx, model.GetRefreshTaskRequestBodyTypeEnum().PREFIX, "url", urls)
}

func (r *HuaWei) RefreshPath(ctx context.Context, paths []string) error {
	return r.refresh(ctx, model.GetRefreshTaskRequestBodyTypeEnum().DIRECTORY, "path", paths)
}

func (r *HuaWei) GetUsage(ctx context.Context, domain string, startTime, endTime time.Time) (uint, error) {
	if err := ctx.Err(); err != nil {
		return 0, err
	}

	client, err := r.client()
	if err != nil {
		return 0, err
	}

	request := &model.ShowDomainStatsRequest{}
	request.Action = "summary"
	request.StartTime = startTime.UnixMilli()
	request.EndTime = endTime.UnixMilli()
	request.DomainName = domain
	request.StatType = "req_num"
	response, err := client.ShowDomainStats(request)
	if err != nil {
		return 0, err
	}

	if response.HttpStatusCode != 200 {
		return 0, fmt.Errorf("cdn: fail to get huawei usage: %v", response.Result)
	}

	if v, ok := response.Result["req_num"]; ok {
		return cast.ToUint(v), nil
	}

	return 0, nil
}

func (r *HuaWei) refresh(ctx context.Context, typ model.RefreshTaskRequestBodyType, kind string, urls []string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	client, err := r.client()
	if err != nil {
		return err
	}

	mode := model.GetRefreshTaskRequestBodyModeEnum().ALL
	request := &model.CreateRefreshTasksRequest{
		Body: &model.RefreshTaskRequest{
			RefreshTask: &model.RefreshTaskRequestBody{
				Type: &typ,
				Mode: &mode,
				Urls: urls,
			},
		},
	}

	response, err := client.CreateRefreshTasks(request)
	if err != nil {
		return err
	}

	if response.HttpStatusCode != 200 {
		task := ""
		if response.RefreshTask != nil {
			task = *response.RefreshTask
		}
		return fmt.Errorf("cdn: fail to refresh huawei %s: %s", kind, task)
	}

	return nil
}

func (r *HuaWei) client() (*cdn.CdnClient, error) {
	auth, err := global.NewCredentialsBuilder().
		WithAk(r.accessKey).
		WithSk(r.secretKey).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	build, err := cdn.CdnClientBuilder().
		WithRegion(region.CN_NORTH_1).
		WithCredential(auth).
		SafeBuild()
	if err != nil {
		return nil, err
	}

	return cdn.NewCdnClient(build), nil
}
