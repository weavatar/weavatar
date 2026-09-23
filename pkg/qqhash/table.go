package qqhash

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"hash/crc32"
	"io/fs"
	"math/rand/v2"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	"golang.org/x/sync/errgroup"

	"github.com/weavatar/weavatar/pkg/mphf"
)

// Table 是一种哈希类型的映射表，通过 mmap 加载，可并发查询。
type Table struct {
	typ      string
	keyBytes int
	hdr      header
	raw      []partition
	parts    []part
	values   []uint32 // val 文件的零拷贝视图
	idxData  []byte
	valData  []byte
}

type part struct {
	mph        *mphf.MPHF
	keyCount   uint64
	slotOffset uint64
}

// Stats 是表的统计信息。
type Stats struct {
	Type       string
	KeyBytes   int
	Start      uint64
	End        uint64
	KeyCount   uint64
	Partitions int
	IdxSize    uint64
	ValSize    uint64
	BitsPerKey float64 // 仅 MPHF 部分
	MaxLevels  int
	BuildTime  time.Time
}

// OpenTable 打开 dir 下指定哈希类型的表。
func OpenTable(dir, typ string) (*Table, error) {
	keyBytes, err := keyBytesOf(typ)
	if err != nil {
		return nil, err
	}
	if !nativeLittleEndian {
		return nil, ErrBigEndian
	}

	idxPath, valPath := fileNames(dir, typ)
	t := &Table{typ: typ, keyBytes: keyBytes}
	if t.idxData, err = mapFile(idxPath); err != nil {
		return nil, err
	}
	success := false
	defer func() {
		if !success {
			_ = t.Close()
		}
	}()

	h, raw, err := decodeIndexHead(t.idxData)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", idxPath, err)
	}
	if int(h.keyBytes) != keyBytes {
		return nil, fmt.Errorf("%s: %w: key size %d does not match %s", idxPath, ErrCorrupt, h.keyBytes, typ)
	}
	if uint64(len(t.idxData)) != h.idxSize {
		return nil, fmt.Errorf("%s: %w: file size %d, want %d", idxPath, ErrCorrupt, len(t.idxData), h.idxSize)
	}
	t.hdr = *h
	t.raw = raw
	t.parts = make([]part, len(raw))
	for i, p := range raw {
		m, err := mphf.Load(t.idxData[p.mphOffset : p.mphOffset+p.mphLen])
		if err != nil {
			return nil, fmt.Errorf("%s: %w: partition %d: %w", idxPath, ErrCorrupt, i, err)
		}
		if m.KeyCount() != p.keyCount {
			return nil, fmt.Errorf("%s: %w: partition %d has %d keys, want %d", idxPath, ErrCorrupt, i, m.KeyCount(), p.keyCount)
		}
		t.parts[i] = part{mph: m, keyCount: p.keyCount, slotOffset: p.slotOffset}
	}

	if t.valData, err = mapFile(valPath); err != nil {
		return nil, err
	}
	if uint64(len(t.valData)) != h.valSize {
		return nil, fmt.Errorf("%s: %w: file size %d, want %d", valPath, ErrCorrupt, len(t.valData), h.valSize)
	}
	t.values = unsafe.Slice((*uint32)(unsafe.Pointer(unsafe.SliceData(t.valData))), h.keyCount)

	adviseWillNeed(t.idxData)
	adviseRandom(t.valData)
	success = true
	return t, nil
}

// Close 释放映射的内存，之后不得再查询。
func (t *Table) Close() error {
	var errs []error
	if t.idxData != nil {
		errs = append(errs, unmapFile(t.idxData))
		t.idxData = nil
	}
	if t.valData != nil {
		errs = append(errs, unmapFile(t.valData))
		t.valData = nil
	}
	t.parts, t.values = nil, nil
	return errors.Join(errs...)
}

// Type 返回哈希类型。
func (t *Table) Type() string {
	return t.typ
}

// Lookup 通过十六进制哈希查找 QQ 号。
func (t *Table) Lookup(hash string) (uint32, bool) {
	var d [sha256.Size]byte
	if !decodeHex(d[:t.keyBytes], hash) {
		return 0, false
	}
	return t.LookupDigest(d[:t.keyBytes])
}

