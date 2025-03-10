package biz

import "github.com/dromara/carbon/v2"

type Image struct {
	Hash      string          `gorm:"type:char(64);primaryKey" json:"hash"`
	Ban       bool            `json:"ban"`
	Remark    string          `gorm:"type:text" json:"remark"`
	CreatedAt carbon.DateTime `gorm:"type:datetime" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"type:datetime" json:"updated_at"`
}
