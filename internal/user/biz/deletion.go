package biz

import (
	"context"
	"strings"

	"github.com/libtnb/cache"
	"github.com/libtnb/utils/str"
	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/shared/appinfo"
	"github.com/weavatar/weavatar/internal/shared/registry"
)

// deletionStatePrefix keeps a login callback from confirming a deletion and
// vice versa.
const deletionStatePrefix = "delete-"

// DeletionUsecase is separate from UserUsecase because its cleanups come from
// modules that depend on UserUsecase.
type DeletionUsecase struct {
	repo     UserRepo
	oauth    OAuthProvider
	cache    cache.Cache
	tx       TxRunner
	cleanups registry.UserCleanups
	domain   string
	client   appinfo.OAuthClient
}

func NewDeletionUsecase(
	repo UserRepo,
	oauth OAuthProvider,
	c cache.Cache,
	tx TxRunner,
	cleanups registry.UserCleanups,
	domain appinfo.Domain,
	client appinfo.OAuthClient,
) *DeletionUsecase {
	return &DeletionUsecase{
		repo:     repo,
		oauth:    oauth,
		cache:    c,
		tx:       tx,
		cleanups: cleanups,
		domain:   string(domain),
		client:   client,
	}
}

// DeletionURL starts the re-authorization that confirms a deletion, with a
// state only userID's callback can redeem.
func (uc *DeletionUsecase) DeletionURL(_ context.Context, userID string) (string, error) {
	state := deletionStatePrefix + str.Random(16)
	if err := uc.cache.Put(state, userID, stateTTL); err != nil {
		return "", oops.In("user").Wrapf(err, "store deletion state")
	}

	return authorizeURL(uc.client, redirectURI(uc.domain), state), nil
}

// ConfirmDeletion deletes userID's account once the OAuth identity matches;
// the cleanups and the soft delete share one transaction.
func (uc *DeletionUsecase) ConfirmDeletion(ctx context.Context, userID, code, state string) error {
	if !strings.HasPrefix(state, deletionStatePrefix) || uc.cache.Pull(state) != userID {
		return ErrStateExpired()
	}

	user, err := uc.repo.Find(ctx, userID)
	if err != nil {
		return err
	}

	identity, err := uc.oauth.Exchange(ctx, code, redirectURI(uc.domain))
	if err != nil {
		return err
	}
	if identity.UnionID != user.UnionID {
		return ErrDeletionIdentityMismatch()
	}

	return uc.tx.Run(ctx, func(ctx context.Context) error {
		for _, cleanup := range uc.cleanups {
			if err := cleanup.Run(ctx, userID); err != nil {
				return oops.In("user").With("cleanup", cleanup.Name).
					Wrapf(err, "clean up %s of user %s", cleanup.Name, userID)
			}
		}
		return uc.repo.Delete(ctx, user)
	})
}
