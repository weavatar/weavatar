package data_test

import (
	"bytes"
	"image/png"
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"

	"github.com/weavatar/weavatar/internal/avatar/data"
)

func TestGenerator_DrawsEveryKind(t *testing.T) {
	g, err := data.NewGenerator()
	must.NoError(t, err)

	for _, kind := range []string{"", "mp", "mm", "mystery", "blank", "identicon", "monsterid", "wavatar", "retro", "robohash", "color", "initials", "letter"} {
		t.Run(kind, func(t *testing.T) {
			img, err := g.Generate(kind, hash, 64, "W")
			must.NoError(t, err)
			_, err = png.DecodeConfig(bytes.NewReader(img))
			check.NoError(t, err)
		})
	}
}

func TestGenerator_IsStablePerSeed(t *testing.T) {
	g, err := data.NewGenerator()
	must.NoError(t, err)

	a, err := g.Generate("color", hash, 8, "")
	must.NoError(t, err)
	b, err := g.Generate("color", hash, 8, "")
	must.NoError(t, err)

	check.True(t, bytes.Equal(a, b))
}

func TestGenerator_InitialsTakeEmoji(t *testing.T) {
	g, err := data.NewGenerator()
	must.NoError(t, err)

	_, err = g.Generate("initials", hash, 80, "🐶x")

	check.NoError(t, err)
}
