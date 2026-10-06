package avatars

import (
	"context"
	"fmt"
	"strings"
)

// gravatarBaseURL is a variable so tests can point it at a local server.
var gravatarBaseURL = "https://gravatar.com"

// Gravatar rejects non-image replies: callers cache the bytes for days, so a
// proxy's error page must not pass for an avatar.
func Gravatar(ctx context.Context, hash string) ([]byte, error) {
	resp, err := client().R().SetContext(ctx).SetQueryParams(map[string]string{
		"r": "g",
		"d": "404",
		"s": "1000",
	}).Get(gravatarBaseURL + "/avatar/" + hash)
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
