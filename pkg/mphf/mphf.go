// Package mphf implements a BBHash-style minimal perfect hash function.
//
// Serialized layout, as little-endian uint64 words:
//
//	[0]        key count n
//	[1]        seed
//	[2]        level count L
//	[3, 3+2L)  per level { block count, keys placed by earlier levels }
//	[3+2L, …)  per-level blocks of 8 words: the rank before the block, then a 7-word bit vector
package mphf

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/bits"
	"slices"
	"unsafe"
)

const (
	blockWords = 8
	blockBits  = (blockWords - 1) * 64 // 448

	headerWords      = 3
	levelHeaderWords = 2

	DefaultGamma     = 2.0
	DefaultMaxLevels = 32

	maxAttempts = 16
	golden      = 0x9E3779B97F4A7C15
)

var (
	ErrDuplicateKey = errors.New("mphf: duplicate key")
	ErrTooManyKeys  = errors.New("mphf: too many keys")
	ErrBuildFailed  = errors.New("mphf: build failed after retries")
	ErrCorrupt      = errors.New("mphf: corrupt data")
)

type Options struct {
	Gamma     float64 // bits per key, default 2.0; larger is faster to query but bigger
	MaxLevels int     // default 32; exceeding it retries with another seed
	Seed      uint64
}

type MPHF struct {
	words    []uint64 // backs every levels[i].data
	keyCount uint64
	seed     uint64
	levels   []level
}

type level struct {
	data       []uint64
	bits       uint64
	keysBefore uint64
	seed       uint64
}

// Build also returns the slot of every key, in key order.
func Build(keys []byte, keySize int, opts Options) (*MPHF, []uint32, error) {
	if keySize <= 0 || len(keys)%keySize != 0 {
		return nil, nil, fmt.Errorf("mphf: invalid key size %d for %d bytes", keySize, len(keys))
	}
	n := len(keys) / keySize
	if uint64(n) >= math.MaxUint32 {
		return nil, nil, ErrTooManyKeys
	}

	gamma := opts.Gamma
	if gamma <= 0 {
		gamma = DefaultGamma
	}
	if gamma < 1 {
		gamma = 1
	}
	maxLevels := opts.MaxLevels
	if maxLevels <= 0 {
		maxLevels = DefaultMaxLevels
	}

	slots := make([]uint32, n)
	for attempt := range uint64(maxAttempts) {
		seed := mix64(opts.Seed*golden + (attempt+1)*0xD6E8FEB86659FD93)
		levels, leftover := buildLevels(keys, keySize, slots, gamma, maxLevels, seed)
		if len(leftover) == 0 {
			return assemble(uint64(n), seed, levels), slots, nil
		}
		// no seed can separate equal keys, so rule them out before retrying
		if hasDuplicate(keys, keySize, leftover) {
			return nil, nil, ErrDuplicateKey
		}
	}

	return nil, nil, ErrBuildFailed
}

type builtLevel struct {
	data       []uint64
	keysBefore uint64
	seed       uint64
}

