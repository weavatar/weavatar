//go:build unix

package qqhash

import (
	"fmt"
	"os"

	"golang.org/x/sys/unix"
)

// mapFile 以只读方式把整个文件映射到内存。
func mapFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		_ = f.Close()
	}()

	st, err := f.Stat()
	if err != nil {
		return nil, err
	}
	if st.Size() <= 0 {
		return nil, fmt.Errorf("%s: %w: empty file", path, ErrCorrupt)
	}

	b, err := unix.Mmap(int(f.Fd()), 0, int(st.Size()), unix.PROT_READ, unix.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("%s: mmap: %w", path, err)
	}
	return b, nil
}

func unmapFile(b []byte) error {
	return unix.Munmap(b)
}

// adviseRandom 告知内核随机访问，避免无谓的预读。
func adviseRandom(b []byte) {
	_ = unix.Madvise(b, unix.MADV_RANDOM)
}

// adviseWillNeed 告知内核尽快把整段数据读入页缓存。
func adviseWillNeed(b []byte) {
	_ = unix.Madvise(b, unix.MADV_WILLNEED)
}
