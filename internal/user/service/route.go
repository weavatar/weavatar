package service

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/utils/jwt"
	"github.com/libtnb/validator/contrib/openapi"

	"github.com/weavatar/weavatar/internal/shared/transport"
	"github.com/weavatar/weavatar/internal/user/biz"
)

func UserRoutes(user *UserService, parser *jwt.JWT) transport.Endpoints {
	login := []fiber.Handler{transport.MustLogin(parser)}
	tags := []string{"user"}

	return transport.Endpoints{
		{Method: fiber.MethodGet, Path: "/api/user/login", Handler: user.Login,
			Summary: "获取登录地址", Tags: tags,
			Document: transport.Describe[openapi.NoBody, transport.Envelope[LoginURL]](http.StatusOK)},
		{Method: fiber.MethodPost, Path: "/api/user/callback", Handler: user.Callback,
			Summary: "登录回调", Tags: tags,
			Document: transport.Describe[UserCallback, transport.Envelope[LoginToken]](http.StatusOK)},
		{Method: fiber.MethodGet, Path: "/api/user/info", Handler: user.Info, Middlewares: login,
			Summary: "获取当前用户", Tags: tags,
			Document: transport.Describe[openapi.NoBody, transport.Envelope[biz.User]](http.StatusOK)},
		{Method: fiber.MethodPut, Path: "/api/user/info", Handler: user.UpdateInfo, Middlewares: login,
			Summary: "更新当前用户", Tags: tags,
			Document: transport.Describe[UserUpdate, transport.Envelope[biz.User]](http.StatusOK)},
		{Method: fiber.MethodPost, Path: "/api/user/logout", Handler: user.Logout, Middlewares: login,
			Summary: "退出登录", Tags: tags,
			Document: transport.Describe[openapi.NoBody, transport.Envelope[any]](http.StatusOK)},
		{Method: fiber.MethodGet, Path: "/api/user/deletion/login", Handler: user.DeletionLogin, Middlewares: login,
			Summary: "获取注销账号的确认地址", Tags: tags,
			Document: transport.Describe[openapi.NoBody, transport.Envelope[LoginURL]](http.StatusOK)},
		{Method: fiber.MethodPost, Path: "/api/user/deletion/confirm", Handler: user.DeletionConfirm, Middlewares: login,
			Summary: "确认注销账号", Tags: tags,
			Document: transport.Describe[UserCallback, transport.Envelope[any]](http.StatusOK)},
	}
}
