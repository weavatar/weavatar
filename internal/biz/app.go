package biz

import "github.com/dromara/carbon/v2"

type App struct {
	ID        string          `gorm:"type:char(10);primaryKey" json:"id"`
	UserID    string          `gorm:"type:char(10);index;not null" json:"user_id"`
	Name      string          `gorm:"type:varchar(255);not null" json:"name"`
	Secret    string          `gorm:"type:varchar(255);not null" json:"-"`
	CreatedAt carbon.DateTime `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"type:datetime;not null" json:"updated_at"`

	User *User `json:"-"`
}
