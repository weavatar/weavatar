package service

import "github.com/gofiber/fiber/v3"

type SystemService struct {
}

func NewSystemService() *SystemService {
	return &SystemService{}
}

func (r *SystemService) Count(c fiber.Ctx) error {
	return Success(c, fiber.Map{
		"usage": 0,
	})
}