func buildLevels(keys []byte, keySize int, slots []uint32, gamma float64, maxLevels int, seed uint64) ([]builtLevel, []uint32) {
	cur := make([]uint32, len(slots))
	for i := range cur {
		cur[i] = uint32(i)
	}

	var levels []builtLevel
	var keysBefore uint64
	for lvl := 0; lvl < maxLevels && len(cur) > 0; lvl++ {
		blocks := uint64(math.Ceil(gamma * float64(len(cur)) / blockBits))
		if blocks == 0 {
			blocks = 1
		}
		nbits := blocks * blockBits
		lseed := levelSeed(seed, lvl)

		// seen marks bits hit at least once, coll bits hit more than once
		flat := blocks * (blockWords - 1)
		seen := make([]uint64, flat)
		coll := make([]uint64, flat)
		for _, idx := range cur {
			pos := position(hash64(keyAt(keys, keySize, idx), lseed), nbits)
			w, b := pos>>6, uint64(1)<<(pos&63)
			if seen[w]&b != 0 {
				coll[w] |= b
			} else {
				seen[w] |= b
			}
		}

		// bits hit exactly once place their keys on this level
		data := make([]uint64, blocks*blockWords)
		var placed uint64
		for b := range blocks {
			data[b*blockWords] = placed
			for k := range uint64(blockWords - 1) {
				w := seen[b*(blockWords-1)+k] &^ coll[b*(blockWords-1)+k]
				data[b*blockWords+1+k] = w
				placed += uint64(bits.OnesCount64(w))
			}
		}

		// colliding keys move to the next level in place; recording the slots of
		// the rest now saves a Find per key after the build
		next := cur[:0]
		for _, idx := range cur {
			pos := position(hash64(keyAt(keys, keySize, idx), lseed), nbits)
			if coll[pos>>6]&(uint64(1)<<(pos&63)) != 0 {
				next = append(next, idx)
				continue
			}
			slots[idx] = uint32(keysBefore + rank(data, pos))
		}

		levels = append(levels, builtLevel{data: data, keysBefore: keysBefore, seed: lseed})
		keysBefore += placed
		cur = next
	}

	return levels, cur
}

func hasDuplicate(keys []byte, keySize int, idx []uint32) bool {
	sorted := slices.Clone(idx)
	slices.SortFunc(sorted, func(a, b uint32) int {
		return bytes.Compare(keyAt(keys, keySize, a), keyAt(keys, keySize, b))
	})
	for i := 1; i < len(sorted); i++ {
		if bytes.Equal(keyAt(keys, keySize, sorted[i-1]), keyAt(keys, keySize, sorted[i])) {
			return true
		}
	}
	return false
}

func assemble(n, seed uint64, built []builtLevel) *MPHF {
	total := headerWords + levelHeaderWords*len(built)
	for _, l := range built {
		total += len(l.data)
	}

	words := make([]uint64, total)
	words[0] = n
	words[1] = seed
	words[2] = uint64(len(built))

	m := &MPHF{words: words, keyCount: n, seed: seed, levels: make([]level, len(built))}
	off := headerWords + levelHeaderWords*len(built)
	for i, l := range built {
		blocks := uint64(len(l.data) / blockWords)
		words[headerWords+levelHeaderWords*i] = blocks
		words[headerWords+levelHeaderWords*i+1] = l.keysBefore
		copy(words[off:], l.data)
		m.levels[i] = level{
			data:       words[off : off+len(l.data)],
			bits:       blocks * blockBits,
			keysBefore: l.keysBefore,
			seed:       l.seed,
		}
		off += len(l.data)
	}

	return m
}

// Load aliases data, which must stay valid and unmodified; unaligned data is
// copied instead.
func Load(data []byte) (*MPHF, error) {
	if len(data) < headerWords*8 || len(data)%8 != 0 {
		return nil, ErrCorrupt
	}
	words := asWords(data)
	n, seed, nl := words[0], words[1], words[2]
	if nl > math.MaxInt32 || uint64(len(words)) < headerWords+levelHeaderWords*nl {
		return nil, ErrCorrupt
	}

	m := &MPHF{words: words, keyCount: n, seed: seed, levels: make([]level, nl)}
	off := headerWords + levelHeaderWords*nl
	var prev uint64
	for i := range nl {
		blocks := words[headerWords+levelHeaderWords*i]
		before := words[headerWords+levelHeaderWords*i+1]
		if blocks == 0 || blocks > (uint64(len(words))-off)/blockWords || before < prev || before > n {
			return nil, ErrCorrupt
		}
		m.levels[i] = level{
			data:       words[off : off+blocks*blockWords],
			bits:       blocks * blockBits,
			keysBefore: before,
			seed:       levelSeed(seed, int(i)),
		}
		off += blocks * blockWords
		prev = before
	}
	if off != uint64(len(words)) {
		return nil, ErrCorrupt
	}

	return m, nil
}

