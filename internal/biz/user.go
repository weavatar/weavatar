package biz

import (
	"github.com/dromara/carbon/v2"
	"gorm.io/gorm"
)

type User struct {
	ID        string          `gorm:"type:char(10);primaryKey" json:"id"`
	OpenID    string          `gorm:"type:char(10);not null" json:"open_id"`
	UnionID   string          `gorm:"type:char(10);not null" json:"union_id"`
	Nickname  string          `gorm:"type:varchar(255);not null" json:"nickname"`
	Avatar    string          `gorm:"type:varchar(255);not null" json:"avatar"`
	RealName  bool            `gorm:"not null" json:"real_name"`
	CreatedAt carbon.DateTime `gorm:"type:datetime;not null" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"type:datetime;not null" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"type:datetime" json:"-"`

	App []*App `json:"-"`
}

type UserRepo interface {
	List(page, limit uint) ([]*User, int64, error)
	Get(id uint) (*User, error)
	Save(user *User) error
	Delete(id uint) error
}
