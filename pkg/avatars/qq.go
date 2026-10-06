package avatars

import (
	"context"
	"fmt"

	"github.com/weavatar/weavatar/pkg/imaging"
)

// Qq falls back to the 100px avatar when qlogo serves the 640px one missing,
// undecodable or undersized; the fallback must decode too.
func Qq(ctx context.Context, qq string) ([]byte, error) {
	img, err := fetchQq(ctx, qq, "640")
	if err != nil {
		return nil, err
	}
	if decoded, _, err := imaging.Decode(img); err == nil && decoded.Bounds().Dx() >= 100 && decoded.Bounds().Dy() >= 100 {
		return img, nil
	}

	if img, err = fetchQq(ctx, qq, "100"); err != nil {
		return nil, err
	}
	if _, _, err = imaging.Decode(img); err != nil {
		return nil, err
	}
	return img, nil
}

func fetchQq(ctx context.Context, qq, size string) ([]byte, error) {
	resp, err := client().R().SetContext(ctx).SetQueryParams(map[string]string{
		"b":  "qq",
		"nk": qq,
		"s":  size,
	}).Get("https://q.qlogo.cn/g")
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("qq: unexpected status %d for %spx avatar", resp.StatusCode, size)
	}
	return resp.Bytes(), nil
}
