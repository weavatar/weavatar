package service

import (
	"encoding/base64"
	"net/http"
	"path/filepath"
	"time"

	"github.com/davidbyttow/govips/v2/vips"
	"github.com/go-rat/utils/str"
	"github.com/gofiber/fiber/v3"

	"github.com/weavatar/weavatar/internal/biz"
	"github.com/weavatar/weavatar/internal/http/request"
	"github.com/weavatar/weavatar/pkg/avatars"
	"github.com/weavatar/weavatar/pkg/embed"
)

type AvatarService struct {
	avatarRepo biz.AvatarRepo
}

func NewAvatarService(avatar biz.AvatarRepo) *AvatarService {
	return &AvatarService{
		avatarRepo: avatar,
	}
}

func (r *AvatarService) Avatar(c fiber.Ctx) error {
	req, err := Bind[request.Avatar](c)
	if err != nil {
		return Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	var avatar []byte
	var lmt time.Time
	from := "weavatar"
	nickname := ""

	// 快速路径
	if req.Force {
		if req.Default == "404" {
			return c.Status(fiber.StatusNotFound).SendString("404 Not Found\nWeAvatar")
		}
		if str.IsURL(req.Default) {
			return c.Redirect().Status(fiber.StatusFound).To(req.Default)
		}
	}

	nickname, avatar, lmt, err = r.avatarRepo.GetWeAvatar(req.Hash, req.AppID)
	if err != nil && !req.Force {
		avatar, lmt, err = r.avatarRepo.GetGravatarByHash(req.Hash)
		from = "gravatar"
	}
	if err != nil && !req.Force {
		_, avatar, lmt, err = r.avatarRepo.GetQqByHash(req.Hash)
		from = "qq"
	}
	if err == nil && (from == "weavatar" || from == "gravatar") {
		if ban, _ := r.avatarRepo.IsBanned(avatar); ban {
			avatar, err = embed.DefaultFS.ReadFile(filepath.Join("default", "ban.png"))
			lmt = time.Now()
		}
	}

	// 如果前面取不到头像或者要求强制默认头像
	if err != nil || avatar == nil || req.Force {
		options := []string{req.Hash}
		if req.Default == "letter" || req.Default == "initials" {
			initials := c.Query("initials", c.Query("letter")) // TODO deprecated letter in the future
			if initials == "" {
				name := c.Query("name", nickname) // 保持和 Gravatar 一致，name 取第一位
				initials = r.getFirstRune(name)
			}
			options = append(options, initials)
		}
		if options[0] == "" {
			options[0] = "weavatar" // 默认赋值，防止颜色乱跳
		}
		avatar, lmt, err = r.avatarRepo.GetByType(req.Default, options...)
		from = "weavatar"
	}

	if err != nil {
		return ErrorSystem(c)
	}

	avatar, err = r.convert(avatar, req.Ext, req.Size)
	if err != nil {
		return ErrorSystem(c)
	}

	c.Vary("Accept-Encoding", "Accept")
	c.Set("X-Avatar-By", "weavatar.com")
	c.Set("X-Avatar-From", from)
	c.Set("Cache-Control", "public, max-age=300")
	c.Set("Last-Modified", lmt.UTC().Format(http.TimeFormat))
	c.Set("Expires", time.Now().UTC().Add(5*time.Minute).Format(http.TimeFormat))

	return c.Type(req.Ext).Send(avatar)
}

func (r *AvatarService) Check(c fiber.Ctx) error {
	req, err := Bind[request.AvatarCheck](c)
	if err != nil {
		return Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	avatar, err := r.avatarRepo.GetByRaw(req.Raw)
	if err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	return Success(c, fiber.Map{
		"bind": avatar.UserID != "",
	})
}

func (r *AvatarService) Qq(c fiber.Ctx) error {
	req, err := Bind[request.AvatarQq](c)
	if err != nil {
		return Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	avatar, err := avatars.Qq(req.Qq)
	if err != nil {
		return Error(c, fiber.StatusInternalServerError, "%v", err)
	}

	return Success(c, base64.StdEncoding.EncodeToString(avatar))
}

func (r *AvatarService) convert(avatar []byte, ext string, size int) ([]byte, error) {
	img, err := vips.NewImageFromBuffer(avatar)
	if err != nil {
		return nil, err
	}

	if err = img.Thumbnail(size, size, vips.InterestingAttention); err != nil {
		return nil, err
	}

	var data []byte
	switch ext {
	case "jpg", "jpeg":
		data, _, err = img.ExportJpeg(vips.NewJpegExportParams())
	case "png":
		data, _, err = img.ExportPng(vips.NewPngExportParams())
	case "webp":
		data, _, err = img.ExportWebp(vips.NewWebpExportParams())
	case "heif", "heic":
		data, _, err = img.ExportHeif(vips.NewHeifExportParams())
	case "tiff":
		data, _, err = img.ExportTiff(vips.NewTiffExportParams())
	case "avif":
		data, _, err = img.ExportAvif(vips.NewAvifExportParams())
	case "gif":
		data, _, err = img.ExportGIF(vips.NewGifExportParams())
	case "jxl":
		data, _, err = img.ExportJxl(vips.NewJxlExportParams())
	default:
		data, _, err = img.ExportWebp(vips.NewWebpExportParams())
	}

	return data, err
}

func (r *AvatarService) getFirstRune(s string) string {
	runes := []rune(s)
	if len(runes) > 0 {
		return string(runes[0])
	}
	return "?"
}
