package service

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/utils/jwt"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/transport"
)

func AvatarRoutes(avatar *AvatarService, parser *jwt.JWT) transport.Endpoints {
	login := []fiber.Handler{transport.MustLogin(parser)}
	tags := []string{"avatar"}

	return transport.Endpoints{
		{Method: fiber.MethodGet, Path: "/api/avatar", Handler: avatar.Avatar},
		{Method: fiber.MethodHead, Path: "/api/avatar", Handler: avatar.Avatar},
		{Method: fiber.MethodGet, Path: "/api/avatar/:hash", Handler: avatar.Avatar},
		{Method: fiber.MethodHead, Path: "/api/avatar/:hash", Handler: avatar.Avatar},
		{Method: fiber.MethodGet, Path: "/api/avatars", Handler: avatar.List, Middlewares: login,
			Summary: "获取我的头像列表", Tags: tags,
			Document: transport.Describe[transport.Paginate, transport.Envelope[transport.Page[*biz.Avatar]]](http.StatusOK)},
		{Method: fiber.MethodPost, Path: "/api/avatars", Handler: avatar.Create, Middlewares: login,
			Summary: "上传头像", Tags: tags,
			Document: transport.Describe[AvatarCreate, transport.Envelope[biz.Avatar]](http.StatusOK)},
		{Method: fiber.MethodGet, Path: "/api/avatars/check", Handler: avatar.Check, Middlewares: login,
			Summary: "检查邮箱或手机号是否已绑定头像", Tags: tags,
			Document: transport.Describe[AvatarCheck, transport.Envelope[AvatarBound]](http.StatusOK)},
		{Method: fiber.MethodGet, Path: "/api/avatars/qq", Handler: avatar.Qq, Middlewares: login,
			Summary: "获取社交头像（Base64）", Tags: tags,
			Document: transport.Describe[AvatarQq, transport.Envelope[string]](http.StatusOK)},
		{Method: fiber.MethodPut, Path: "/api/avatars/:hash", Handler: avatar.Update, Middlewares: login,
			Summary: "更新头像", Tags: tags,
			Document: transport.Describe[AvatarUpdate, transport.Envelope[biz.Avatar]](http.StatusOK)},
		{Method: fiber.MethodDelete, Path: "/api/avatars/:hash", Handler: avatar.Delete, Middlewares: login,
			Summary: "删除头像", Tags: tags,
			Document: transport.DescribeNoBody[AvatarDelete](http.StatusOK)},
	}
}
