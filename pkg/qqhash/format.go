// Package qqhash 提供 QQ 邮箱哈希到 QQ 号的映射表
package qqhash

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"math"
	"path/filepath"
	"strconv"
	"unsafe"
)

const (
	TypeMD5    = "md5"
	TypeSHA256 = "sha256"

	Version = 1

	MinQq = 10000
	MaxQq = math.MaxUint32 // 值数组是 uint32

	headerSize         = 4096
	partitionEntrySize = 40
	blobAlign          = 64
	maxPartBits        = 16
	emailSuffix        = "@qq.com"
)

var Types = []string{TypeMD5, TypeSHA256}

var idxMagic = [8]byte{'W', 'A', 'Q', 'Q', 'M', 'P', 'H', '1'}

var (
	ErrNotFound  = errors.New("qqhash: not found")
	ErrCorrupt   = errors.New("qqhash: corrupt file")
	ErrBigEndian = errors.New("qqhash: big-endian hosts are not supported")
)

// 文件头字段偏移
const (
	offMagic     = 0  // [8]byte
	offVersion   = 8  // uint32
	offKeyBytes  = 12 // uint32
	offStart     = 16 // uint64
	offEnd       = 24 // uint64
	offKeyCount  = 32 // uint64
	offPartBits  = 40 // uint32
	offGamma     = 48 // float64
	offSeed      = 56 // uint64
	offBuildTime = 64 // int64
	offIdxSize   = 72 // uint64
	offValSize   = 80 // uint64
	offCRC       = 88 // uint32，覆盖文件头前 88 字节与整张分区表
)

type header struct {
	keyBytes  uint32
	start     uint64
	end       uint64
	keyCount  uint64
	partBits  uint32
	gamma     float64
	seed      uint64
	buildTime int64
	idxSize   uint64
	valSize   uint64
}

type partition struct {
	keyCount   uint64
	slotOffset uint64
	mphOffset  uint64
	mphLen     uint64
	crc        uint32
}

func (h *header) partitions() int {
	return 1 << h.partBits
}

func (h *header) tableSize() int64 {
	return int64(h.partitions()) * partitionEntrySize
}

func (h *header) blobBase() int64 {
	return alignUp(headerSize+h.tableSize(), headerSize)
}

func encodeIndexHead(h *header, parts []partition) []byte {
	b := make([]byte, h.blobBase())
	copy(b[offMagic:], idxMagic[:])
	binary.LittleEndian.PutUint32(b[offVersion:], Version)
	binary.LittleEndian.PutUint32(b[offKeyBytes:], h.keyBytes)
	binary.LittleEndian.PutUint64(b[offStart:], h.start)
	binary.LittleEndian.PutUint64(b[offEnd:], h.end)
	binary.LittleEndian.PutUint64(b[offKeyCount:], h.keyCount)
	binary.LittleEndian.PutUint32(b[offPartBits:], h.partBits)
	binary.LittleEndian.PutUint64(b[offGamma:], math.Float64bits(h.gamma))
	binary.LittleEndian.PutUint64(b[offSeed:], h.seed)
	binary.LittleEndian.PutUint64(b[offBuildTime:], uint64(h.buildTime))
	binary.LittleEndian.PutUint64(b[offIdxSize:], h.idxSize)
	binary.LittleEndian.PutUint64(b[offValSize:], h.valSize)

	table := b[headerSize:]
	for i, p := range parts {
		e := table[i*partitionEntrySize:]
		binary.LittleEndian.PutUint64(e[0:], p.keyCount)
		binary.LittleEndian.PutUint64(e[8:], p.slotOffset)
		binary.LittleEndian.PutUint64(e[16:], p.mphOffset)
		binary.LittleEndian.PutUint64(e[24:], p.mphLen)
		binary.LittleEndian.PutUint32(e[32:], p.crc)
	}

	binary.LittleEndian.PutUint32(b[offCRC:], headCRC(b, h))
	return b
}

func headCRC(b []byte, h *header) uint32 {
	c := crc32.ChecksumIEEE(b[:offCRC])
	return crc32.Update(c, crc32.IEEETable, b[headerSize:headerSize+h.tableSize()])
}

