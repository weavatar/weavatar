package biz

import (
	"github.com/weavatar/weavatar/internal/shared/apperr"
	"github.com/weavatar/weavatar/pkg/imaging"
)

const (
	minUploadSize = 40
	maxUploadSize = 2048
)

// formatUpload checks an upload is a square of at least minUploadSize and
// scales it down to maxUploadSize; smaller uploads keep their original bytes.
func formatUpload(b []byte) ([]byte, error) {
	img, format, err := imaging.Decode(b)
	if err != nil {
		return nil, apperr.Invalid("avatar.invalid_image", "无法识别的图片").In("avatar").Wrap(err)
	}

	bounds := img.Bounds()
	if bounds.Dx() != bounds.Dy() {
		return nil, apperr.Invalid("avatar.not_square", "头像必须是正方形图片").In("avatar").Errorf("avatar is %dx%d", bounds.Dx(), bounds.Dy())
	}
	if bounds.Dx() < minUploadSize {
		return nil, apperr.Invalid("avatar.too_small", "头像必须大于 40px").In("avatar").Errorf("avatar is %dpx", bounds.Dx())
	}
	if bounds.Dx() <= maxUploadSize {
		return b, nil
	}

	return imaging.Encode(imaging.Resize(img, maxUploadSize, maxUploadSize), format)
}
