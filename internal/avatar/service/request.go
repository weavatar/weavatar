package service

import (
	"encoding/json"
	"errors"
	"mime/multipart"
	"slices"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/utils/str"

	"github.com/weavatar/weavatar/pkg/geetest"
)

const defaultAvatarSize = 80

var (
	avatarExts     = []string{"png", "jpg", "jpeg", "gif", "webp", "tiff", "heif", "heic", "avif", "jxl"}
	avatarDefaults = []string{"404", "mp", "mm", "mystery", "identicon", "monsterid", "wavatar", "retro", "robohash", "blank", "color", "letter", "initials"}
)

// Avatar is the Gravatar-compatible avatar query; Prepare derives every field
// from the path and query. Range checks and the invalid-hash fallback belong to
// biz.Resolve.
type Avatar struct {
	Hash     string `uri:"hash"`
	AppID    string `query:"-"`
	Ext      string `query:"-"`
	Size     int    `query:"-"`
	Force    bool   `query:"-"`
	Default  string `query:"-"`
	Name     string `query:"-"`
	Initials string `query:"-"`
}

// AvatarCreate is a multipart form; the json tags only name the fields in the
// OpenAPI document, which describes bodies as JSON.
type AvatarCreate struct {
	Raw        string                `form:"raw" json:"raw" validate:"required && not_exists:avatars,raw" openapi:"description=邮箱或手机号"`
	VerifyCode string                `form:"verify_code" json:"verify_code" validate:"required && verify_code:raw,avatar"`
	Avatar     *multipart.FileHeader `form:"avatar" json:"avatar" validate:"required" openapi:"description=正方形图片文件，边长至少 40px"`
	Captcha    geetest.Ticket        `form:"-" json:"captcha" validate:"required && geetest" openapi:"description=极验验证结果，JSON 字符串"`
}

// AvatarUpdate is a multipart form, like AvatarCreate.
type AvatarUpdate struct {
	Hash    string                `uri:"hash" validate:"required"`
	Avatar  *multipart.FileHeader `form:"avatar" json:"avatar" validate:"required" openapi:"description=正方形图片文件，边长至少 40px"`
	Captcha geetest.Ticket        `form:"-" json:"captcha" validate:"required && geetest" openapi:"description=极验验证结果，JSON 字符串"`
}

type AvatarDelete struct {
	Hash string `uri:"hash" validate:"required"`
}

type AvatarCheck struct {
	Raw string `query:"raw" validate:"required"`
}

type AvatarQq struct {
	Qq string `query:"qq" validate:"required && number"`
}

type AvatarBound struct {
	Bind bool `json:"bind"`
}

func (r *Avatar) Prepare(c fiber.Ctx) error {
	parts := strings.Split(r.Hash, ".")
	r.Hash = strings.ToLower(parts[0])
	r.Ext = "webp"
	if len(parts) > 1 && slices.Contains(avatarExts, parts[1]) {
		r.Ext = parts[1]
	}

	r.AppID = fiber.Query(c, "app", fiber.Query(c, "appid", ""))
	r.Size = fiber.Query(c, "s", fiber.Query(c, "size", defaultAvatarSize))
	force := fiber.Query(c, "f", fiber.Query(c, "forcedefault", "n"))
	r.Force = force == "y" || force == "yes"

	r.Default = fiber.Query(c, "d", fiber.Query(c, "default", ""))
	if !slices.Contains(avatarDefaults, r.Default) && !str.IsURL(r.Default) {
		r.Default = ""
	}
	r.Name = fiber.Query(c, "name", "")
	r.Initials = fiber.Query(c, "initials", fiber.Query(c, "letter", ""))

	return nil
}

func (r *AvatarCreate) Prepare(c fiber.Ctx) error {
	return parseCaptcha(c, &r.Captcha)
}

func (r *AvatarUpdate) Prepare(c fiber.Ctx) error {
	return parseCaptcha(c, &r.Captcha)
}

// parseCaptcha leaves a missing captcha to the required rule.
func parseCaptcha(c fiber.Ctx, ticket *geetest.Ticket) error {
	raw := c.FormValue("captcha")
	if raw == "" {
		return nil
	}
	if err := json.Unmarshal([]byte(raw), ticket); err != nil {
		return errors.New("captcha 格式错误")
	}
	return nil
}
