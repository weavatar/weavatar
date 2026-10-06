package service

import (
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/validator/contrib/openapi"

	"github.com/weavatar/weavatar/internal/shared/transport"
)

func SystemRoutes(system *SystemService) transport.Endpoints {
	tags := []string{"system"}

	return transport.Endpoints{
		{Method: fiber.MethodGet, Path: "/api/system/count", Handler: system.Count,
			Summary: "获取昨日请求量", Tags: tags,
			Document: transport.Describe[openapi.NoBody, transport.Envelope[Count]](http.StatusOK)},
		{Method: fiber.MethodGet, Path: "/api/system/random_avatars", Handler: system.RandomAvatars,
			Summary: "获取随机头像", Tags: tags,
			Document: transport.Describe[openapi.NoBody, transport.Envelope[RandomAvatars]](http.StatusOK)},
	}
}
