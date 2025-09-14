//go:build !cgo

package avatars

import (
	"fmt"
)

func Qq(qq string) ([]byte, error) {
	resp, err := client().R().SetQueryParams(map[string]string{
		"b":  "qq",
		"nk": qq,
		"s":  "640",
	}).Get("https://q.qlogo.cn/g")
	if err != nil {
		return nil, err
	}
	if !resp.IsSuccessState() {
		return nil, fmt.Errorf("failed to get QQ 640 avatar: %s", resp.String())
	}

	return resp.Bytes(), nil
}
