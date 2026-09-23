package qqhash

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/weavatar/weavatar/pkg/mphf"
)

const (
	// DefaultStart 是默认的起始 QQ 号。
	DefaultStart = MinQq
	// DefaultEnd 是默认的结束 QQ 号。
	DefaultEnd = 4_000_000_000
	// DefaultPartBits 是默认的分区位数，256 个分区正好对应哈希的前两位十六进制。
	DefaultPartBits = 8

	enumerateChunk   = 1 << 20
	bucketBuffer     = 32 << 10 // 每个 worker 对每个桶的写缓冲
	progressInterval = 10 * time.Second
)

// BuildOptions 是构建参数。
type BuildOptions struct {
	Dir      string   // 输出目录
	Types    []string // 要构建的哈希类型，默认 md5 和 sha256
	Start    uint64   // 起始 QQ 号（含），默认 10000
	End      uint64   // 结束 QQ 号（含），默认 4000000000
	PartBits uint32   // 分区位数，分区数 = 2^PartBits，默认 8
	Gamma    float64  // MPHF 的 γ 参数，默认 2.0
	Workers  int      // 并行数，默认 CPU 数
	// Logf 输出进度，可为 nil。
	Logf func(format string, args ...any)
}

func (o BuildOptions) normalize() (BuildOptions, error) {
	if o.Dir == "" {
		return o, errors.New("qqhash: output dir is required")
	}
	if len(o.Types) == 0 {
		o.Types = slices.Clone(Types)
	}
	for i, typ := range o.Types {
		if _, err := keyBytesOf(typ); err != nil {
			return o, err
		}
		if slices.Contains(o.Types[:i], typ) {
			return o, fmt.Errorf("qqhash: duplicate hash type %q", typ)
		}
	}
	if o.Start == 0 {
		o.Start = DefaultStart
	}
	if o.End == 0 {
		o.End = DefaultEnd
	}
	if o.Start < MinQq {
		return o, fmt.Errorf("qqhash: start %d is less than %d", o.Start, MinQq)
	}
	if o.End < o.Start {
		return o, fmt.Errorf("qqhash: end %d is less than start %d", o.End, o.Start)
	}
	if o.End > MaxQq {
		return o, fmt.Errorf("qqhash: end %d exceeds %d", o.End, uint64(MaxQq))
	}
	if o.PartBits == 0 {
		o.PartBits = DefaultPartBits
	}
	if o.PartBits > maxPartBits {
		return o, fmt.Errorf("qqhash: partition bits %d exceeds %d", o.PartBits, maxPartBits)
	}
	if o.Gamma <= 0 {
		o.Gamma = mphf.DefaultGamma
	}
	if o.Workers <= 0 {
		o.Workers = runtime.NumCPU()
	}
	if o.Logf == nil {
		o.Logf = func(string, ...any) {}
	}
	return o, nil
}

// Build 构建全部哈希类型的映射表并写入 o.Dir。
//
// 先把所有 QQ 号按分区写入桶文件，再逐分区重算摘要、构建 MPHF 并写入值数组，
// 最后原子替换目标文件。构建期间的中间文件位于 o.Dir/tmp。
func Build(ctx context.Context, o BuildOptions) error {
	o, err := o.normalize()
	if err != nil {
		return err
	}

	tmp := filepath.Join(o.Dir, "tmp")
	if err = os.MkdirAll(tmp, 0755); err != nil {
		return err
	}
	defer func() {
		_ = os.RemoveAll(tmp)
	}()

	o.Logf("QQ 号范围 %d ~ %d（%d 个），哈希类型 %v，分区 %d，并行 %d", o.Start, o.End, o.End-o.Start+1, o.Types, 1<<o.PartBits, o.Workers)

	start := time.Now()
	counts, err := enumerate(ctx, o, tmp)
	if err != nil {
		return err
	}
	o.Logf("枚举完成，耗时 %s", time.Since(start).Round(time.Second))

	for _, typ := range o.Types {
		if err = buildType(ctx, o, tmp, typ, counts[typ]); err != nil {
			return err
		}
	}

	return nil
}

type bucket struct {
	mu sync.Mutex
	f  *os.File
}

func (b *bucket) write(p []byte) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	_, err := b.f.Write(p)
	return err
}

func bucketPath(tmp, typ string, p int) string {
	return filepath.Join(tmp, fmt.Sprintf("%s_%d.bin", typ, p))
}

