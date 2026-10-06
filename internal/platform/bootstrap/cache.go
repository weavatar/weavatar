package bootstrap

import "github.com/libtnb/cache"

// NewCache builds the in-process cache shared by verify codes, login states
// and short-lived lookups.
func NewCache() cache.Cache {
	return cache.NewCache()
}
