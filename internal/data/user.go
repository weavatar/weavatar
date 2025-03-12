package data

import (
	"time"

	"github.com/go-rat/utils/jwt"
	"github.com/knadh/koanf/v2"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/pkg/id"
)

type userRepo struct {
	conf *koanf.Koanf
	db   *gorm.DB
	jwt  *jwt.JWT
}

func NewUserRepo(conf *koanf.Koanf, db *gorm.DB) biz.UserRepo {
	return &userRepo{
		conf: conf,
		db:   db,
		jwt:  jwt.NewJWT(conf.MustString("app.key"), time.Hour),
	}
}

func (r *userRepo) LoginByOauth(openID, unionID string, realName bool) (string, error) {
	user := new(biz.User)
	if err := r.db.Where("union_id = ?", unionID).Attrs(&biz.User{
		ID:       id.Generate(),
		OpenID:   openID,
		UnionID:  unionID,
		Nickname: "新用户",
		Avatar:   "https://weavatar.com/avatar/?d=mp",
	}).Assign(&biz.User{
		RealName: realName,
	}).FirstOrCreate(user).Error; err != nil {
		return "", err
	}

	token, err := r.jwt.Generate(&jwt.Claims{
		Subject:  user.ID,
		Audience: []string{r.conf.MustString("http.domain")},
		Issuer:   "https://" + r.conf.MustString("http.domain"),
	})
	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *userRepo) List(page, limit uint) ([]*biz.User, int64, error) {
	var total int64
	if err := r.db.Model(&biz.User{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*biz.User
	if err := r.db.Offset(int((page - 1) * limit)).Limit(int(limit)).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *userRepo) Get(id string) (*biz.User, error) {
	user := new(biz.User)
	if err := r.db.Where("id = ?", id).First(user).Error; err != nil {
		return nil, err
	}

	return user, nil
}

func (r *userRepo) Save(user *biz.User) error {
	return r.db.Save(user).Error
}

func (r *userRepo) Delete(id string) error {
	return r.db.Delete(&biz.User{}, id).Error
}
