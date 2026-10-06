// Package service adapts HTTP, CLI and the scheduler to the avatar usecase.
package service

import (
	"encoding/base64"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/libtnb/validator"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/transport"
)

// avatarMaxAge is how long clients and CDNs may reuse a served avatar.
const avatarMaxAge = 5 * time.Minute

type AvatarService struct {
	avatar   *biz.AvatarUsecase
	validate *validator.Validator
}

func NewAvatarService(avatar *biz.AvatarUsecase, validate *validator.Validator) *AvatarService {
	return &AvatarService{
		avatar:   avatar,
		validate: validate,
	}
}

// Avatar serves the avatar image of a hash.
func (r *AvatarService) Avatar(c fiber.Ctx) error {
	req, err := transport.Bind[Avatar](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	res, err := r.avatar.Resolve(c.Context(), biz.ResolveRequest{
		Hash:     req.Hash,
		AppID:    req.AppID,
		Ext:      req.Ext,
		Size:     req.Size,
		Force:    req.Force,
		Default:  req.Default,
		Name:     req.Name,
		Initials: req.Initials,
	})
	if err != nil {
		return transport.ErrorFrom(c, err)
	}
	if res.NotFound {
		return c.Status(fiber.StatusNotFound).SendString("404 Not Found\nWeAvatar")
	}
	if res.Redirect != "" {
		return c.Redirect().Status(fiber.StatusFound).To(res.Redirect)
	}

	c.Vary(fiber.HeaderAcceptEncoding, fiber.HeaderAccept)
	c.Set("X-Avatar-By", "weavatar.com")
	c.Set("X-Avatar-From", res.From)
	c.Set(fiber.HeaderCacheControl, "public, max-age="+strconv.Itoa(int(avatarMaxAge.Seconds())))
	c.Set(fiber.HeaderLastModified, res.LastModified.UTC().Format(http.TimeFormat))
	c.Set(fiber.HeaderExpires, time.Now().UTC().Add(avatarMaxAge).Format(http.TimeFormat))

	return c.Type(req.Ext).Send(res.Image)
}

func (r *AvatarService) List(c fiber.Ctx) error {
	req, err := transport.Bind[transport.Paginate](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	avatars, total, err := r.avatar.List(c.Context(), transport.UserID(c), req.Page, req.Limit)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, transport.Page[*biz.Avatar]{
		Total: total,
		Items: avatars,
	})
}

func (r *AvatarService) Create(c fiber.Ctx) error {
	req, err := transport.Bind[AvatarCreate](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}
	img, err := readFile(req.Avatar)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	avatar, err := r.avatar.Create(c.Context(), transport.UserID(c), req.Raw, img)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, avatar)
}

func (r *AvatarService) Update(c fiber.Ctx) error {
	req, err := transport.Bind[AvatarUpdate](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}
	img, err := readFile(req.Avatar)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	avatar, err := r.avatar.Update(c.Context(), transport.UserID(c), req.Hash, img)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, avatar)
}

func (r *AvatarService) Delete(c fiber.Ctx) error {
	req, err := transport.Bind[AvatarDelete](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	if err = r.avatar.Delete(c.Context(), transport.UserID(c), req.Hash); err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success[any](c, nil)
}

// Check reports whether an email or phone already has an avatar.
func (r *AvatarService) Check(c fiber.Ctx) error {
	req, err := transport.Bind[AvatarCheck](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	bound, err := r.avatar.Bound(c.Context(), req.Raw)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, AvatarBound{Bind: bound})
}

// Qq returns a QQ number's avatar as base64, for importing it as an upload.
func (r *AvatarService) Qq(c fiber.Ctx) error {
	req, err := transport.Bind[AvatarQq](c, r.validate)
	if err != nil {
		return transport.Error(c, fiber.StatusUnprocessableEntity, "%v", err)
	}

	img, err := r.avatar.FetchQQ(c.Context(), req.Qq)
	if err != nil {
		return transport.ErrorFrom(c, err)
	}

	return transport.Success(c, base64.StdEncoding.EncodeToString(img))
}

func readFile(header *multipart.FileHeader) ([]byte, error) {
	f, err := header.Open()
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	return io.ReadAll(f)
}
