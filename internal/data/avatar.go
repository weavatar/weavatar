package data

import (
	"bytes"
	"errors"
	"fmt"
	"hash/fnv"
	"image/color"
	"image/png"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/go-rat/utils/convert"
	"github.com/go-rat/utils/file"
	"github.com/go-rat/utils/str"
	"github.com/goki/freetype/truetype"
	"github.com/imroc/req/v3"
	"github.com/ipsn/go-adorable"
	"github.com/issue9/identicon/v2"
	"github.com/o1egl/govatar"
	"github.com/tnb-labs/letteravatar"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/pkg/avatars"
	"github.com/weavatar/weavatar/pkg/embed"
)

type avatarRepo struct {
	db     *gorm.DB
	font   *truetype.Font
	client *req.Client
}

func NewAvatarRepo(db *gorm.DB) (biz.AvatarRepo, error) {
	fontData, err := embed.FontFS.ReadFile("font/SourceHanSans-VF-700.ttf")
	if err != nil {
		return nil, err
	}
	font, err := truetype.Parse(fontData)
	if err != nil {
		return nil, err
	}

	client := req.C()
	client.SetTimeout(5 * time.Second)
	client.SetCommonRetryCount(2)
	client.ImpersonateSafari()

	return &avatarRepo{
		db:     db,
		font:   font,
		client: client,
	}, nil
}

func (r *avatarRepo) GetByRaw(raw string) (*biz.Avatar, error) {
	avatar := new(biz.Avatar)
	if err := r.db.Where("raw = ?", raw).First(avatar).Error; err != nil {
		return nil, err
	}
	return avatar, nil
}

func (r *avatarRepo) GetWeAvatar(hash, appID string) (string, []byte, time.Time, error) {
	avatar := new(biz.Avatar)
	if err := r.db.Preload("User").Preload("AppSHA256", "app_id = ?", appID).Preload("AppMD5", "app_id = ?", appID).Where("sha256 = ?", hash).Or("md5 = ?", hash).First(avatar).Error; err != nil {
		return "", nil, time.Now(), err
	}

	var img []byte
	var err error
	// 优先加载 App 头像
	if avatar.AppSHA256 != nil || avatar.AppMD5 != nil {
		fp := filepath.Join("storage", "upload", "app", appID, avatar.SHA256[:2], avatar.SHA256)
		img, err = os.ReadFile(fp)
		if avatar.AppSHA256 != nil {
			return avatar.User.Nickname, img, avatar.AppSHA256.UpdatedAt.StdTime(), err
		}
		if avatar.AppMD5 != nil {
			return avatar.User.Nickname, img, avatar.AppMD5.UpdatedAt.StdTime(), err
		}
	}

	fp := filepath.Join("storage", "upload", "default", avatar.SHA256[:2], avatar.SHA256)
	img, err = os.ReadFile(fp)

	return avatar.User.Nickname, img, avatar.UpdatedAt.StdTime(), err
}

// GetQqByHash 通过哈希获取 Q 头像
// 系统有前 16 位的 MD5 和 SHA256 哈希表
// 哈希表通过前两位十六进制数分表存储
func (r *avatarRepo) GetQqByHash(hash string) (string, []byte, time.Time, error) {
	hashType := "sha256"
	if len(hash) == 32 {
		hashType = "md5"
	}
	index, err := strconv.ParseUint(hash[:2], 16, 64)
	if err != nil {
		return "", nil, time.Now(), err
	}

	table := fmt.Sprintf("hash.qq_%s_%d", hashType, index)
	qqHash := new(biz.QqHash)
	if err = r.db.Table(table).Where("hash = UNHEX(?)", hash[:16]).First(qqHash).Error; err != nil {
		return "", nil, time.Now(), err
	}

	cache := filepath.Join("storage", "cache", "qq", qqHash.Q[:2], qqHash.Q)
	if file.Exists(cache) {
		img, err := os.ReadFile(cache)
		lastModified, err2 := file.LastModified(cache, "UTC")
		if err == nil && err2 == nil && lastModified.Add(14*24*time.Hour).After(time.Now()) {
			return qqHash.Q, img, lastModified, nil
		}
	}

	img, err := avatars.Qq(qqHash.Q)
	if err != nil {
		return "", nil, time.Now(), err
	}

	if err = os.WriteFile(cache, img, 0644); err != nil {
		return "", nil, time.Now(), err
	}

	return qqHash.Q, img, time.Now(), nil
}

