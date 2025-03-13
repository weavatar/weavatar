package service

import (
	"fmt"
	"net/url"
	"time"

	"github.com/go-rat/cache"
	"github.com/go-rat/utils/convert"
	"github.com/go-rat/utils/str"
	"github.com/gofiber/fiber/v3"
	"github.com/knadh/koanf/v2"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/internal/http/request"
	"github.com/weavatar/weavatar/pkg/oauth"
)

type UserService struct {
	cache cache.Cache
	conf  *koanf.Koanf
	oauth *oauth.Oauth
	user  biz.UserRepo
}

func NewUserService(cache cache.Cache, conf *koanf.Koanf, user biz.UserRepo) *UserService {
	return &UserService{
		cache: cache,
		conf:  conf,
		oauth: oauth.NewOauth(conf.MustString("oauth.clientID"), conf.MustString("oauth.clientSecret"), conf.MustString("oauth.baseUrl")),
		user:  user,
	}
}

func (r *UserService) Login(c fiber.Ctx) error {
	state := fmt.Sprintf("login-%s", str.Random(16))
	if err := r.cache.Put(state, convert.CopyString(c.IP()), 5*time.Minute); err != nil {
		return ErrorSystem(c)
	}

	return Success(c, fiber.Map{
		"url": r.conf.MustString("oauth.baseUrl") + "/oauth/authorize?client_id=" + r.conf.MustString("oauth.clientID") + "&redirect_uri=" + url.QueryEscape("https://"+r.conf.MustString("http.domain")+"/oauth/callback") + "&response_type=code&scope=basic&state=" + state,
	})
}

func (r *UserService) Callback(c fiber.Ctx) error {
	req, err := Bind[request.UserCallback](c)
	if err != nil {
		return Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	if !r.cache.Has(req.State) {
		return Error(c, fiber.StatusBadRequest, "状态已过期")
	}

	accessToken, err := r.oauth.GetToken(req.Code, fmt.Sprintf("https://%s/oauth/callback", r.conf.MustString("http.domain")))
	if err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	user, err := r.oauth.GetUserInfo(accessToken.AccessToken)
	if err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	token, err := r.user.LoginByOauth(user.Data.OpenID, user.Data.UnionID, user.Data.RealName)
	if err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	return Success(c, fiber.Map{
		"token": token,
	})
}

func (r *UserService) Get(c fiber.Ctx) error {
	user, err := r.user.Get(fiber.Locals[string](c, "user_id"))
	if err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	return Success(c, user)
}

func (r *UserService) Update(c fiber.Ctx) error {
	req, err := Bind[request.UserUpdate](c)
	if err != nil {
		return Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	user, err := r.user.Get(fiber.Locals[string](c, "user_id"))
	if err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	user.Nickname = req.Nickname
	user.Avatar = req.Avatar

	if err = r.user.Save(user); err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	return Success(c, nil)
}

func (r *UserService) Logout(c fiber.Ctx) error {
	return Success(c, nil)
}
