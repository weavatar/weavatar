package oauth

import (
	"fmt"

	"github.com/imroc/req/v3"
)

type BasicInfo struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data struct {
		Nickname  string `json:"nickname"`
		OpenID    string `json:"open_id"`
		UnionID   string `json:"union_id"`
		PhoneBind bool   `json:"phone_bind"`
		RealName  bool   `json:"real_name"`
	} `json:"data"`
}

type Token struct {
	AccessToken  string   `json:"access_token"`
	ExpiresIn    int      `json:"expires_in"`
	RefreshToken string   `json:"refresh_token"`
	Scope        []string `json:"scope"`
	TokenType    string   `json:"token_type"`
}

type Oauth struct {
	clientID     string
	clientSecret string
	baseURL      string
	client       *req.Client
}

func NewOauth(clientID, clientSecret, baseURL string) *Oauth {
	return &Oauth{
		clientID:     clientID,
		clientSecret: clientSecret,
		baseURL:      baseURL,
		client:       req.C().SetBaseURL(baseURL),
	}
}

// GetToken 获取 AccessToken 和 RefreshToken 信息
func (r *Oauth) GetToken(code string) (Token, error) {
	var token Token
	resp, err := r.client.R().SetQueryParams(map[string]string{
		"grant_type":    "authorization_code",
		"client_id":     r.clientID,
		"client_secret": r.clientSecret,
		"code":          code,
		"redirect_uri":  "https://weavatar/oauth/callback",
	}).SetSuccessResult(&token).Get("/api/v1/oauth/token")
	if err != nil {
		return token, err
	}
	if !resp.IsSuccessState() {
		return token, fmt.Errorf("failed to get token: %s", resp.String())
	}
	if token.AccessToken == "" || token.RefreshToken == "" {
		return token, fmt.Errorf("failed to unmarshal token: %s", resp.String())
	}

	return token, nil
}

// GetUserInfo 获取用户信息
func (r *Oauth) GetUserInfo(accessToken string) (BasicInfo, error) {
	var basicInfo BasicInfo
	resp, err := r.client.R().SetQueryParams(map[string]string{
		"access_token": accessToken,
	}).SetSuccessResult(&basicInfo).Get("/api/v1/oauth/user_info")
	if err != nil {
		return basicInfo, err
	}
	if !resp.IsSuccessState() {
		return basicInfo, fmt.Errorf("failed to get user info: %s", resp.String())
	}
	if basicInfo.Code != 0 || len(basicInfo.Data.OpenID) == 0 || len(basicInfo.Data.UnionID) == 0 {
		return basicInfo, fmt.Errorf("failed to unmarshal user info: %s", resp.String())
	}

	return basicInfo, nil
}
