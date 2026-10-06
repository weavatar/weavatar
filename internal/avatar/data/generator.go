package data

import (
	"bytes"
	"hash/fnv"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/png"
	"math/rand/v2"
	"strings"

	"github.com/forPelevin/gomoji"
	"github.com/weavatar/identicon"
	"github.com/weavatar/initials"
	"github.com/weavatar/monsterid"
	"github.com/weavatar/retricon"
	"github.com/weavatar/robohash"
	"github.com/weavatar/wavatar"
	"golang.org/x/image/font/opentype"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/pkg/embed"
)

const minIdenticonSize = 16

type generator struct {
	font  *opentype.Font
	emoji *opentype.Font
}

func NewGenerator() (biz.Generator, error) {
	font, err := loadFont("font/SourceHanSansSC-Bold.otf")
	if err != nil {
		return nil, err
	}
	emoji, err := loadFont("font/NotoEmoji-Bold.ttf")
	if err != nil {
		return nil, err
	}

	return &generator{font: font, emoji: emoji}, nil
}

func (g *generator) Generate(kind, seed string, size int, text string) ([]byte, error) {
	switch kind {
	case "mp", "mm", "mystery":
		return embed.DefaultFS.ReadFile("default/mp.png")
	case "blank":
		return embed.DefaultFS.ReadFile("default/blank.png")
	case "identicon":
		id, err := identicon.New(max(size, minIdenticonSize), color.White, identicon.DarkColors...)
		if err != nil {
			return nil, err
		}
		return encodePNG(id.Make([]byte(seed)))
	case "monsterid":
		return encodePNG(monsterid.New([]byte(seed)))
	case "wavatar":
		return encodePNG(wavatar.New([]byte(seed)))
	case "retro":
		img, err := retricon.New(seed, retricon.Gravatar)
		if err != nil {
			return nil, err
		}
		return encodePNG(img)
	case "robohash":
		rh, err := robohash.New([]byte(seed), "set1", "")
		if err != nil {
			return nil, err
		}
		img, err := rh.Assemble()
		if err != nil {
			return nil, err
		}
		return encodePNG(img)
	case "color":
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.Draw(img, img.Bounds(), image.NewUniform(paletteColor(seed)), image.Point{}, draw.Src)
		return encodePNG(img)
	case "letter", "initials":
		return g.initials(seed, text)
	default:
		return embed.DefaultFS.ReadFile("default/default.png")
	}
}

// initials draws up to two characters, or the first emoji alone.
func (g *generator) initials(seed, text string) ([]byte, error) {
	font, fontSize := g.font, 500
	words := []rune(strings.ToUpper(text))
	if len(words) > 1 {
		fontSize = 400
		words = words[:2]
	}
	if gomoji.FindAll(string(words)) != nil {
		font, fontSize = g.emoji, 500
		words = words[:1]
	}

	img, err := initials.Draw(1000, words, &initials.Options{
		Font:       font,
		FontSize:   fontSize,
		PaletteKey: seed, // the same hash keeps the same color
	})
	if err != nil {
		return nil, err
	}
	return encodePNG(img)
}

func loadFont(name string) (*opentype.Font, error) {
	b, err := embed.FontFS.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return opentype.Parse(b)
}

// paletteColor picks a web-safe color deterministically from seed.
func paletteColor(seed string) color.Color {
	h := fnv.New64a()
	_, _ = h.Write([]byte(seed))
	sum := h.Sum64()
	r := rand.New(rand.NewPCG(sum, (sum>>1)|1)) //nolint:gosec // deterministic pick, not security
	return palette.WebSafe[r.IntN(len(palette.WebSafe))]
}

func encodePNG(img image.Image) ([]byte, error) {
	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
