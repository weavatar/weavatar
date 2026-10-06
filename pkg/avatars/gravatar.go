package avatars

import (
	"context"
	"fmt"
	"strings"
)

// Gravatar fetches hash from baseURL, the origin or a mirror, and rejects
// non-image replies because callers cache the bytes for days.
func Gravatar(ctx context.Context, baseURL, hash string) ([]byte, error) {
	resp, err := client().R().SetContext(ctx).SetQueryParams(map[string]string{
		"r": "g",
		"d": "404",
		"s": "1000",
	}).Get(strings.TrimRight(baseURL, "/") + "/avatar/" + hash)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("gravatar: unexpected status %d", resp.StatusCode)
	}
	if ct := resp.GetContentType(); !strings.HasPrefix(ct, "image/") {
		return nil, fmt.Errorf("gravatar: unexpected content type %q", ct)
	}

	return resp.Bytes(), nil
}
