package data

import (
	"context"

	"github.com/go-rio/rio"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/database"
)

type imageRepo struct {
	db *rio.DB
}

func NewImageRepo(db *rio.DB) biz.ImageRepo {
	return &imageRepo{db: db}
}

func (r *imageRepo) Find(ctx context.Context, hash string) (*biz.Image, error) {
	image, err := rio.Find[biz.Image](ctx, database.Q(ctx, r.db), hash)
	if err != nil {
		return nil, wrap(err, "find image %s", hash)
	}
	return image, nil
}

func (r *imageRepo) Create(ctx context.Context, image *biz.Image) error {
	if err := rio.Insert(ctx, database.Q(ctx, r.db), image); err != nil {
		return wrap(err, "create image %s", image.Hash)
	}
	return nil
}
