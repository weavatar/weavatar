package service

import (
	"net/http"

	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/internal/shared/transport"
)

func SystemRoutes(system *SystemService) transport.Endpoints {
	return transport.Endpoints{
		{Method: fiber.MethodGet, Path: "/api/system/count", Handler: system.Count,
			Summary: "获取昨日请求量", Tags: []string{"system"},
			Document: transport.Describe[struct{}, transport.Envelope[Count]](http.StatusOK)},
		{Method: fiber.MethodGet, Path: "/api/system/random_avatars", Handler: system.RandomAvatars,
			Summary: "获取随机头像", Tags: []string{"system"},
			Document: transport.Describe[struct{}, transport.Envelope[RandomAvatars]](http.StatusOK)},
	}
}
