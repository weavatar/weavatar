package avatars

import (
	"fmt"
	"os"
	"sync"
)

var gravatarBase = sync.OnceValue(func() string {
	if v, ok := os.LookupEnv("GRAVATAR_URL"); ok {
		return v
	}
	return "https://gravatar.com"
})()

func Gravatar(hash string) ([]byte, error) {
	resp, err := client.R().SetQueryParams(map[string]string{
		"r": "g",
		"d": "404",
		"s": "1000",
	}).Get(gravatarBase + "/avatar/" + hash)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("failed to get Gravatar avatar: %s", resp.String())
	}

	return resp.Bytes(), nil
}
