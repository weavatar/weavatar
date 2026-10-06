package data

import (
	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/registry"
)

// NewUserCleanup removes a user's avatars when the account is deleted.
func NewUserCleanup(uc *biz.AvatarUsecase) registry.UserCleanup {
	return registry.UserCleanup{Name: "avatar", Run: uc.DeleteByUser}
}
