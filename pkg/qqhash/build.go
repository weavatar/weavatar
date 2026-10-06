package qqhash

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/crc32"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/errgroup"

	"github.com/weavatar/weavatar/pkg/mphf"
)

const (
	DefaultStart    = MinQq
	DefaultEnd      = MaxQq
	DefaultPartBits = 8 // 256 partitions, keyed by the first two hex digits

	enumerateChunk   = 1 << 20
	bucketBuffer     = 32 << 10 // write buffer per worker per bucket
	progressInterval = 10 * time.Second
)

type BuildOptions struct {
	Dir      string
	Types    []string                         // default all
	Start    uint64                           // inclusive, default 10000
	End      uint64                           // inclusive, default max uint32
	PartBits uint32                           // default 8
	Gamma    float64                          // default 2.0
	Workers  int                              // default CPU count
	Logf     func(format string, args ...any) // may be nil
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
	// the heap is almost all pointer-free buffers, so marking is cheap; a low
	// GC target trades a few extra cycles for about 30% less peak memory
	defer debug.SetGCPercent(debug.SetGCPercent(20))

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

func buildType(ctx context.Context, o BuildOptions, tmp, typ string, counts []uint64) error {
	keyBytes, err := keyBytesOf(typ)
	if err != nil {
		return err
	}
	parts := len(counts)
	n := o.End - o.Start + 1
	h := &header{
		keyBytes: uint32(keyBytes),
		start:    o.Start,
		end:      o.End,
		keyCount: n,
		partBits: o.PartBits,
		gamma:    o.Gamma,
		seed:     buildSeed(typ, o.Start, o.End),
		valSize:  n * 4,
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
	w := &blobWriter{f: idx, table: table, pending: make(map[int][]byte), off: h.blobBase()}
	var next, done atomic.Uint64
	g, ctx := errgroup.WithContext(ctx)
	for range o.Workers {
		g.Go(func() error {
			var pb partitionBuilder
			for {
				p := int(next.Add(1) - 1)
				if p >= parts {
					return nil
				}
				if err := ctx.Err(); err != nil {
					return err
				}

				m, vals, err := pb.build(tmp, typ, p, keyBytes, counts[p], o.Gamma, h.seed)
				if err != nil {
					return fmt.Errorf("qqhash: %s partition %d: %w", typ, p, err)
				}
				if _, err = val.WriteAt(vals, int64(table[p].slotOffset*4)); err != nil {
					return err
				}

				if err = w.put(p, m.Bytes()); err != nil {
					return err
				}
				_ = os.Remove(bucketPath(tmp, typ, p))

				o.Logf("[%s] 分区 %d/%d 完成：%d 键，%d 层，%.2f bit/键", typ, done.Add(1), parts, counts[p], m.Levels(), m.BitsPerKey())
			}
		})
	}
	if err = g.Wait(); err != nil {
		return err
	}
	if w.next != parts {
		return fmt.Errorf("qqhash: %s: %d partitions not written", typ, parts-w.next)
	}

	h.idxSize = uint64(w.off)
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

type blobWriter struct {
	mu      sync.Mutex
	f       *os.File
	table   []partition
	pending map[int][]byte
	next    int
	off     int64
}

func (w *blobWriter) put(p int, blob []byte) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.pending[p] = blob
	for {
		b, ok := w.pending[w.next]
		if !ok {
			return nil
		}
		delete(w.pending, w.next)
		if _, err := w.f.WriteAt(b, w.off); err != nil {
			return err
		}
		w.table[w.next].mphOffset = uint64(w.off)
		w.table[w.next].mphLen = uint64(len(b))
		w.table[w.next].crc = crc32.ChecksumIEEE(b)
		w.off = alignUp(w.off+int64(len(b)), blobAlign)
		w.next++
	}
}

type partitionBuilder struct {
	raw, keys, vals []byte
	seen            []uint64
}

func (b *partitionBuilder) build(tmp, typ string, p, keyBytes int, count uint64, gamma float64, seed uint64) (*mphf.MPHF, []byte, error) {
	b.raw = grow(b.raw, int(count*4))
	if err := readFull(bucketPath(tmp, typ, p), b.raw); err != nil {
		return nil, nil, err
	}

	kb := uint64(keyBytes)
	b.keys = grow(b.keys, int(count*kb))
	var email [32]byte
	var out [sha256.Size]byte
	for i := range count {
		qq := binary.LittleEndian.Uint32(b.raw[i*4:])
		copy(b.keys[i*kb:], digestOf(typ, appendEmail(email[:0], uint64(qq)), &out))
	}

	m, slots, err := mphf.Build(b.keys, keyBytes, mphf.Options{Gamma: gamma, Seed: seed + uint64(p)})
	if err != nil {
		return nil, nil, err
	}

	b.vals = grow(b.vals, int(count*4))
	b.seen = grow(b.seen, int((count+63)/64))
	clear(b.seen)
	for i := range count {
		slot := uint64(slots[i])
		if slot >= count {
			return nil, nil, fmt.Errorf("key %d maps to slot %d out of %d", i, slot, count)
		}
		w, bit := slot>>6, uint64(1)<<(slot&63)
		if b.seen[w]&bit != 0 {
			return nil, nil, fmt.Errorf("slot %d assigned twice", slot)
		}
		b.seen[w] |= bit
		copy(b.vals[slot*4:], b.raw[i*4:i*4+4])
	}

	return m, b.vals, nil
}

func grow[T any](s []T, n int) []T {
	return slices.Grow(s[:0], n)[:n]
}

// readFull fails unless the file is exactly len(buf) bytes.
func readFull(path string, buf []byte) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer func() {
		_ = f.Close()
	}()

	st, err := f.Stat()
	if err != nil {
		return err
	}
	if st.Size() != int64(len(buf)) {
		return fmt.Errorf("bucket size %d, want %d", st.Size(), len(buf))
	}
	_, err = io.ReadFull(f, buf)
	return err
}

// buildSeed derives the seed from the inputs so builds are reproducible.
func buildSeed(typ string, start, end uint64) uint64 {
	var out [sha256.Size]byte
	d := digestOf(TypeSHA256, fmt.Appendf(nil, "%s:%d:%d", typ, start, end), &out)
	return binary.LittleEndian.Uint64(d)
}
