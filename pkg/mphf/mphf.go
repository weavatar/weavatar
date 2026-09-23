// Package mphf 实现 BBHash 风格的最小完美哈希函数
//
// 序列化布局（小端 uint64 字）：
//
//	[0]        magic
//	[1]        键数量 n
//	[2]        种子
//	[3]        层数 L
//	[4, 4+2L)  每层 { 块数, 该层之前已放置的键数 }
//	[4+2L, …)  每层的块数据，每块 8 个字：首字是块前累计 rank，其余 7 个字是位向量
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

	headerWords      = 4
	levelHeaderWords = 2

	DefaultGamma     = 2.0
	DefaultMaxLevels = 32

	maxAttempts = 16
	golden      = 0x9E3779B97F4A7C15
)

var magic = binary.LittleEndian.Uint64([]byte("WAMPHF01"))

var (
	ErrDuplicateKey = errors.New("mphf: duplicate key")
	ErrTooManyKeys  = errors.New("mphf: too many keys")
	ErrBuildFailed  = errors.New("mphf: build failed after retries")
	ErrCorrupt      = errors.New("mphf: corrupt data")
)

type Options struct {
	Gamma     float64 // 每键位数，默认 2.0，越大查询越快、体积越大
	MaxLevels int     // 默认 32，超过后换种子重试
	Seed      uint64
}

type MPHF struct {
	words    []uint64 // levels 的切片都指向这里
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

func Build(keys []byte, keySize int, opts Options) (*MPHF, error) {
	if keySize <= 0 || len(keys)%keySize != 0 {
		return nil, fmt.Errorf("mphf: invalid key size %d for %d bytes", keySize, len(keys))
	}
	n := len(keys) / keySize
	if uint64(n) >= math.MaxUint32 {
		return nil, ErrTooManyKeys
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

	for attempt := range uint64(maxAttempts) {
		seed := mix64(opts.Seed*golden + (attempt+1)*0xD6E8FEB86659FD93)
		levels, leftover := buildLevels(keys, keySize, n, gamma, maxLevels, seed)
		if len(leftover) == 0 {
			return assemble(uint64(n), seed, levels), nil
		}
		// 相同的键永远分不开，先排除再换种子重试
		if hasDuplicate(keys, keySize, leftover) {
			return nil, ErrDuplicateKey
		}
	}

	return nil, ErrBuildFailed
}

type builtLevel struct {
	data       []uint64
	keysBefore uint64
	seed       uint64
}

func buildLevels(keys []byte, keySize, n int, gamma float64, maxLevels int, seed uint64) ([]builtLevel, []uint32) {
	cur := make([]uint32, n)
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

		// seen：被命中的位；coll：被命中两次以上的位
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

		// 撞车的键进入下一层，原地复用 cur
		next := cur[:0]
		for _, idx := range cur {
			pos := position(hash64(keyAt(keys, keySize, idx), lseed), nbits)
			if coll[pos>>6]&(uint64(1)<<(pos&63)) != 0 {
				next = append(next, idx)
			}
		}

		// 只命中一次的位即本层放置的键，按块交错写入并累计 rank
		data := make([]uint64, blocks*blockWords)
		var rank uint64
		for b := range blocks {
			data[b*blockWords] = rank
			for k := range uint64(blockWords - 1) {
				w := seen[b*(blockWords-1)+k] &^ coll[b*(blockWords-1)+k]
				data[b*blockWords+1+k] = w
				rank += uint64(bits.OnesCount64(w))
			}
		}

		levels = append(levels, builtLevel{data: data, keysBefore: keysBefore, seed: lseed})
		keysBefore += rank
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
	words[0] = magic
	words[1] = n
	words[2] = seed
	words[3] = uint64(len(built))

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

// Load 返回值引用 data，data 需保持有效且不可修改；不对齐时会拷贝
func Load(data []byte) (*MPHF, error) {
	if len(data) < headerWords*8 || len(data)%8 != 0 {
		return nil, ErrCorrupt
	}
	words := asWords(data)
	if words[0] != magic {
		return nil, ErrCorrupt
	}

	n, seed, nl := words[1], words[2], words[3]
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

// Find key 不在原始键集合中时结果不可信，调用方需自行校验
func (m *MPHF) Find(key []byte) (uint64, bool) {
	for i := range m.levels {
		l := &m.levels[i]
		pos := position(hash64(key, l.seed), l.bits)
		blk := pos / blockBits
		off := pos - blk*blockBits
		wi := off >> 6
		base := blk * blockWords
		word := l.data[base+1+wi]
		mask := uint64(1) << (off & 63)
		if word&mask == 0 {
			continue
		}

		rank := l.data[base]
		for k := range wi {
			rank += uint64(bits.OnesCount64(l.data[base+1+k]))
		}
		rank += uint64(bits.OnesCount64(word & (mask - 1)))
		return l.keysBefore + rank, true
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

// Bytes 与内部存储共享内存，不得修改
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

// 乘法取高位代替取模
func position(h, n uint64) uint64 {
	hi, _ := bits.Mul64(h, n)
	return hi
}

// key 应当已经均匀分布（如摘要），逐字与种子混合即可
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

// MurmurHash3 的 64 位终结函数
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
