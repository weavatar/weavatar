package service_test

import (
	"regexp"
	"testing"
	"time"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/utils/jwt"
	"github.com/libtnb/validator"
	"github.com/libtnb/validator/contrib/openapi"

	"github.com/weavatar/weavatar/internal/avatar/service"
	"github.com/weavatar/weavatar/internal/shared/rule"
)

// TestRoutesDocument guards router startup, which fails on a bad document.
func TestRoutesDocument(t *testing.T) {
	v, err := validator.New(rule.Options(nil, nil, nil, true)...)
	must.NoError(t, err)
	g, err := openapi.New("weavatar", "dev",
		openapi.WithValidator(v),
		openapi.WithSchema[time.Time](&openapi.Schema{Type: "string", Format: "date-time"}),
	)
	must.NoError(t, err)

	params := regexp.MustCompile(`:([A-Za-z0-9_]+)`)
	for _, e := range service.AvatarRoutes(service.NewAvatarService(nil, v), jwt.NewJWT(testKey, time.Hour)) {
		if e.Document != nil {
			check.NoError(t, e.Document(g, e.Method, params.ReplaceAllString(e.Path, "{$1}"), e.Summary, e.Tags))
		}
	}
	_, err = g.JSON()
	must.NoError(t, err)
}
