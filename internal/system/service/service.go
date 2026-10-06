// Package service adapts HTTP to the system usecase.
package service

import (
	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/internal/shared/transport"
	"github.com/weavatar/weavatar/internal/system/biz"
)

type Count struct {
	Usage int64 `json:"usage"`
}

type RandomAvatars struct {
	Avatars []string `json:"avatars"`
}

type SystemService struct {
	system *biz.SystemUsecase
}

func NewSystemService(system *biz.SystemUsecase) *SystemService {
	return &SystemService{system: system}
}

func (r *SystemService) Count(c fiber.Ctx) error {
	return transport.Success(c, Count{Usage: r.system.Count(c.Context())})
}

func (r *SystemService) RandomAvatars(c fiber.Ctx) error {
	return transport.Success(c, RandomAvatars{Avatars: r.system.RandomAvatars(c.Context())})
}