// LookupDigest 通过摘要字节查找 QQ 号。
func (t *Table) LookupDigest(d []byte) (uint32, bool) {
	if len(d) != t.keyBytes {
		return 0, false
	}
	p := &t.parts[partitionOf(d, t.hdr.partBits)]
	slot, ok := p.mph.Find(d)
	if !ok || slot >= p.keyCount {
		return 0, false
	}

	// MPHF 对表外的键也会给出槽位，必须用槽位里的 QQ 号重新计算摘要校验
	qq := t.values[p.slotOffset+slot]
	var out [sha256.Size]byte
	if !bytes.Equal(digestQq(t.typ, uint64(qq), &out), d) {
		return 0, false
	}
	return qq, true
}

// Stats 返回统计信息。
func (t *Table) Stats() Stats {
	s := Stats{
		Type:       t.typ,
		KeyBytes:   t.keyBytes,
		Start:      t.hdr.start,
		End:        t.hdr.end,
		KeyCount:   t.hdr.keyCount,
		Partitions: len(t.parts),
		IdxSize:    t.hdr.idxSize,
		ValSize:    t.hdr.valSize,
		BuildTime:  time.Unix(t.hdr.buildTime, 0),
	}
	var mphBytes uint64
	for _, p := range t.parts {
		mphBytes += uint64(p.mph.Size())
		s.MaxLevels = max(s.MaxLevels, p.mph.Levels())
	}
	if s.KeyCount > 0 {
		s.BitsPerKey = float64(mphBytes*8) / float64(s.KeyCount)
	}
	return s
}

// FormatSize 把字节数格式化为人类可读的大小。
func FormatSize(n uint64) string {
	const unit = 1024
	if n < unit {
		return fmt.Sprintf("%d B", n)
	}
	div, exp := uint64(unit), 0
	for m := n / unit; m >= unit; m /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.2f %ciB", float64(n)/float64(div), "KMGTPE"[exp])
}

// VerifyOptions 是校验参数。
type VerifyOptions struct {
	Workers int
	// Coverage 额外检查每个 QQ 号恰好出现一次，需要 (End-Start+1)/8 字节内存。
	Coverage bool
	// Logf 输出进度，可为 nil。
	Logf func(format string, args ...any)
}

const (
	verifyChunk     = 1 << 20
	verifyMaxErrors = 20
)

// Verify 全量校验：先核对各分区 MPHF 的校验和，再按槽位顺序扫描值数组，
// 每个槽位的 QQ 号重新计算摘要后必须映射回同一槽位。
func (t *Table) Verify(ctx context.Context, o VerifyOptions) error {
	if o.Workers <= 0 {
		o.Workers = runtime.NumCPU()
	}
	if o.Logf == nil {
		o.Logf = func(string, ...any) {}
	}

	for i, p := range t.raw {
		if crc32.ChecksumIEEE(t.idxData[p.mphOffset:p.mphOffset+p.mphLen]) != p.crc {
			return fmt.Errorf("%w: partition %d checksum mismatch", ErrCorrupt, i)
		}
	}
	o.Logf("[%s] 分区校验和通过", t.typ)

	n := t.hdr.keyCount
	var bitmap []uint64
	if o.Coverage {
		bitmap = make([]uint64, (n+63)/64)
	}

	var mu sync.Mutex
	var msgs []string
	var failed atomic.Uint64
	report := func(msg string) {
		if failed.Add(1) <= verifyMaxErrors {
			mu.Lock()
			msgs = append(msgs, msg)
			mu.Unlock()
		}
	}

	chunks := (n + verifyChunk - 1) / verifyChunk
	var next, done atomic.Uint64
	g, ctx := errgroup.WithContext(ctx)
	for range o.Workers {
		g.Go(func() error {
			var out [sha256.Size]byte
			for {
				c := next.Add(1) - 1
				if c >= chunks {
					return nil
				}
				if err := ctx.Err(); err != nil {
					return err
				}

				from := c * verifyChunk
				to := min(from+verifyChunk, n)
				for slot := from; slot < to; slot++ {
					qq := uint64(t.values[slot])
					if qq < t.hdr.start || qq > t.hdr.end {
						report(fmt.Sprintf("slot %d: qq %d out of range", slot, qq))
						continue
					}
					d := digestQq(t.typ, qq, &out)
					p := &t.parts[partitionOf(d, t.hdr.partBits)]
					if slot < p.slotOffset || slot >= p.slotOffset+p.keyCount {
						report(fmt.Sprintf("slot %d: qq %d belongs to another partition", slot, qq))
						continue
					}
					if s, ok := p.mph.Find(d); !ok || p.slotOffset+s != slot {
						report(fmt.Sprintf("slot %d: qq %d maps to slot %d", slot, qq, p.slotOffset+s))
						continue
					}
					if bitmap != nil {
						bit := qq - t.hdr.start
						mask := uint64(1) << (bit & 63)
						if atomic.OrUint64(&bitmap[bit>>6], mask)&mask != 0 {
							report(fmt.Sprintf("slot %d: qq %d appears twice", slot, qq))
						}
					}
				}
				if d := done.Add(to - from); d == n || (d/verifyChunk)%64 == 0 {
					o.Logf("[%s] 校验进度 %.1f%%", t.typ, float64(d)*100/float64(n))
				}
			}
		})
	}
	if err := g.Wait(); err != nil {
		return err
	}

	// 槽位数等于键数、每个值都在范围内且互不重复，就说明每个 QQ 号恰好出现一次
	if f := failed.Load(); f > 0 {
		return fmt.Errorf("%w: %d slots failed verification, first %d: %v", ErrCorrupt, f, len(msgs), msgs)
	}
	return nil
}

