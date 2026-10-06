package biz_test

import (
	"bytes"
	"context"
	"image"
	"image/png"
	"testing"

	"github.com/go-rio/rio"
	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
	"github.com/libtnb/utils/str"

	"github.com/weavatar/weavatar/internal/avatar/biz"
	"github.com/weavatar/weavatar/internal/shared/apperr"
)

func TestCreate_StoresAndQueuesPurgeOfBothURLs(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.CreateFunc = func(context.Context, *biz.Avatar) error { return nil }
	d.store.WriteAvatarFunc = func(string, []byte) error { return nil }
	purged := make(chan []string, 1)
	d.purger.RefreshFunc = func(_ context.Context, urls []string) error {
		purged <- urls
		return nil
	}
	img := pngOf(t, 100)

	avatar, err := uc.Create(t.Context(), "u1", "a@example.com", img)

	must.NoError(t, err)
	check.Equal(t, avatar.SHA256, str.SHA256("a@example.com"))
	check.Equal(t, avatar.MD5, str.MD5("a@example.com"))
	check.Equal(t, avatar.UserID, "u1")
	check.Len(t, d.tx.RunCalls(), 1)
	written := d.store.WriteAvatarCalls()
	must.Len(t, written, 1)
	check.Equal(t, written[0].Sha256, avatar.SHA256)
	check.True(t, bytes.Equal(written[0].Img, img)) // small uploads keep their bytes
	check.Equal(t, d.queue.Len(), 1)                // the request does not wait on the CDN

	must.NoError(t, d.queue.Start())
	check.DeepEqual(t, receive(t, purged), []string{
		"https://weavatar.com/avatar/" + avatar.SHA256,
		"https://weavatar.com/avatar/" + avatar.MD5,
	})
}

func TestCreate_ShrinksLargeUploads(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.CreateFunc = func(context.Context, *biz.Avatar) error { return nil }
	d.store.WriteAvatarFunc = func(string, []byte) error { return nil }

	_, err := uc.Create(t.Context(), "u1", "a@example.com", pngOf(t, 2100))

	must.NoError(t, err)
	check.Equal(t, sizeOf(t, d.store.WriteAvatarCalls()[0].Img), 2048)
}

func TestCreate_RejectsBadImages(t *testing.T) {
	var wide bytes.Buffer
	must.NoError(t, png.Encode(&wide, image.NewRGBA(image.Rect(0, 0, 100, 50))))

	tests := []struct {
		name string
		img  []byte
		code string
	}{
		{name: "not an image", img: []byte("nope"), code: "avatar.invalid_image"},
		{name: "not square", img: wide.Bytes(), code: "avatar.not_square"},
		{name: "too small", img: pngOf(t, 39), code: "avatar.too_small"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc, _ := newUsecase(t) // repo funcs stay nil: nothing may be written

			_, err := uc.Create(t.Context(), "u1", "a@example.com", tt.img)

			check.Equal(t, apperr.KindOf(err), apperr.KindInvalid)
			check.Equal(t, apperr.CodeOf(err), tt.code)
		})
	}
}

func TestCreate_TakenRawMapsToConflict(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.CreateFunc = func(context.Context, *biz.Avatar) error { return biz.ErrAvatarExists }

	_, err := uc.Create(t.Context(), "u1", "a@example.com", pngOf(t, 100))

	check.Equal(t, apperr.KindOf(err), apperr.KindConflict)
	check.Len(t, d.store.WriteAvatarCalls(), 0)
	check.Equal(t, d.queue.Len(), 0)
}

func TestUpdate_TouchesAndReplacesImage(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.FindFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return &biz.Avatar{SHA256: sha256Hash, MD5: md5Hash, UserID: "u1"}, nil
	}
	d.repo.TouchFunc = func(context.Context, *biz.Avatar) error { return nil }
	d.store.WriteAvatarFunc = func(string, []byte) error { return nil }

	_, err := uc.Update(t.Context(), "u1", md5Hash, pngOf(t, 100))

	must.NoError(t, err)
	found := d.repo.FindCalls()
	must.Len(t, found, 1)
	check.Equal(t, found[0].UserID, "u1")
	check.Equal(t, found[0].Hash, md5Hash)
	check.Len(t, d.repo.TouchCalls(), 1)
	check.Equal(t, d.store.WriteAvatarCalls()[0].Sha256, sha256Hash)
	check.Equal(t, d.queue.Len(), 1)
}

