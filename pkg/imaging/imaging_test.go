package imaging

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"image"
	"image/color"
	"math/rand/v2"
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	xdraw "golang.org/x/image/draw"
)

func testImage(w, h int) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := range h {
		for x := range w {
			img.SetNRGBA(x, y, color.NRGBA{R: uint8(x * 255 / w), G: uint8(y * 255 / h), B: 128, A: 255})
		}
	}
	return img
}

func TestEncodeDecode(t *testing.T) {
	src := testImage(64, 64)
	for format, want := range map[string]string{
		"jpeg": "jpeg", "jpg": "jpeg",
		"png":  "png",
		"gif":  "gif",
		"webp": "webp",
		"tiff": "tiff",
		"avif": "avif",
		"heic": "heic", "heif": "heic",
		"jxl": "jxl",
	} {
		t.Run(format, func(t *testing.T) {
			data, err := Encode(src, format)
			must.NoError(t, err)

			img, got, err := Decode(data)
			must.NoError(t, err)
			check.Equal(t, got, want)
			check.Equal(t, img.Bounds().Size(), src.Bounds().Size())
		})
	}
}

func TestEncodeUnsupported(t *testing.T) {
	_, err := Encode(testImage(8, 8), "bmp")
	check.ErrorIs(t, err, ErrUnsupported)
}

func TestDecodeUnsupported(t *testing.T) {
	_, _, err := Decode([]byte("not an image"))
	check.ErrorIs(t, err, ErrUnsupported)
}

func TestDecodeTooLarge(t *testing.T) {
	chunk := func(typ string, data []byte) []byte {
		b := binary.BigEndian.AppendUint32(nil, uint32(len(data)))
		b = append(b, typ...)
		b = append(b, data...)
		return binary.BigEndian.AppendUint32(b, crc32.ChecksumIEEE(append([]byte(typ), data...)))
	}
	ihdr := binary.BigEndian.AppendUint32(nil, 10000)
	ihdr = binary.BigEndian.AppendUint32(ihdr, 10000)
	ihdr = append(ihdr, 8, 6, 0, 0, 0)
	data := append([]byte("\x89PNG\r\n\x1a\n"), chunk("IHDR", ihdr)...)
	data = append(data, chunk("IEND", nil)...)

	_, _, err := Decode(data)
	check.ErrorIs(t, err, ErrTooLarge)
}

func TestResize(t *testing.T) {
	src := testImage(100, 50)
	check.Equal(t, Resize(src, 40, 40).Bounds().Size(), image.Pt(40, 40))
	check.Equal(t, Resize(src, 200, 200).Bounds().Size(), image.Pt(200, 200))
	check.Equal(t, Resize(src, 100, 50), image.Image(src))
}

func TestResizeMatchesBiLinear(t *testing.T) {
	rnd := rand.New(rand.NewPCG(1, 2))
	nrgba := image.NewNRGBA(image.Rect(0, 0, 257, 193))
	for i := range nrgba.Pix {
		nrgba.Pix[i] = uint8(rnd.IntN(256))
	}
	ycbcr := image.NewYCbCr(image.Rect(0, 0, 300, 300), image.YCbCrSubsampleRatio420)
	for _, p := range [][]uint8{ycbcr.Y, ycbcr.Cb, ycbcr.Cr} {
		for i := range p {
			p[i] = uint8(rnd.IntN(256))
		}
	}
	gray := image.NewGray(image.Rect(0, 0, 120, 120))
	for i := range gray.Pix {
		gray.Pix[i] = uint8(rnd.IntN(256))
	}
	offset := nrgba.SubImage(image.Rect(13, 7, 213, 177))

	for name, src := range map[string]image.Image{"nrgba": nrgba, "ycbcr": ycbcr, "gray": gray, "offset": offset} {
		for _, size := range []image.Point{{80, 80}, {7, 3}, {150, 150}, {640, 480}, {1, 1}} {
			t.Run(fmt.Sprintf("%s-%dx%d", name, size.X, size.Y), func(t *testing.T) {
				got, ok := Resize(src, size.X, size.Y).(*image.RGBA)
				must.True(t, ok, "Resize should return *image.RGBA")
				want := image.NewRGBA(got.Rect)
				xdraw.BiLinear.Scale(want, want.Rect, src, src.Bounds(), xdraw.Src, nil)

				maxDiff := 0
				for i := range got.Pix {
					maxDiff = max(maxDiff, int(got.Pix[i])-int(want.Pix[i]), int(want.Pix[i])-int(got.Pix[i]))
				}
				check.LessOrEqual(t, maxDiff, 2)
			})
		}
	}
}

func TestDetectISOBMFF(t *testing.T) {
	ftyp := func(major string, compat ...string) []byte {
		b := binary.BigEndian.AppendUint32(nil, uint32(16+4*len(compat)))
		b = append(b, "ftyp"+major+"\x00\x00\x00\x00"...)
		for _, c := range compat {
			b = append(b, c...)
		}
		return append(b, make([]byte, 16)...)
	}

	check.Equal(t, detect(ftyp("avif", "mif1", "miaf")), "avif")
	check.Equal(t, detect(ftyp("mif1", "avif", "miaf")), "avif")
	check.Equal(t, detect(ftyp("heic", "mif1", "heic")), "heic")
	check.Equal(t, detect(ftyp("mif1", "heic")), "heic")
	check.Equal(t, detect(ftyp("isom", "mp41")), "")
	check.Equal(t, detect(bytes.Repeat([]byte{0}, 4)), "")
}
