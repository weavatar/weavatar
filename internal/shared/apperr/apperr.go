// Package apperr defines client-facing errors whose closed set of kinds
// transports map to HTTP statuses.
package apperr

import (
	"github.com/samber/oops"
)

// Kind classifies an application error for transport mapping.
type Kind string

const (
	// KindInvalid rejects malformed input (HTTP 400).
	KindInvalid Kind = "invalid"
	// KindUnauthorized requires authentication (HTTP 401).
	KindUnauthorized Kind = "unauthorized"
	// KindForbidden denies access (HTTP 403).
	KindForbidden Kind = "forbidden"
	// KindNotFound reports a missing resource (HTTP 404).
	KindNotFound Kind = "not_found"
	// KindConflict reports a state conflict, e.g. a taken name (HTTP 409).
	KindConflict Kind = "conflict"
	// KindUnprocessable rejects a well-formed but impossible request (HTTP 422).
	KindUnprocessable Kind = "unprocessable"
)

const kindKey = "apperr.kind"

// New starts an error with code, the stable key sent to clients, and public,
// the message safe to expose; finish with .Wrap or .Errorf.
func New(kind Kind, code, public string) oops.OopsErrorBuilder {
	return oops.Code(code).Public(public).With(kindKey, kind)
}

func Invalid(code, public string) oops.OopsErrorBuilder { return New(KindInvalid, code, public) }
func Unauthorized(code, public string) oops.OopsErrorBuilder {
	return New(KindUnauthorized, code, public)
}
func Forbidden(code, public string) oops.OopsErrorBuilder { return New(KindForbidden, code, public) }
func NotFound(code, public string) oops.OopsErrorBuilder  { return New(KindNotFound, code, public) }
func Conflict(code, public string) oops.OopsErrorBuilder  { return New(KindConflict, code, public) }
func Unprocessable(code, public string) oops.OopsErrorBuilder {
	return New(KindUnprocessable, code, public)
}

// KindOf reports the error's Kind, or "" when it carries none.
func KindOf(err error) Kind {
	if oopsErr, ok := oops.AsError[oops.OopsError](err); ok {
		if kind, ok := oopsErr.Context()[kindKey].(Kind); ok {
			return kind
		}
	}
	return ""
}

// CodeOf reports the error's machine-readable code, or "" when absent.
func CodeOf(err error) string {
	if oopsErr, ok := oops.AsError[oops.OopsError](err); ok {
		code, _ := oopsErr.Code().(string)
		return code
	}
	return ""
}
