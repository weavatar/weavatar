//go:build !cgo

package service

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/weavatar/weavatar/internal/biz"
)

type AvatarService struct {
	log        *slog.Logger
	avatarRepo biz.AvatarRepo
}

func NewAvatarService(log *slog.Logger, avatar biz.AvatarRepo) *AvatarService {
	return &AvatarService{
		log:        log,
		avatarRepo: avatar,
	}
}

func (r *AvatarService) Avatar(c fiber.Ctx) error {
	return nil
}

func (r *AvatarService) List(c fiber.Ctx) error {
	return nil
}

func (r *AvatarService) Create(c fiber.Ctx) error {
	return nil
}

func (r *AvatarService) Update(c fiber.Ctx) error {
	return nil
}

func (r *AvatarService) Delete(c fiber.Ctx) error {
	return nil
}

func (r *AvatarService) Check(c fiber.Ctx) error {
	return nil
}

func (r *AvatarService) Qq(c fiber.Ctx) error {
	return nil
}