// enumerate 枚举全部 QQ 号，按分区把 QQ 号写入桶文件，返回每种类型每个分区的键数。
func enumerate(ctx context.Context, o BuildOptions, tmp string) (map[string][]uint64, error) {
	parts := 1 << o.PartBits
	buckets := make(map[string][]*bucket, len(o.Types))
	closeAll := func() error {
		var errs []error
		for _, bs := range buckets {
			for _, b := range bs {
				if b != nil {
					errs = append(errs, b.f.Close())
				}
			}
		}
		return errors.Join(errs...)
	}
	for _, typ := range o.Types {
		buckets[typ] = make([]*bucket, parts)
		for p := range parts {
			f, err := os.OpenFile(bucketPath(tmp, typ, p), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0644)
			if err != nil {
				_ = closeAll()
				return nil, err
			}
			buckets[typ][p] = &bucket{f: f}
		}
	}

	n := o.End - o.Start + 1
	chunks := (n + enumerateChunk - 1) / enumerateChunk
	var next, done atomic.Uint64
	g, ctx := errgroup.WithContext(ctx)
	for range o.Workers {
		g.Go(func() error {
			bufs := make([][][]byte, len(o.Types))
			for ti := range bufs {
				bufs[ti] = make([][]byte, parts)
				for p := range bufs[ti] {
					bufs[ti][p] = make([]byte, 0, bucketBuffer)
				}
			}

			var email [32]byte
			var out [sha256.Size]byte
			for {
				c := next.Add(1) - 1
				if c >= chunks {
					break
				}
				if err := ctx.Err(); err != nil {
					return err
				}

				from := o.Start + c*enumerateChunk
				to := min(from+enumerateChunk-1, o.End)
				for qq := from; qq <= to; qq++ {
					e := appendEmail(email[:0], qq)
					for ti, typ := range o.Types {
						d := digestOf(typ, e, &out)
						p := partitionOf(d, o.PartBits)
						buf := binary.LittleEndian.AppendUint32(bufs[ti][p], uint32(qq))
						if len(buf) == cap(buf) {
							if err := buckets[typ][p].write(buf); err != nil {
								return err
							}
							buf = buf[:0]
						}
						bufs[ti][p] = buf
					}
				}
				done.Add(to - from + 1)
			}

			for ti, typ := range o.Types {
				for p, buf := range bufs[ti] {
					if len(buf) > 0 {
						if err := buckets[typ][p].write(buf); err != nil {
							return err
						}
					}
				}
			}
			return nil
		})
	}

	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(progressInterval)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				o.Logf("枚举进度 %.1f%%", float64(done.Load())*100/float64(n))
			}
		}
	}()

	err := g.Wait()
	close(stop)
	if closeErr := closeAll(); err == nil {
		err = closeErr
	}
	if err != nil {
		return nil, err
	}

	counts := make(map[string][]uint64, len(o.Types))
	for _, typ := range o.Types {
		counts[typ] = make([]uint64, parts)
		var total uint64
		for p := range parts {
			st, err := os.Stat(bucketPath(tmp, typ, p))
			if err != nil {
				return nil, err
			}
			if st.Size()%4 != 0 {
				return nil, fmt.Errorf("qqhash: bucket %s_%d has odd size %d", typ, p, st.Size())
			}
			counts[typ][p] = uint64(st.Size() / 4)
			total += counts[typ][p]
		}
		if total != n {
			return nil, fmt.Errorf("qqhash: %s buckets hold %d keys, want %d", typ, total, n)
		}
	}

	return counts, nil
}

