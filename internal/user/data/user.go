// Package data implements the user module's ports.
package data

import (
	"context"

	"github.com/go-rio/rio"
	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/shared/database"
	"github.com/weavatar/weavatar/internal/user/biz"
)

var userByUnionIDQuery = rio.From[biz.User]().Where("union_id = ?").Must()

type userRepo struct {
	db *rio.DB
}

func NewUserRepo(db *rio.DB) biz.UserRepo {
	return &userRepo{db: db}
}

func (r *userRepo) FindByUnionID(ctx context.Context, unionID string) (*biz.User, error) {
	user, err := userByUnionIDQuery.First(ctx, database.Q(ctx, r.db), unionID)
	if err != nil {
		return nil, oops.In("user").Wrapf(err, "find user by union id")
	}

	return user, nil
}

func (r *userRepo) Find(ctx context.Context, id string) (*biz.User, error) {
	user, err := rio.Find[biz.User](ctx, database.Q(ctx, r.db), id)
	if err != nil {
		return nil, oops.In("user").Wrapf(err, "find user %s", id)
	}

	return user, nil
}

func (r *userRepo) Create(ctx context.Context, user *biz.User) error {
	if err := rio.Insert(ctx, database.Q(ctx, r.db), user); err != nil {
		return oops.In("user").Wrapf(err, "create user")
	}

	return nil
}

func (r *userRepo) Update(ctx context.Context, user *biz.User) error {
	if err := rio.Update(ctx, database.Q(ctx, r.db), user); err != nil {
		return oops.In("user").Wrapf(err, "update user %s", user.ID)
	}

	return nil
}

// Delete relies on the softdelete column: rio only stamps deleted_at.
func (r *userRepo) Delete(ctx context.Context, user *biz.User) error {
	if err := rio.Delete(ctx, database.Q(ctx, r.db), user); err != nil {
		return oops.In("user").Wrapf(err, "delete user %s", user.ID)
	}

	return nil
}