// Sample 随机抽取 count 个 QQ 号，走完整的十六进制查询路径校验。
func (t *Table) Sample(ctx context.Context, count int, seed uint64) error {
	r := rand.New(rand.NewPCG(seed, seed^0x5DEECE66D))
	var out [sha256.Size]byte
	for i := range count {
		if err := ctx.Err(); err != nil {
			return err
		}
		qq := t.hdr.start + r.Uint64N(t.hdr.keyCount)
		hash := hex.EncodeToString(digestQq(t.typ, qq, &out))
		got, ok := t.Lookup(hash)
		if !ok || uint64(got) != qq {
			return fmt.Errorf("%w: sample %d: qq %d hash %s returned %d, %v", ErrCorrupt, i, qq, hash, got, ok)
		}
	}
	return nil
}

// Tables 汇总各哈希类型的表，按哈希长度分发查询。
type Tables struct {
	tables map[string]*Table
}

// Open 打开 dir 下存在的表，缺失的类型跳过。
func Open(dir string) (*Tables, error) {
	ts := &Tables{tables: make(map[string]*Table, len(Types))}
	for _, typ := range Types {
		idxPath, _ := fileNames(dir, typ)
		if _, err := os.Stat(idxPath); err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				continue
			}
			_ = ts.Close()
			return nil, err
		}
		t, err := OpenTable(dir, typ)
		if err != nil {
			_ = ts.Close()
			return nil, err
		}
		ts.tables[typ] = t
	}
	return ts, nil
}

// Lookup 通过十六进制哈希查找 QQ 号，按长度自动识别 MD5 或 SHA256。
func (ts *Tables) Lookup(hash string) (uint32, bool) {
	if ts == nil {
		return 0, false
	}
	typ, ok := typeOfHashLen(len(hash))
	if !ok {
		return 0, false
	}
	t := ts.tables[typ]
	if t == nil {
		return 0, false
	}
	return t.Lookup(hash)
}

// Table 返回指定类型的表，未加载时返回 nil。
func (ts *Tables) Table(typ string) *Table {
	if ts == nil {
		return nil
	}
	return ts.tables[typ]
}

// All 按 Types 的顺序返回已加载的表。
func (ts *Tables) All() []*Table {
	if ts == nil {
		return nil
	}
	all := make([]*Table, 0, len(ts.tables))
	for _, typ := range Types {
		if t := ts.tables[typ]; t != nil {
			all = append(all, t)
		}
	}
	return all
}

// Close 关闭全部表。
func (ts *Tables) Close() error {
	if ts == nil {
		return nil
	}
	var errs []error
	for typ, t := range ts.tables {
		errs = append(errs, t.Close())
		delete(ts.tables, typ)
	}
	return errors.Join(errs...)
}
