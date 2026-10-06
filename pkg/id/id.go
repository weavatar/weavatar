// Package id generates short random identifiers.
package id

import "github.com/jaevor/go-nanoid"

const alphabet = `0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz`

// generate is safe for concurrent use; nanoid guards its buffer with a mutex.
var generate = nanoid.MustCustomASCII(alphabet, 10)

// Generate returns a random 10-character alphanumeric ID.
func Generate() string {
	return generate()
}