// Find returns an arbitrary slot for a key outside the built set; callers
// must verify the hit themselves.
func (m *MPHF) Find(key []byte) (uint64, bool) {
	for i := range m.levels {
		l := &m.levels[i]
		pos := position(hash64(key, l.seed), l.bits)
		blk := pos / blockBits
		off := pos - blk*blockBits
		if l.data[blk*blockWords+1+(off>>6)]&(uint64(1)<<(off&63)) == 0 {
			continue
		}
		return l.keysBefore + rank(l.data, pos), true
	}

	return 0, false
}

func (m *MPHF) KeyCount() uint64 {
	return m.keyCount
}

func (m *MPHF) Levels() int {
	return len(m.levels)
}

func (m *MPHF) Size() int {
	return len(m.words) * 8
}

func (m *MPHF) BitsPerKey() float64 {
	if m.keyCount == 0 {
		return 0
	}
	return float64(m.Size()*8) / float64(m.keyCount)
}

// Bytes shares memory with m and must not be modified.
func (m *MPHF) Bytes() []byte {
	if nativeLittleEndian {
		return unsafe.Slice((*byte)(unsafe.Pointer(unsafe.SliceData(m.words))), len(m.words)*8)
	}
	b := make([]byte, len(m.words)*8)
	for i, w := range m.words {
		binary.LittleEndian.PutUint64(b[i*8:], w)
	}
	return b
}

func keyAt(keys []byte, keySize int, idx uint32) []byte {
	start := int(idx) * keySize
	return keys[start : start+keySize]
}

func levelSeed(seed uint64, lvl int) uint64 {
	return mix64(seed ^ (uint64(lvl)+1)*golden)
}

// position maps h onto [0, n) by a high multiply instead of a modulo.
func position(h, n uint64) uint64 {
	hi, _ := bits.Mul64(h, n)
	return hi
}

// hash64 assumes keys are already uniform, such as digests, so mixing each
// word with the seed suffices.
func hash64(key []byte, seed uint64) uint64 {
	h := seed ^ uint64(len(key))*golden
	for len(key) >= 8 {
		h = mix64(h ^ binary.LittleEndian.Uint64(key))
		key = key[8:]
	}
	if len(key) > 0 {
		var tail [8]byte
		copy(tail[:], key)
		h = mix64(h ^ binary.LittleEndian.Uint64(tail[:]))
	}
	return h
}

// mix64 is the MurmurHash3 64-bit finalizer.
func mix64(x uint64) uint64 {
	x ^= x >> 33
	x *= 0xff51afd7ed558ccd
	x ^= x >> 33
	x *= 0xc4ceb9fe1a85ec53
	x ^= x >> 33
	return x
}

var nativeLittleEndian = func() bool {
	var x uint16 = 1
	return *(*byte)(unsafe.Pointer(&x)) == 1
}()

func asWords(b []byte) []uint64 {
	if nativeLittleEndian && uintptr(unsafe.Pointer(unsafe.SliceData(b)))%8 == 0 {
		return unsafe.Slice((*uint64)(unsafe.Pointer(unsafe.SliceData(b))), len(b)/8)
	}
	w := make([]uint64, len(b)/8)
	for i := range w {
		w[i] = binary.LittleEndian.Uint64(b[i*8:])
	}
	return w
}

// rank counts the set bits before pos.
func rank(data []uint64, pos uint64) uint64 {
	blk := pos / blockBits
	off := pos - blk*blockBits
	wi := off >> 6
	base := blk * blockWords
	r := data[base]
	for k := range wi {
		r += uint64(bits.OnesCount64(data[base+1+k]))
	}
	return r + uint64(bits.OnesCount64(data[base+1+wi]&(uint64(1)<<(off&63)-1)))
}