func TestUpdate_OthersAvatarIsNotFound(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.FindFunc = func(context.Context, string, string) (*biz.Avatar, error) { return nil, rio.ErrNotFound }

	_, err := uc.Update(t.Context(), "u2", md5Hash, pngOf(t, 100))

	must.ErrorIs(t, err, rio.ErrNotFound)
}

func TestDelete_RemovesRowAndFile(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.FindFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return &biz.Avatar{SHA256: sha256Hash, MD5: md5Hash, UserID: "u1"}, nil
	}
	d.repo.DeleteFunc = func(context.Context, *biz.Avatar) error { return nil }
	d.store.RemoveAvatarFunc = func(string) error { return nil }

	must.NoError(t, uc.Delete(t.Context(), "u1", sha256Hash))

	check.Len(t, d.tx.RunCalls(), 1)
	check.Equal(t, d.repo.DeleteCalls()[0].Avatar.SHA256, sha256Hash)
	check.Equal(t, d.store.RemoveAvatarCalls()[0].Sha256, sha256Hash)
	check.Equal(t, d.queue.Len(), 1)
}

func TestBound(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.ExistsByRawFunc = func(_ context.Context, raw string) (bool, error) { return raw == "a@example.com", nil }

	bound, err := uc.Bound(t.Context(), "a@example.com")
	must.NoError(t, err)
	check.True(t, bound)

	bound, err = uc.Bound(t.Context(), "b@example.com")
	must.NoError(t, err)
	check.False(t, bound)
}

func TestAudit_RecordsVerdictAndPurgesBanned(t *testing.T) {
	uc, d := newUsecase(t)
	d.noUpload()
	d.noCache()
	img := pngOf(t, 80)
	d.fetcher.GravatarFunc = func(context.Context, string) ([]byte, error) { return img, nil }
	d.store.WriteCheckerFunc = func(string, []byte) error { return nil }
	d.images.FindFunc = func(context.Context, string) (*biz.Image, error) { return nil, rio.ErrNotFound }
	d.images.CreateFunc = func(context.Context, *biz.Image) error { return nil }
	d.auditor.CheckFunc = func(context.Context, string) (bool, string, error) { return true, "porn", nil }
	d.purger.RefreshFunc = func(context.Context, []string) error { return nil }

	must.NoError(t, uc.Audit(t.Context(), md5Hash, ""))

	check.Equal(t, d.auditor.CheckCalls()[0].URL, "https://weavatar.com/avatar/"+md5Hash+".png?s=600&d=404")
	created := d.images.CreateCalls()
	must.Len(t, created, 1)
	check.True(t, created[0].Image.Banned)
	check.Equal(t, created[0].Image.Remark, "porn")
	check.Equal(t, created[0].Image.Hash, d.store.WriteCheckerCalls()[0].ImgHash)
	check.DeepEqual(t, d.purger.RefreshCalls()[0].Urls, []string{"https://weavatar.com/avatar/" + md5Hash})
}

func TestAudit_ReusesRecordedVerdict(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.FindForServeFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		return &biz.Avatar{SHA256: sha256Hash}, nil
	}
	d.store.ReadAvatarFunc = func(string) ([]byte, error) { return pngOf(t, 80), nil }
	d.store.WriteCheckerFunc = func(string, []byte) error { return nil }
	d.audited(false)
	// auditor.CheckFunc stays nil: a second moderation call would panic

	must.NoError(t, uc.Audit(t.Context(), sha256Hash, ""))

	check.Len(t, d.purger.RefreshCalls(), 0)
}

func TestAudit_ChecksTheAppOverride(t *testing.T) {
	uc, d := newUsecase(t)
	d.repo.FindForServeFunc = func(context.Context, string, string) (*biz.Avatar, error) {
		avatar := &biz.Avatar{SHA256: sha256Hash}
		avatar.AppSHA256.Set(&biz.AppAvatar{AppID: "app 1"})
		avatar.AppMD5.Set(nil)
		return avatar, nil
	}
	d.store.ReadAppAvatarFunc = func(string, string) ([]byte, error) { return pngOf(t, 80), nil }
	d.store.WriteCheckerFunc = func(string, []byte) error { return nil }
	d.images.FindFunc = func(context.Context, string) (*biz.Image, error) { return nil, rio.ErrNotFound }
	d.images.CreateFunc = func(context.Context, *biz.Image) error { return nil }
	d.auditor.CheckFunc = func(context.Context, string) (bool, string, error) { return false, "", nil }

	must.NoError(t, uc.Audit(t.Context(), sha256Hash, "app 1"))

	check.Equal(t, d.auditor.CheckCalls()[0].URL, "https://weavatar.com/avatar/"+sha256Hash+".png?s=600&d=404&app=app+1")
}
