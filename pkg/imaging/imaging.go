// Package imaging 提供纯 Go 的图片解码、缩放和编码
package imaging

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"image"
	"image/gif"
	"io"
	"slices"

	"github.com/ericpauley/go-quantize/quantize"
	"github.com/gen2brain/gav1d/avif"
	"github.com/gen2brain/h265/heic"
	"github.com/gen2brain/jpegn"
	"github.com/gen2brain/jxl"
	"github.com/gen2brain/pngn"
	"github.com/gen2brain/vpx/webp"
	"golang.org/x/image/tiff"
)

// MaxPixels 允许解码的最大像素数
const MaxPixels = 8192 * 8192

var (
	ErrUnsupported = errors.New("imaging: unsupported image format")
	ErrTooLarge    = errors.New("imaging: image too large")
)

type codec struct {
	decode func(io.Reader) (image.Image, error)
	config func(io.Reader) (image.Config, error)
	encode func(io.Writer, image.Image) error
}

var codecs = map[string]codec{
	"jpeg": {
		decode: func(r io.Reader) (image.Image, error) { return jpegn.Decode(r) },
		config: jpegn.DecodeConfig,
		encode: func(w io.Writer, m image.Image) error { return jpegn.Encode(w, m) },
	},
	"png": {
		decode: func(r io.Reader) (image.Image, error) {
			return pngn.Decode(r, &pngn.Options{PixelLimit: MaxPixels})
		},
		config: pngn.DecodeConfig,
		encode: func(w io.Writer, m image.Image) error { return pngn.Encode(w, m) },
	},
	"gif": {
		decode: gif.Decode,
		config: gif.DecodeConfig,
		encode: func(w io.Writer, m image.Image) error {
			// 标准库默认的 Plan9 调色板没有透明色，需要量化器生成带透明色的调色板
			return gif.Encode(w, m, &gif.Options{NumColors: 256, Quantizer: quantize.MedianCutQuantizer{AddTransparent: true}})
		},
	},
	"webp": {
		decode: func(r io.Reader) (image.Image, error) {
			return webp.Decode(r, webp.Options{FrameSizeLimit: MaxPixels})
		},
		config: webp.DecodeConfig,
		encode: func(w io.Writer, m image.Image) error { return webp.Encode(w, m) },
	},
	"tiff": {
		decode: tiff.Decode,
		config: tiff.DecodeConfig,
		encode: func(w io.Writer, m image.Image) error {
			// 默认不压缩，体积过大
			return tiff.Encode(w, m, &tiff.Options{Compression: tiff.Deflate})
		},
	},
	"avif": {
		decode: func(r io.Reader) (image.Image, error) {
			return avif.Decode(r, avif.Options{FrameSizeLimit: MaxPixels})
		},
		config: avif.DecodeConfig,
		encode: func(w io.Writer, m image.Image) error {
			return avif.Encode(w, m, avif.EncodeOptions{Speed: avif.DefaultSpeed})
		},
	},
	"heic": {
		decode: func(r io.Reader) (image.Image, error) {
			return heic.Decode(r, heic.Options{FrameSizeLimit: MaxPixels})
		},
		config: heic.DecodeConfig,
		encode: func(w io.Writer, m image.Image) error { return heic.Encode(w, m) },
	},
	"jxl": {
		decode: func(r io.Reader) (image.Image, error) {
			return jxl.Decode(r, jxl.Options{FrameSizeLimit: MaxPixels})
		},
		config: jxl.DecodeConfig,
		encode: func(w io.Writer, m image.Image) error { return jxl.Encode(w, m) },
	},
}

var aliases = map[string]string{
	"jpg":  "jpeg",
	"heif": "heic",
}

// Decode 解码图片，返回图片和格式名
func Decode(b []byte) (image.Image, string, error) {
	format := detect(b)
	c, ok := codecs[format]
	if !ok {
		return nil, "", ErrUnsupported
	}

	cfg, err := c.config(bytes.NewReader(b))
	if err != nil {
		return nil, "", err
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width)*int64(cfg.Height) > MaxPixels {
		return nil, "", fmt.Errorf("%w: %dx%d", ErrTooLarge, cfg.Width, cfg.Height)
	}

	img, err := c.decode(bytes.NewReader(b))
	if err != nil {
		return nil, "", err
	}

	return img, format, nil
}

// Encode 按格式编码图片，支持 jpeg(jpg)、png、gif、webp、tiff、avif、heic(heif)、jxl
func Encode(img image.Image, format string) ([]byte, error) {
	if alias, ok := aliases[format]; ok {
		format = alias
	}
	c, ok := codecs[format]
	if !ok {
		return nil, ErrUnsupported
	}

	buf := new(bytes.Buffer)
	if err := c.encode(buf, img); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// detect 根据文件头识别图片格式
func detect(b []byte) string {
	switch {
	case bytes.HasPrefix(b, []byte("\x89PNG\r\n\x1a\n")):
		return "png"
	case bytes.HasPrefix(b, []byte("\xff\xd8")):
		return "jpeg"
	case bytes.HasPrefix(b, []byte("GIF87a")), bytes.HasPrefix(b, []byte("GIF89a")):
		return "gif"
	case len(b) >= 12 && string(b[:4]) == "RIFF" && string(b[8:12]) == "WEBP":
		return "webp"
	case bytes.HasPrefix(b, []byte("II*\x00")), bytes.HasPrefix(b, []byte("MM\x00*")):
		return "tiff"
	case bytes.HasPrefix(b, []byte("\xff\x0a")), bytes.HasPrefix(b, []byte("\x00\x00\x00\x0cJXL \r\n\x87\n")):
		return "jxl"
	case len(b) >= 12 && string(b[4:8]) == "ftyp":
		return detectISOBMFF(b)
	}

	return ""
}

// detectISOBMFF 根据 ftyp box 中的 brand 区分 AVIF 和 HEIC
func detectISOBMFF(b []byte) string {
	size := int(binary.BigEndian.Uint32(b))
	if size < 16 || size > len(b) {
		size = 16
	}

	brands := []string{string(b[8:12])}
	for i := 16; i+4 <= size; i += 4 {
		brands = append(brands, string(b[i:i+4]))
	}

	if slices.ContainsFunc(brands, func(brand string) bool {
		return brand == "avif" || brand == "avis"
	}) {
		return "avif"
	}
	if slices.ContainsFunc(brands, func(brand string) bool {
		return slices.Contains([]string{"heic", "heix", "heim", "heis", "hevc", "hevx", "hevm", "hevs", "mif1", "msf1"}, brand)
	}) {
		return "heic"
	}

	return ""
}
