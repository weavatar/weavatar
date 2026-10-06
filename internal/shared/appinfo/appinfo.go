// Package appinfo carries configuration values to business modules, which
// cannot import platform/conf; each has its own type because wire keys
// bindings by type.
package appinfo

import "time"

// Domain is the public host (http.domain), without scheme, e.g. "weavatar.com".
type Domain string

// HashDir is the QQ hash table directory (hash.dir).
type HashDir string

// GravatarURL is the Gravatar origin or mirror (gravatar.url), without a trailing slash.
type GravatarURL string

// CodeExpire is how long a verification code stays valid (code.expire).
type CodeExpire time.Duration

// OAuthClient identifies this application to the OAuth server.
type OAuthClient struct {
	BaseURL  string
	ClientID string
}
