package biz

import "github.com/dromara/carbon/v2"

type App struct {
	ID        string          `gorm:"type:char(10);primaryKey" json:"id"`
	UserID    string          `gorm:"type:char(10)" json:"user_id"`
	Name      string          `json:"name"`
	Secret    string          `json:"-"`
	CreatedAt carbon.DateTime `gorm:"type:datetime" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"type:datetime" json:"updated_at"`

	User *User `json:"-"`
}
