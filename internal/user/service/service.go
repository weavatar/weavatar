// Package service adapts HTTP to the user usecase.
package service

import (
	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/shared/transport"
	"github.com/weavatar/weavatar/internal/user/biz"
)

type UserService struct {
	user     *biz.UserUsecase
	validate *validator.Validator
}

func NewUserService(user *biz.UserUsecase, validate *validator.Validator) *UserService {
	return &UserService{
		user:     user,
		validate: validate,
	}
}

func (r *UserService) Login(c fiber.Ctx) error {
	url, err := r.user.LoginURL(c.Context())
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, LoginURL{URL: url})
}

func (r *UserService) Callback(c fiber.Ctx) error {
	req, err := transport.Bind[UserCallback](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	token, err := r.user.Callback(c.Context(), req.Code, req.State)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, LoginToken{Token: token})
}

func (r *UserService) Info(c fiber.Ctx) error {
	user, err := r.user.Get(c.Context(), transport.UserID(c))
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, user)
}

func (r *UserService) UpdateInfo(c fiber.Ctx) error {
	req, err := transport.Bind[UserUpdate](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	user, err := r.user.Update(c.Context(), transport.UserID(c), req.Nickname, req.Avatar)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, user)
}

// Logout is a no-op: tokens are stateless and the client discards its own.
func (r *UserService) Logout(c fiber.Ctx) error {
	return transport.Success[any](c, nil)
}
