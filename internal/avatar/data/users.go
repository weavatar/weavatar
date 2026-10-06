package data

import (
	"context"
	"errors"

	"github.com/go-rio/rio"

	avatarbiz "github.com/weavatar/weavatar/internal/avatar/biz"
	userbiz "github.com/weavatar/weavatar/internal/user/biz"
)

// users adapts the user module's usecase to the biz.Users port.
type users struct {
	uc *userbiz.UserUsecase
}

func NewUsers(uc *userbiz.UserUsecase) avatarbiz.Users {
	return &users{uc: uc}
}

func (u *users) Nickname(ctx context.Context, userID string) (string, error) {
	user, err := u.uc.Get(ctx, userID)
	if err != nil {
		if errors.Is(err, rio.ErrNotFound) {
			return "", nil
		}
		return "", err
	}
	return user.Nickname, nil
}
