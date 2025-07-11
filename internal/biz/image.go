package biz

import "github.com/dromara/carbon/v2"

type Image struct {
	Hash      string          `gorm:"type:char(64);primaryKey" json:"hash"`
	Banned    bool            `gorm:"not null" json:"banned"`
	Remark    string          `gorm:"type:text;not null" json:"remark"`
	CreatedAt carbon.DateTime `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"type:datetime;not null" json:"updated_at"`
}
