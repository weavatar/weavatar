package service

import (
	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/internal/http/request"
)

type UserService struct {
	user biz.UserRepo
}

func NewUserService(user biz.UserRepo) *UserService {
	return &UserService{
		user: user,
	}
}

func (r *UserService) List(c fiber.Ctx) error {
	req, err := Bind[request.Paginate](c)
	if err != nil {
		return Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}
	users, total, err := r.user.List(req.Page, req.Limit)
	if err != nil {
		return ErrorSystem(c)
	}

	return Success(c, map[string]any{
		"total": total,
		"items": users,
	})
}