func decodeIndexHead(b []byte) (*header, []partition, error) {
	if len(b) < headerSize || [8]byte(b[:8]) != idxMagic {
		return nil, nil, fmt.Errorf("%w: bad magic", ErrCorrupt)
	}
	if v := binary.LittleEndian.Uint32(b[offVersion:]); v != Version {
		return nil, nil, fmt.Errorf("%w: unsupported version %d", ErrCorrupt, v)
	}

	h := &header{
		keyBytes:  binary.LittleEndian.Uint32(b[offKeyBytes:]),
		start:     binary.LittleEndian.Uint64(b[offStart:]),
		end:       binary.LittleEndian.Uint64(b[offEnd:]),
		keyCount:  binary.LittleEndian.Uint64(b[offKeyCount:]),
		partBits:  binary.LittleEndian.Uint32(b[offPartBits:]),
		gamma:     math.Float64frombits(binary.LittleEndian.Uint64(b[offGamma:])),
		seed:      binary.LittleEndian.Uint64(b[offSeed:]),
		buildTime: int64(binary.LittleEndian.Uint64(b[offBuildTime:])),
		idxSize:   binary.LittleEndian.Uint64(b[offIdxSize:]),
		valSize:   binary.LittleEndian.Uint64(b[offValSize:]),
	}
	if h.keyBytes != md5.Size && h.keyBytes != sha256.Size {
		return nil, nil, fmt.Errorf("%w: bad key size %d", ErrCorrupt, h.keyBytes)
	}
	if h.partBits == 0 || h.partBits > maxPartBits {
		return nil, nil, fmt.Errorf("%w: bad partition bits %d", ErrCorrupt, h.partBits)
	}
	if h.start > h.end || h.end > MaxQq || h.keyCount != h.end-h.start+1 || h.valSize != h.keyCount*4 {
		return nil, nil, fmt.Errorf("%w: inconsistent key range", ErrCorrupt)
	}
	if int64(len(b)) < h.blobBase() || h.idxSize < uint64(h.blobBase()) {
		return nil, nil, fmt.Errorf("%w: file too short", ErrCorrupt)
	}
	if binary.LittleEndian.Uint32(b[offCRC:]) != headCRC(b, h) {
		return nil, nil, fmt.Errorf("%w: header checksum mismatch", ErrCorrupt)
	}

	parts := make([]partition, h.partitions())
	table := b[headerSize:]
	var slots uint64
	for i := range parts {
		e := table[i*partitionEntrySize:]
		p := partition{
			keyCount:   binary.LittleEndian.Uint64(e[0:]),
			slotOffset: binary.LittleEndian.Uint64(e[8:]),
			mphOffset:  binary.LittleEndian.Uint64(e[16:]),
			mphLen:     binary.LittleEndian.Uint64(e[24:]),
			crc:        binary.LittleEndian.Uint32(e[32:]),
		}
		if p.slotOffset != slots || p.keyCount > h.keyCount-slots {
			return nil, nil, fmt.Errorf("%w: partition %d slots out of range", ErrCorrupt, i)
		}
		if p.mphOffset < uint64(h.blobBase()) || p.mphOffset%8 != 0 || p.mphLen%8 != 0 || p.mphLen > h.idxSize-p.mphOffset {
			return nil, nil, fmt.Errorf("%w: partition %d blob out of range", ErrCorrupt, i)
		}
		slots += p.keyCount
		parts[i] = p
	}
	if slots != h.keyCount {
		return nil, nil, fmt.Errorf("%w: partition slots do not sum to key count", ErrCorrupt)
	}

	return h, parts, nil
}

func keyBytesOf(typ string) (int, error) {
	switch typ {
	case TypeMD5:
		return md5.Size, nil
	case TypeSHA256:
		return sha256.Size, nil
	}
	return 0, fmt.Errorf("qqhash: unsupported hash type %q", typ)
}

func typeOfHashLen(n int) (string, bool) {
	switch n {
	case md5.Size * 2:
		return TypeMD5, true
	case sha256.Size * 2:
		return TypeSHA256, true
	}
	return "", false
}

func fileNames(dir, typ string) (idx, val string) {
	return filepath.Join(dir, "qq_"+typ+".idx"), filepath.Join(dir, "qq_"+typ+".val")
}

func alignUp(v, a int64) int64 {
	return (v + a - 1) &^ (a - 1)
}

func appendEmail(buf []byte, qq uint64) []byte {
	buf = strconv.AppendUint(buf, qq, 10)
	return append(buf, emailSuffix...)
}

func digestOf(typ string, email []byte, out *[sha256.Size]byte) []byte {
	switch typ {
	case TypeMD5:
		sum := md5.Sum(email)
		copy(out[:], sum[:])
		return out[:md5.Size]
	case TypeSHA256:
		*out = sha256.Sum256(email)
		return out[:]
	}
	return nil
}

func digestQq(typ string, qq uint64, out *[sha256.Size]byte) []byte {
	var buf [32]byte
	return digestOf(typ, appendEmail(buf[:0], qq), out)
}

func partitionOf(d []byte, partBits uint32) uint32 {
	return uint32(binary.BigEndian.Uint16(d[:2]) >> (16 - partBits))
}

// 不用 encoding/hex 是为了避免分配
func decodeHex(dst []byte, src string) bool {
	if len(src) != len(dst)*2 {
		return false
	}
	for i := range dst {
		hi, lo := hexNibble(src[2*i]), hexNibble(src[2*i+1])
		if hi > 15 || lo > 15 {
			return false
		}
		dst[i] = hi<<4 | lo
	}
	return true
}

func hexNibble(c byte) byte {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 255
}

var nativeLittleEndian = func() bool {
	var x uint16 = 1
	return *(*byte)(unsafe.Pointer(&x)) == 1
}()
