//go:build cgo

package data

import (
	"bytes"
	"fmt"
	"hash/fnv"
	"image"
	"image/color"
	"image/color/palette"
	"image/draw"
	"image/png"
	"io"
	"log/slog"
	"math/rand/v2"
	"mime/multipart"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/forPelevin/gomoji"
	"github.com/imroc/req/v3"
	"github.com/knadh/koanf/v2"
	"github.com/libtnb/cache"
	"github.com/libtnb/utils/convert"
	"github.com/libtnb/utils/file"
	"github.com/libtnb/utils/str"
	"github.com/spf13/cast"
	"github.com/weavatar/identicon"
	"github.com/weavatar/initials"
	"github.com/weavatar/monsterid"
	"github.com/weavatar/retricon"
	"github.com/weavatar/robohash"
	"github.com/weavatar/wavatar"
	"golang.org/x/image/font/opentype"
	"gorm.io/gorm"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/internal/http/request"
	"github.com/weavatar/weavatar/internal/queuejob"
	"github.com/weavatar/weavatar/pkg/avatars"
	"github.com/weavatar/weavatar/pkg/cdn"
	"github.com/weavatar/weavatar/pkg/embed"
	"github.com/weavatar/weavatar/pkg/queue"
)

const (
	CacheThreshold = 14 * 24 * time.Hour
)

type avatarRepo struct {
	cache  cache.Cache
	conf   *koanf.Koanf
	db     *gorm.DB
	log    *slog.Logger
	queue  *queue.Queue
	font   *opentype.Font
	emoji  *opentype.Font
	client *req.Client
	cdn    *cdn.Cdn
}

func NewAvatarRepo(cache cache.Cache, conf *koanf.Koanf, db *gorm.DB, log *slog.Logger, queue *queue.Queue) (biz.AvatarRepo, error) {
	font1, err := embed.FontFS.ReadFile("font/SourceHanSans-VF-700.ttf")
	if err != nil {
		return nil, err
	}
	font, err := opentype.Parse(font1)
	if err != nil {
		return nil, err
	}
	font2, err := embed.FontFS.ReadFile("font/NotoEmoji-Bold.ttf")
	if err != nil {
		return nil, err
	}
	emoji, err := opentype.Parse(font2)
	if err != nil {
		return nil, err
	}

	client := req.C()
	client.SetTimeout(5 * time.Second)
	client.SetCommonRetryCount(2)
	client.ImpersonateSafari()

	return &avatarRepo{
		cache:  cache,
		conf:   conf,
		db:     db,
		log:    log,
		queue:  queue,
		font:   font,
		emoji:  emoji,
		client: client,
		cdn:    cdn.NewCdn(conf),
	}, nil
}

