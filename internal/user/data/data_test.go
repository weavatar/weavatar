package data

import (
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"
	"time"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/utils/jwt"

	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/internal/user/biz"
	"github.com/weavatar/weavatar/pkg/oauth"
)

func TestIssuedTokenParsesBack(t *testing.T) {
	parser := jwt.NewJWT("a-long-string-with-32-characters", time.Hour)

	token, err := NewTokens(parser, appinfo.Domain("weavatar.com")).Issue("u1")
	must.NoError(t, err)

	claims, err := parser.Parse(token)
	must.NoError(t, err)
	check.Equal(t, claims.Subject, "u1")
	check.Equal(t, claims.Issuer, "https://weavatar.com")
	check.True(t, slices.Contains(claims.Audience, "weavatar.com"))
}

func TestExchangeReturnsTheIdentity(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/api/v1/oauth/token":
			check.Equal(t, r.URL.Query().Get("code"), "code-1")
			check.Equal(t, r.URL.Query().Get("redirect_uri"), "https://weavatar.com/oauth/callback")
			_, _ = w.Write([]byte(`{"access_token":"at","refresh_token":"rt","expires_in":3600}`))
		case "/api/v1/oauth/user_info":
			check.Equal(t, r.URL.Query().Get("access_token"), "at")
			_, _ = w.Write([]byte(`{"code":0,"data":{"nickname":"alice","open_id":"open-1","union_id":"union-1","real_name":true}}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	provider := NewOAuthProvider(oauth.NewOauth("id", "secret", srv.URL))
	identity, err := provider.Exchange(t.Context(), "code-1", "https://weavatar.com/oauth/callback")

	must.NoError(t, err)
	check.Equal(t, identity, biz.Identity{OpenID: "open-1", UnionID: "union-1", Nickname: "alice", RealName: true})
}

func TestExchangeFailsOnRejectedCode(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"invalid_grant"}`, http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	_, err := NewOAuthProvider(oauth.NewOauth("id", "secret", srv.URL)).
		Exchange(t.Context(), "bad", "https://weavatar.com/oauth/callback")

	must.Error(t, err)
}