// buildType 构建一种哈希类型的 idx 与 val 文件。
func buildType(ctx context.Context, o BuildOptions, tmp, typ string, counts []uint64) error {
	keyBytes, err := keyBytesOf(typ)
	if err != nil {
		return err
	}
	parts := len(counts)
	n := o.End - o.Start + 1
	h := &header{
		keyBytes:  uint32(keyBytes),
		start:     o.Start,
		end:       o.End,
		keyCount:  n,
		partBits:  o.PartBits,
		gamma:     o.Gamma,
		seed:      buildSeed(typ, o.Start, o.End),
		buildTime: time.Now().Unix(),
		valSize:   n * 4,
	}
	table := make([]partition, parts)
	var slots uint64
	for p := range parts {
		table[p].keyCount = counts[p]
		table[p].slotOffset = slots
		slots += counts[p]
	}

	idxPath, valPath := fileNames(o.Dir, typ)
	idxTmp, valTmp := idxPath+".tmp", valPath+".tmp"
	idx, err := os.Create(idxTmp)
	if err != nil {
		return err
	}
	val, err := os.Create(valTmp)
	if err != nil {
		_ = idx.Close()
		return err
	}
	success := false
	defer func() {
		if !success {
			_ = idx.Close()
			_ = val.Close()
			_ = os.Remove(idxTmp)
			_ = os.Remove(valTmp)
		}
	}()
	if err = val.Truncate(int64(h.valSize)); err != nil {
		return err
	}

	start := time.Now()
	var mu sync.Mutex
	blobNext := h.blobBase()
	var next, done atomic.Uint64
	g, ctx := errgroup.WithContext(ctx)
	for range o.Workers {
		g.Go(func() error {
			for {
				p := int(next.Add(1) - 1)
				if p >= parts {
					return nil
				}
				if err := ctx.Err(); err != nil {
					return err
				}

				m, vals, err := buildPartition(tmp, typ, p, keyBytes, counts[p], o.Gamma, h.seed)
				if err != nil {
					return fmt.Errorf("qqhash: %s partition %d: %w", typ, p, err)
				}
				if _, err = val.WriteAt(vals, int64(table[p].slotOffset*4)); err != nil {
					return err
				}

				blob := m.Bytes()
				mu.Lock()
				off := blobNext
				blobNext = alignUp(off+int64(len(blob)), blobAlign)
				mu.Unlock()
				if _, err = idx.WriteAt(blob, off); err != nil {
					return err
				}
				table[p].mphOffset = uint64(off)
				table[p].mphLen = uint64(len(blob))
				table[p].crc = crc32.ChecksumIEEE(blob)
				_ = os.Remove(bucketPath(tmp, typ, p))

				o.Logf("[%s] 分区 %d/%d 完成：%d 键，%d 层，%.2f bit/键", typ, done.Add(1), parts, counts[p], m.Levels(), m.BitsPerKey())
			}
		})
	}
	if err = g.Wait(); err != nil {
		return err
	}

	h.idxSize = uint64(blobNext)
	if _, err = idx.WriteAt(encodeIndexHead(h, table), 0); err != nil {
		return err
	}
	if err = idx.Truncate(int64(h.idxSize)); err != nil {
		return err
	}
	for _, f := range []*os.File{idx, val} {
		if err = f.Sync(); err != nil {
			return err
		}
		if err = f.Close(); err != nil {
			return err
		}
	}
	success = true
	if err = os.Rename(valTmp, valPath); err != nil {
		return err
	}
	if err = os.Rename(idxTmp, idxPath); err != nil {
		return err
	}

	o.Logf("[%s] 构建完成，耗时 %s：%s %s，%s %s", typ, time.Since(start).Round(time.Second), filepath.Base(idxPath), FormatSize(h.idxSize), filepath.Base(valPath), FormatSize(h.valSize))
	return nil
}

// buildPartition 读取桶文件，重算摘要并构建该分区的 MPHF，返回 MPHF 与按槽位排列的值数组。
func buildPartition(tmp, typ string, p, keyBytes int, count uint64, gamma float64, seed uint64) (*mphf.MPHF, []byte, error) {
	raw, err := os.ReadFile(bucketPath(tmp, typ, p))
	if err != nil {
		return nil, nil, err
	}
	if uint64(len(raw)) != count*4 {
		return nil, nil, fmt.Errorf("bucket size %d, want %d", len(raw), count*4)
	}

	kb := uint64(keyBytes)
	keys := make([]byte, count*kb)
	var email [32]byte
	var out [sha256.Size]byte
	for i := range count {
		qq := binary.LittleEndian.Uint32(raw[i*4:])
		copy(keys[i*kb:], digestOf(typ, appendEmail(email[:0], uint64(qq)), &out))
	}

	m, err := mphf.Build(keys, keyBytes, mphf.Options{Gamma: gamma, Seed: seed + uint64(p)})
	if err != nil {
		return nil, nil, err
	}

	vals := make([]byte, count*4)
	seen := make([]uint64, (count+63)/64)
	for i := range count {
		slot, ok := m.Find(keys[i*kb : (i+1)*kb])
		if !ok || slot >= count {
			return nil, nil, fmt.Errorf("key %d maps to slot %d out of %d", i, slot, count)
		}
		w, b := slot>>6, uint64(1)<<(slot&63)
		if seen[w]&b != 0 {
			return nil, nil, fmt.Errorf("slot %d assigned twice", slot)
		}
		seen[w] |= b
		copy(vals[slot*4:], raw[i*4:i*4+4])
	}

	return m, vals, nil
}

// buildSeed 由类型和范围派生种子，使同样的输入得到同样的输出。
func buildSeed(typ string, start, end uint64) uint64 {
	var out [sha256.Size]byte
	d := digestOf(TypeSHA256, fmt.Appendf(nil, "%s:%d:%d", typ, start, end), &out)
	return binary.LittleEndian.Uint64(d)
}
