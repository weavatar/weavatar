package biz

import "github.com/dromara/carbon/v2"

type AppAvatar struct {
	AvatarSHA256 string                             `gorm:"type:char(64);primaryKey" json:"avatar_sha256"`
	AvatarMD5    string                             `gorm:"type:char(32);primaryKey" json:"avatar_md5"`
	AppID        string                             `gorm:"type:char(10);primaryKey" json:"app_id"`
	CreatedAt    carbon.LayoutType[carbon.DateTime] `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt    carbon.LayoutType[carbon.DateTime] `gorm:"type:datetime;not null" json:"updated_at"`
}
