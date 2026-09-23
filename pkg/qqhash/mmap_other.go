//go:build !unix

package qqhash

import (
	"fmt"
	"os"
)

func mapFile(path string) ([]byte, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(b) == 0 {
		return nil, fmt.Errorf("%s: %w: empty file", path, ErrCorrupt)
	}
	return b, nil
}

func unmapFile([]byte) error {
	return nil
}

func adviseRandom([]byte) {}

func adviseWillNeed([]byte) {}
