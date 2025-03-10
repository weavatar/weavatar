package biz

import (
	"github.com/dromara/carbon/v2"
	"gorm.io/gorm"
)

type User struct {
	ID        string          `gorm:"type:char(10);primaryKey" json:"id"`
	OpenID    string          `gorm:"type:char(10)" json:"open_id"`
	UnionID   string          `gorm:"type:char(10)" json:"union_id"`
	Nickname  string          `json:"nickname"`
	Avatar    string          `json:"avatar"`
	RealName  bool            `json:"real_name"`
	CreatedAt carbon.DateTime `gorm:"type:datetime" json:"created_at"`
	UpdatedAt carbon.DateTime `gorm:"type:datetime" json:"updated_at"`
	DeletedAt gorm.DeletedAt  `gorm:"type:datetime" json:"-"`

	App []*App `json:"-"`
}

type UserRepo interface {
	List(page, limit uint) ([]*User, int64, error)
	Get(id uint) (*User, error)
	Save(user *User) error
	Delete(id uint) error
}
