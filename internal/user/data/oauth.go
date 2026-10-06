package data

import (
	"context"

	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/user/biz"
	"github.com/weavatar/weavatar/pkg/oauth"
)

// oauthProvider adapts the OAuth client to the biz.OAuthProvider port.
type oauthProvider struct {
	client *oauth.Oauth
}

func NewOAuthProvider(client *oauth.Oauth) biz.OAuthProvider {
	return &oauthProvider{client: client}
}

func (p *oauthProvider) Exchange(ctx context.Context, code, redirectURI string) (biz.Identity, error) {
	token, err := p.client.GetToken(ctx, code, redirectURI)
	if err != nil {
		return biz.Identity{}, oops.In("user").Wrapf(err, "exchange oauth code")
	}

	info, err := p.client.GetUserInfo(ctx, token.AccessToken)
	if err != nil {
		return biz.Identity{}, oops.In("user").Wrapf(err, "get oauth user info")
	}

	return biz.Identity{
		OpenID:   info.Data.OpenID,
		UnionID:  info.Data.UnionID,
		Nickname: info.Data.Nickname,
		RealName: info.Data.RealName,
	}, nil
}
