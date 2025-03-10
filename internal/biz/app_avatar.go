package biz

import "github.com/dromara/carbon/v2"

type AppAvatar struct {
	AppID        string          `gorm:"type:char(10);primaryKey" json:"app_id"`
	AvatarSHA256 string          `gorm:"type:char(64);primaryKey" json:"avatar_sha256"`
	AvatarMD5    string          `gorm:"type:char(32);primaryKey" json:"avatar_md5"`
	CreatedAt    carbon.DateTime `gorm:"type:datetime" json:"created_at"`
	UpdatedAt    carbon.DateTime `gorm:"type:datetime" json:"updated_at"`
}