// GetGravatarByHash 通过哈希获取 Gravatar 头像
// Gravatar 支持 SHA256 和 MD5，可以直接缓存
// 但这样对于一个邮箱，可能会有两个头像，但是这个概率非常小，且不会造成问题，所以不做处理
func (r *avatarRepo) GetGravatarByHash(hash string) ([]byte, time.Time, error) {
	cache := filepath.Join("storage", "cache", "gravatar", hash[:2], hash)
	if file.Exists(cache) {
		img, err := os.ReadFile(cache)
		lastModified, err2 := file.LastModified(cache, "UTC")
		if err == nil && err2 == nil && lastModified.Add(14*24*time.Hour).After(time.Now()) {
			return img, lastModified, nil
		}
	}

	img, err := avatars.Gravatar(hash)
	if err != nil {
		return nil, time.Now(), err
	}

	if err = os.WriteFile(cache, img, 0644); err != nil {
		return nil, time.Now(), err
	}

	return img, time.Now(), nil
}

// GetByType 通过头像类型获取头像
func (r *avatarRepo) GetByType(avatarType string, options ...string) ([]byte, time.Time, error) {
	switch avatarType {
	case "mp", "mm", "mystery":
		img, err := embed.DefaultFS.ReadFile(filepath.Join("default", "mp.png"))
		return img, time.Now(), err
	case "identicon":
		red, green, blue, err := r.randomColor(options[0])
		if err != nil {
			return nil, time.Now(), err
		}
		red2, green2, blue2, err := r.randomColor(options[0])
		if err != nil {
			return nil, time.Now(), err
		}
		img := identicon.Make(identicon.Style1, 600,
			color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: 255},
			color.RGBA{R: uint8(red2), G: uint8(green2), B: uint8(blue2), A: 255},
			convert.UnsafeBytes(options[0]))
		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "monsterid":
		img, err := govatar.GenerateForUsername(govatar.FEMALE, options[0])
		if err != nil {
			return nil, time.Now(), err
		}
		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "wavatar":
		return adorable.PseudoRandom(convert.UnsafeBytes(options[0])), time.Now(), nil
	case "retro":
		red, green, blue, err := r.randomColor(options[0])
		if err != nil {
			return nil, time.Now(), err
		}
		red2, green2, blue2, err := r.randomColor(options[0])
		if err != nil {
			return nil, time.Now(), err
		}
		img := identicon.Make(identicon.Style2, 600,
			color.RGBA{R: uint8(red), G: uint8(green), B: uint8(blue), A: 255},
			color.RGBA{R: uint8(red2), G: uint8(green2), B: uint8(blue2), A: 255},
			convert.UnsafeBytes(options[0]))
		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "robohash":
		img, err := govatar.GenerateForUsername(govatar.MALE, options[0])
		if err != nil {
			return nil, time.Now(), err
		}
		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "blank":
		img, err := embed.DefaultFS.ReadFile(filepath.Join("default", "blank.png"))
		if err != nil {
			return nil, time.Now(), err
		}
		return img, time.Now(), nil
	case "color":
		img, err := vips.Black(1, 1)
		if err != nil {
			return nil, time.Now(), err
		}
		defer img.Close()
		red, green, blue, err := r.randomColor(options[0])
		if err != nil {
			return nil, time.Now(), err
		}
		if err = img.Linear([]float64{0, 0, 0}, []float64{float64(red), float64(green), float64(blue)}); err != nil {
			return nil, time.Now(), err
		}
		data, _, err := img.ExportNative()
		return data, time.Now(), err
	case "letter", "initials": // TODO deprecated letter in the future
		fontSize := 500
		letters := []rune(strings.ToUpper(options[1]))
		length := len(letters)
		if length > 1 {
			fontSize = 400
			letters = letters[:2]
		}
		img, err := letteravatar.Draw(1000, letters, &letteravatar.Options{
			Font:       r.font,
			FontSize:   fontSize,
			PaletteKey: options[0], // 对相同的字符串使用相同的颜色
		})
		if err != nil {
			return nil, time.Now(), err
		}
		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	default:
		img, err := embed.DefaultFS.ReadFile(filepath.Join("default", "default.png"))
		return img, time.Now(), err
	}
}

// IsBanned 通过哈希判断头像是否被封禁
func (r *avatarRepo) IsBanned(img []byte) (bool, error) {
	var count int64
	if err := r.db.Model(&biz.Image{}).Where("hash = ? AND banned = 1", str.SHA256(convert.UnsafeString(img))).Count(&count).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// TODO 审核头像
			return false, nil
		}
		return false, err
	}

	return count > 0, nil
}

func (r *avatarRepo) randomColor(hash string) (int, int, int, error) {
	h := fnv.New64a()
	if _, err := h.Write(convert.UnsafeBytes(hash)); err != nil {
		return 0, 0, 0, err
	}
	rd := rand.New(rand.NewPCG(h.Sum64(), (h.Sum64()>>1)|1))
	return rd.IntN(256), rd.IntN(256), rd.IntN(256), nil
}
