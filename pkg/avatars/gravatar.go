package avatars

import "fmt"

func Gravatar(hash string) ([]byte, error) {
	resp, err := client.R().SetQueryParams(map[string]string{
		"r": "g",
		"d": "404",
		"s": "600",
	}).Get("https://gravatar.com/avatar/" + hash)
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("failed to get Gravatar avatar: %s", resp.String())
	}

	return resp.Bytes(), nil
}
