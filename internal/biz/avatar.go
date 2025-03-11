package biz

import (
	"time"

	"github.com/dromara/carbon/v2"
)

type Avatar struct {
	SHA256    string          `gorm:"type:char(64);primaryKey" json:"sha256"`
	MD5       string          `gorm:"type:char(32);not null" json:"md5"`
	Raw       string          `gorm:"type:varchar(255);not null" json:"raw"`
	UserID    string          `gorm:"type:char(10);not null" json:"user_id"`
	CreatedAt carbon.DateTime `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"type:datetime;not null" json:"updated_at"`

	AppSHA256 *AppAvatar `gorm:"foreignKey:AvatarSHA256;references:SHA256" json:"-"`
	AppMD5    *AppAvatar `gorm:"foreignKey:AvatarMD5;references:MD5" json:"-"`
}

type AvatarRepo interface {
	GetWeAvatar(hash, appID string) ([]byte, time.Time, error)
	GetQqByHash(hash string) (string, []byte, time.Time, error)
	GetGravatarByHash(hash string) ([]byte, time.Time, error)
	GetByType(avatarType string, option ...string) ([]byte, time.Time, error)
	IsBanned(img []byte) (bool, error)
}
