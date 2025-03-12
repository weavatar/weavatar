package request

import (
	"encoding/json"
	"mime/multipart"
	"regexp"
	"slices"
	"strings"

	"github.com/go-rat/utils/convert"
	"github.com/go-rat/utils/str"
	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/pkg/geetest"
)

var hashRegex = regexp.MustCompile(`^([a-f0-9]{64})|([a-f0-9]{32})$`)

type Avatar struct {
	AppID   string
	Hash    string `uri:"hash"`
	Ext     string
	Size    int
	Force   bool
	Default string
}

func (r *Avatar) Prepare(c fiber.Ctx) error {
	// 从哈希中提取出哈希值和扩展名
	hashExt := strings.Split(r.Hash, ".")
	r.Hash = strings.ToLower(hashExt[0])
	if len(hashExt) > 1 && slices.Contains([]string{"png", "jpg", "jpeg", "gif", "webp", "tiff", "heif", "heic", "avif", "jxl"}, hashExt[1]) {
		r.Ext = hashExt[1]
	} else {
		r.Ext = "webp"
	}

	// 过滤 appid 参数
	r.AppID = fiber.Query(c, "app", fiber.Query(c, "appid", ""))

	// 过滤 s 参数
	r.Size = fiber.Query(c, "s", fiber.Query(c, "size", 80))
	if r.Size > 2048 {
		r.Size = 2048
	}
	if r.Size < 1 {
		r.Size = 1
	}

	// 过滤 f 参数
	fd := fiber.Query(c, "f", fiber.Query(c, "forcedefault", "n"))
	if !slices.Contains([]string{"y", "yes", "n", "no"}, fd) {
		fd = "n"
	}
	r.Force = strings.Contains(fd, "y") || strings.Contains(fd, "yes")

	// 过滤 d 参数
	r.Default = fiber.Query(c, "d", fiber.Query(c, "default", ""))
	if !slices.Contains([]string{"404", "mp", "mm", "mystery", "identicon", "monsterid", "wavatar", "retro", "robohash", "blank", "color", "letter", "initials"}, r.Default) { // TODO deprecated letter in the future
		// 如果不是预设的默认头像，则检查是否是合法的 URL
		if !str.IsURL(r.Default) {
			r.Default = ""
		}
	}

	// 过滤 hash
	if !hashRegex.MatchString(r.Hash) {
		r.Force = true
	}

	return nil
}

type AvatarCreate struct {
	Raw        string                `form:"raw" validate:"required"`
	VerifyCode string                `form:"verify_code" validate:"required|verifyCode:Raw,avatar"`
	Avatar     *multipart.FileHeader `form:"avatar" validate:"required|image"`

	Captcha geetest.Ticket `form:"-" validate:"required|geetest"`
}

func (r *AvatarCreate) Prepare(c fiber.Ctx) error {
	if err := json.Unmarshal(convert.UnsafeBytes(c.FormValue("captcha")), &r.Captcha); err != nil {
		return err
	}
	return nil
}

type AvatarUpdate struct {
	Hash   string                `uri:"hash" validate:"required"`
	Avatar *multipart.FileHeader `form:"avatar" validate:"required|image"`

	Captcha geetest.Ticket `form:"-" validate:"required|geetest"`
}

func (r *AvatarUpdate) Prepare(c fiber.Ctx) error {
	if err := json.Unmarshal(convert.UnsafeBytes(c.FormValue("captcha")), &r.Captcha); err != nil {
		return err
	}
	return nil
}

type AvatarDelete struct {
	Hash string `uri:"hash" validate:"required"`
}

type AvatarCheck struct {
	Raw string `query:"raw"`
}

type AvatarQq struct {
	Qq string `query:"qq"`
}
