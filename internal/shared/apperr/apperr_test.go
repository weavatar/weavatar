package apperr_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/libtnb/assert/must"
	"github.com/samber/oops"

	"github.com/weavatar/weavatar/internal/shared/apperr"
)

var errSentinel = errors.New("sentinel")

func TestKindAndCodeSurviveWrapping(t *testing.T) {
	err := apperr.Conflict("user.name_taken", "name already taken").In("user").Wrap(errSentinel)

	must.Equal(t, apperr.KindOf(err), apperr.KindConflict)
	must.Equal(t, apperr.CodeOf(err), "user.name_taken")
	must.ErrorIs(t, err, errSentinel)

	wrapped := fmt.Errorf("placing order: %w", err)
	must.Equal(t, apperr.KindOf(wrapped), apperr.KindConflict)
	must.Equal(t, apperr.CodeOf(wrapped), "user.name_taken")
}

func TestPlainErrorsCarryNoKind(t *testing.T) {
	must.Equal(t, apperr.KindOf(errors.New("boom")), apperr.Kind(""))
	must.Empty(t, apperr.CodeOf(errors.New("boom")))
	must.Equal(t, apperr.KindOf(nil), apperr.Kind(""))
}

func TestHelpersSetTheirKinds(t *testing.T) {
	for kind, helper := range map[apperr.Kind]func(code, public string) oops.OopsErrorBuilder{
		apperr.KindInvalid:       apperr.Invalid,
		apperr.KindUnauthorized:  apperr.Unauthorized,
		apperr.KindForbidden:     apperr.Forbidden,
		apperr.KindNotFound:      apperr.NotFound,
		apperr.KindConflict:      apperr.Conflict,
		apperr.KindUnprocessable: apperr.Unprocessable,
	} {
		must.Equal(t, apperr.KindOf(helper("c", "p").Errorf("x")), kind, string(kind))
	}
}
