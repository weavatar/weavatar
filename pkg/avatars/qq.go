package avatars

import (
	"fmt"

	"github.com/davidbyttow/govips/v2/vips"
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

	/**
	 * 有一部分 QQ 头像可能是因为腾讯服务器 BUG 的原因，导致在 100 清晰度下是最佳显示效果，但是在 640 清晰度下则显示出了几十分辨率的屎。
	 * 还有部分 QQ 头没 640 尺寸的图片，这时候尝试获取 100 尺寸的。
	 *
	 * 比如：
	 * http://q1.qlogo.cn/g?b=qq&nk=1327444568&s=100
	 * http://q1.qlogo.cn/g?b=qq&nk=1327444568&s=640
	 *
	 * 所以这里判断一下，如果通过 640 尺寸获取到的图的实际大小小于 100 则转而获取尺寸为 100 的图
	 */
	image, err := vips.NewImageFromBuffer(resp.Bytes())
	if err == nil {
		defer image.Close()
	}
	if err != nil || (image.Width() < 100 || image.Height() < 100) {
		resp, err = client().R().SetQueryParams(map[string]string{
			"b":  "qq",
			"nk": qq,
			"s":  "100",
		}).Get("https://q.qlogo.cn/g")
		if err != nil {
			return nil, err
		}
		if !resp.IsSuccessState() {
			return nil, fmt.Errorf("failed to get QQ 100 avatar: %s", resp.String())
		}
	}

	return resp.Bytes(), nil
}