func (r *avatarRepo) List(userID string, page, limit uint) ([]*biz.Avatar, int64, error) {
	var total int64
	if err := r.db.Model(&biz.Avatar{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var list []*biz.Avatar
	if err := r.db.Where("user_id = ?", userID).Order("created_at desc").Offset(int((page - 1) * limit)).Limit(int(limit)).Find(&list).Error; err != nil {
		return nil, 0, err
	}

	return list, total, nil
}

func (r *avatarRepo) Get(userID string, hash string) (*biz.Avatar, error) {
	avatar := new(biz.Avatar)
	if err := r.db.Where("sha256 = ? OR md5 = ?", hash, hash).Where("user_id = ?", userID).First(avatar).Error; err != nil {
		return nil, err
	}
	return avatar, nil
}

func (r *avatarRepo) Create(userID string, req *request.AvatarCreate) (*biz.Avatar, error) {
	f, err := req.Avatar.Open()
	if err != nil {
		return nil, err
	}
	defer func(f multipart.File) {
		_ = f.Close()
	}(f)

	b, _ := io.ReadAll(f)
	img, err := r.formatAvatar(b, 2048)
	if err != nil {
		return nil, err
	}

	avatar := new(biz.Avatar)
	avatar.UserID = userID

	err = r.db.Transaction(func(tx *gorm.DB) error {
		avatar.SHA256 = str.SHA256(req.Raw)
		avatar.MD5 = str.MD5(req.Raw)
		avatar.Raw = req.Raw
		if err = tx.Create(avatar).Error; err != nil {
			return err
		}

		fp := filepath.Join("storage", "upload", "default", avatar.SHA256[:2], avatar.SHA256)
		if err = file.Write(fp, img, 0644); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err = r.cdn.RefreshUrl([]string{
		fmt.Sprintf("https://%s/avatar/%s", r.conf.MustString("http.domain"), avatar.SHA256),
		fmt.Sprintf("https://%s/avatar/%s", r.conf.MustString("http.domain"), avatar.MD5),
	}); err != nil {
		r.log.Error("[AvatarRepo] failed to refresh url", slog.String("hash", avatar.SHA256), slog.String("err", err.Error()))
	}

	return avatar, nil
}

func (r *avatarRepo) Update(userID string, req *request.AvatarUpdate) (*biz.Avatar, error) {
	avatar, err := r.Get(userID, req.Hash)
	if err != nil {
		return nil, err
	}

	f, err := req.Avatar.Open()
	if err != nil {
		return nil, err
	}
	defer func(f multipart.File) {
		_ = f.Close()
	}(f)

	b, _ := io.ReadAll(f)
	img, err := r.formatAvatar(b, 2048)
	if err != nil {
		return nil, err
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Save(avatar).Error; err != nil {
			return err
		}

		fp := filepath.Join("storage", "upload", "default", avatar.SHA256[:2], avatar.SHA256)
		if err = file.Write(fp, img, 0644); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if err = r.cdn.RefreshUrl([]string{
		fmt.Sprintf("https://%s/avatar/%s", r.conf.MustString("http.domain"), avatar.SHA256),
		fmt.Sprintf("https://%s/avatar/%s", r.conf.MustString("http.domain"), avatar.MD5),
	}); err != nil {
		r.log.Error("[AvatarRepo] failed to refresh url", slog.String("hash", avatar.SHA256), slog.String("err", err.Error()))
	}

	return avatar, nil
}

func (r *avatarRepo) Delete(userID string, hash string) error {
	avatar, err := r.Get(userID, hash)
	if err != nil {
		return err
	}

	err = r.db.Transaction(func(tx *gorm.DB) error {
		if err = tx.Delete(avatar).Error; err != nil {
			return err
		}

		fp := filepath.Join("storage", "upload", "default", avatar.SHA256[:2], avatar.SHA256)
		if err = os.Remove(fp); err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return err
	}

	if err = r.cdn.RefreshUrl([]string{
		fmt.Sprintf("https://%s/avatar/%s", r.conf.MustString("http.domain"), avatar.SHA256),
		fmt.Sprintf("https://%s/avatar/%s", r.conf.MustString("http.domain"), avatar.MD5),
	}); err != nil {
		r.log.Error("[AvatarRepo] failed to refresh url", slog.String("hash", avatar.SHA256), slog.String("err", err.Error()))
	}

	return nil
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
	if err = r.db.Table(table).Where("h = UNHEX(?)", hash[:16]).First(qqHash).Error; err != nil {
		return "", nil, time.Now(), err
	}

	fn := filepath.Join("storage", "cache", "qq", qqHash.Q[:2], qqHash.Q)
	if file.Exists(fn) {
		img, err := os.ReadFile(fn)
		lmt, err2 := file.LastModified(fn, "UTC")
		if err == nil && err2 == nil && lmt.Add(CacheThreshold).After(time.Now()) {
			return qqHash.Q, img, lmt, nil
		}
	}

	img, err := avatars.Qq(qqHash.Q)
	if err != nil {
		return "", nil, time.Now(), err
	}
	if err = file.Write(fn, img, 0644); err != nil {
		return "", nil, time.Now(), err
	}

	return qqHash.Q, img, time.Now(), nil
}

// GetGravatarByHash 通过哈希获取 Gravatar 头像
// Gravatar 支持 SHA256 和 MD5，可以直接缓存
// 但这样对于一个邮箱，可能会有两个头像，但是这个概率非常小，且不会造成问题，所以不做处理
func (r *avatarRepo) GetGravatarByHash(hash string) ([]byte, time.Time, error) {
	fn := filepath.Join("storage", "cache", "gravatar", hash[:2], hash)
	if file.Exists(fn) {
		img, err := os.ReadFile(fn)
		lmt, err2 := file.LastModified(fn, "UTC")
		if err == nil && err2 == nil && lmt.Add(CacheThreshold).After(time.Now()) {
			return img, lmt, nil
		}
	}

	img, err := avatars.Gravatar(hash)
	if err != nil {
		return nil, time.Now(), err
	}
	if err = file.Write(fn, img, 0644); err != nil {
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
		size := cast.ToInt(options[1])
		if size < 16 {
			size = 16 // identicon need at least 16px
		}
		id, err := identicon.New(
			size,
			color.White,
			identicon.DarkColors...,
		)
		if err != nil {
			return nil, time.Now(), err
		}
		img := id.Make(convert.UnsafeBytes(options[0]))
		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "monsterid":
		img := monsterid.New(convert.UnsafeBytes(options[0]))
		buf := new(bytes.Buffer)
		err := png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "wavatar":
		img := wavatar.New(convert.UnsafeBytes(options[0]))
		buf := new(bytes.Buffer)
		err := png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "retro":
		img := retricon.MustNew(options[0], retricon.Gravatar)
		buf := new(bytes.Buffer)
		err := png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "robohash":
		rh, err := robohash.New(convert.UnsafeBytes(options[0]), "set1", "")
		if err != nil {
			return nil, time.Now(), err
		}
		img, err := rh.Assemble()
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
		bg, err := r.randomColor(options[0])
		if err != nil {
			return nil, time.Now(), err
		}
		size := cast.ToInt(options[1])
		img := image.NewRGBA(image.Rect(0, 0, size, size))
		draw.Draw(img, img.Bounds(), image.NewUniform(bg), image.Point{}, draw.Src)
		buf := new(bytes.Buffer)
		err = png.Encode(buf, img)
		return buf.Bytes(), time.Now(), err
	case "letter", "initials": // TODO deprecated letter in the future
		font := r.font
		fontSize := 500
		words := []rune(strings.ToUpper(options[1]))
		length := len(words)
		if length > 1 {
			fontSize = 400
			words = words[:2]
		}
		// 存在 emoji 时，只取找到的第一个 emoji
		if gomoji.FindAll(string(words)) != nil {
			words = words[:1]
			font = r.emoji
			fontSize = 500
		}
		img, err := initials.Draw(1000, words, &initials.Options{
			Font:       font,
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
func (r *avatarRepo) IsBanned(hash, appID string, img []byte) (bool, error) {
	var count int64
	if err := r.db.Model(&biz.Image{}).Where("hash = ? AND banned = 1", str.SHA256(convert.UnsafeString(img))).Count(&count).Error; err != nil {
		return false, err
	}

	if count == 0 {
		return false, r.queue.Push(queuejob.NewProcessAvatarAudit(r.cache, r.conf, r.db, r.log, r), []any{
			convert.CopyString(hash),
			convert.CopyString(appID),
		})
	}

	return count > 0, nil
}

func (r *avatarRepo) randomColor(hash string) (color.Color, error) {
	h := fnv.New64a()
	if _, err := h.Write(convert.UnsafeBytes(hash)); err != nil {
		return nil, err
	}
	rd := rand.New(rand.NewPCG(h.Sum64(), (h.Sum64()>>1)|1))
	return palette.WebSafe[rd.IntN(len(palette.WebSafe))], nil
}

/*func (r *avatarRepo) contrastColor(bgR, bgG, bgB int) (int, int, int) {
	fgR := 255 - bgR
	fgG := 255 - bgG
	fgB := 255 - bgB
	fgLuminance := 0.2126*float64(fgR) + 0.7152*float64(fgG) + 0.0722*float64(fgB)
	bgLuminance := 0.2126*float64(bgR) + 0.7152*float64(bgG) + 0.0722*float64(bgB)

	if math.Abs(fgLuminance-bgLuminance) < 128 {
		if bgLuminance > 128 {
			return 0, 0, 0 // 背景亮，前景暗
		} else {
			return 255, 255, 255 // 背景暗，前景亮
		}
	}

	return fgR, fgG, fgB
}*/

func (r *avatarRepo) formatAvatar(avatar []byte, size int) ([]byte, error) {
	img, err := vips.NewImageFromBuffer(avatar)
	if err != nil {
		return nil, err
	}
	defer img.Close()

	if img.Width() != img.Height() {
		return nil, fmt.Errorf("头像必须是正方形图片")
	}
	if img.Width() < 40 {
		return nil, fmt.Errorf("头像必须大于 40px")
	}
	if img.Width() > size {
		if err = img.ResizeWithVScale(float64(size)/float64(img.Width()), float64(size)/float64(img.Height()), vips.KernelLinear); err != nil {
			return nil, err
		}
	}

	data, _, err := img.ExportNative()
	return data, err
}
