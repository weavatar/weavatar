package avatars

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/imroc/req/v3"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
)

func TestGravatar(t *testing.T) {
	base := useGravatar(t, func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/avatar/image":
			w.Header().Set("Content-Type", "image/png")
			_, _ = w.Write([]byte("png"))
		case "/avatar/page":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("<html>"))
		default:
			http.NotFound(w, r)
		}
	})

	img, err := Gravatar(t.Context(), base+"/", "image")
	must.NoError(t, err)
	check.Equal(t, string(img), "png")

	_, err = Gravatar(t.Context(), base, "page")
	check.ErrorContains(t, err, "content type")

	_, err = Gravatar(t.Context(), base, "missing")
	check.ErrorContains(t, err, "status 404")
}

func TestGravatarRejectsOversizedBody(t *testing.T) {
	body := bytes.Repeat([]byte{0}, maxResponseSize+1)
	base := useGravatar(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Content-Length", strconv.Itoa(len(body)))
		_, _ = w.Write(body)
	})

	_, err := Gravatar(t.Context(), base, "huge")

	check.ErrorIs(t, err, req.ErrResponseBodyTooLarge)
}

func useGravatar(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}
