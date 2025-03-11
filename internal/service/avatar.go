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
	options := []string{req.Hash}
	if req.Default == "letter" || req.Default == "initials" {
		options = append(options, req.Default)
	}

	// 快速路径
	if req.Force {
		if req.Default == "404" {
			return c.Status(fiber.StatusNotFound).SendString("404 Not Found\nWeAvatar")
		}
		if str.IsURL(req.Default) {
			return c.Redirect().Status(fiber.StatusFound).To(req.Default)
		}
	}

	if !req.Force {
		avatar, lmt, err = r.avatarRepo.GetWeAvatar(req.Hash, req.AppID)
		if err != nil {
			avatar, lmt, err = r.avatarRepo.GetGravatarByHash(req.Hash)
			from = "gravatar"
		}
		if err != nil {
			_, avatar, lmt, err = r.avatarRepo.GetQqByHash(req.Hash)
			from = "qq"
		}
	}

	if from == "gravatar" && err == nil {
		if ban, _ := r.avatarRepo.IsBanned(avatar); ban {
			avatar, err = embed.DefaultFS.ReadFile(filepath.Join("default", "ban.png"))
			lmt = time.Now()
		}
	}

	if err != nil || avatar == nil {
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
