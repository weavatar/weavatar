package biz

import "github.com/dromara/carbon/v2"

type Avatar struct {
	SHA256    string          `gorm:"type:char(64);primaryKey" json:"sha256"`
	MD5       string          `gorm:"type:char(32)" json:"md5"`
	Raw       string          `json:"raw"`
	UserID    string          `gorm:"type:char(10)" json:"user_id"`
	CreatedAt carbon.DateTime `gorm:"type:datetime" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"type:datetime" json:"updated_at"`

	AppSHA256 *AppAvatar `gorm:"foreignKey:AvatarSHA256;references:SHA256" json:"-"`
	AppMD5    *AppAvatar `gorm:"foreignKey:AvatarMD5;references:MD5" json:"-"`
}
