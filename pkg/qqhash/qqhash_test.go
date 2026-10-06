package qqhash

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
)

//nolint:unparam // spelling out the range start keeps each test readable
func buildTestTables(tb testing.TB, start, end uint64, partBits uint32) string {
	tb.Helper()
	dir := tb.TempDir()
	err := Build(tb.Context(), BuildOptions{
		Dir:      dir,
		Start:    start,
		End:      end,
		PartBits: partBits,
		Workers:  4,
	})
	must.NoError(tb, err)
	return dir
}

func hashOf(typ, email string) string {
	switch typ {
	case TypeMD5:
		sum := md5.Sum([]byte(email))
		return hex.EncodeToString(sum[:])
	default:
		sum := sha256.Sum256([]byte(email))
		return hex.EncodeToString(sum[:])
	}
}

func TestBuildOpenLookup(t *testing.T) {
	const start, end = uint64(MinQq), uint64(MinQq + 120_000 - 1)
	dir := buildTestTables(t, start, end, DefaultPartBits)

	_, err := os.Stat(dir + "/tmp")
	check.True(t, os.IsNotExist(err))

	ts, err := Open(dir)
	must.NoError(t, err)
	defer func() {
		check.NoError(t, ts.Close())
	}()
	must.Len(t, ts.All(), 2)

	for _, typ := range Types {
		tb := ts.Table(typ)
		must.NotNil(t, tb, typ)
		s := tb.Stats()
		check.Equal(t, s.Type, typ)
		check.Equal(t, s.Start, start)
		check.Equal(t, s.End, end)
		check.Equal(t, s.KeyCount, end-start+1)
		check.Equal(t, s.Partitions, 256)
		check.Equal(t, s.ValSize, s.KeyCount*4)
		// too few keys per partition to assert bits/key; TestBitsPerKey does
		t.Logf("%s: idx=%d val=%d bits/key=%.2f levels=%d", typ, s.IdxSize, s.ValSize, s.BitsPerKey, s.MaxLevels)
	}

	// every QQ number in range hits under both hash types
	for qq := start; qq <= end; qq++ {
		email := fmt.Sprintf("%d@qq.com", qq)
		for _, typ := range Types {
			got, ok := ts.Lookup(hashOf(typ, email))
			must.True(t, ok, must.Msgf("%s %d", typ, qq))
			must.Equal(t, uint64(got), qq, must.Msgf("%s %d", typ, qq))
		}
	}

	got, ok := ts.Lookup(strings.ToUpper(hashOf(TypeMD5, "12345@qq.com")))
	check.True(t, ok)
	check.Equal(t, got, uint32(12345))

	// QQ numbers out of range and non-QQ mailboxes never hit
	for _, email := range []string{
		fmt.Sprintf("%d@qq.com", start-1),
		fmt.Sprintf("%d@qq.com", end+1),
		"12345@gmail.com",
		"user@example.com",
		"",
	} {
		for _, typ := range Types {
			_, ok := ts.Lookup(hashOf(typ, email))
			check.False(t, ok, check.Msgf("%s %q", typ, email))
		}
	}
	for i := range 2000 {
		_, ok := ts.Lookup(hashOf(TypeMD5, fmt.Sprintf("user%d@example.com", i)))
		check.False(t, ok)
	}

	for _, hash := range []string{"", "abc", strings.Repeat("g", 32), strings.Repeat("0", 33), strings.Repeat("0", 64) + "0"} {
		_, ok := ts.Lookup(hash)
		check.False(t, ok, hash)
	}

	for _, tb := range ts.All() {
		must.NoError(t, tb.Verify(t.Context(), VerifyOptions{Workers: 4, Coverage: true, Logf: t.Logf}))
		must.NoError(t, tb.Sample(t.Context(), 1000, 1))
	}
}

func TestReproducible(t *testing.T) {
	a := buildTestTables(t, MinQq, MinQq+199_999, DefaultPartBits)
	b := buildTestTables(t, MinQq, MinQq+199_999, DefaultPartBits)
	for _, typ := range Types {
		idxA, valA := fileNames(a, typ)
		idxB, valB := fileNames(b, typ)
		for _, pair := range [][2]string{{idxA, idxB}, {valA, valB}} {
			x, err := os.ReadFile(pair[0])
			must.NoError(t, err)
			y, err := os.ReadFile(pair[1])
			must.NoError(t, err)
			check.True(t, bytes.Equal(x, y), check.Msgf("%s differs from %s", pair[0], pair[1]))
		}
	}
}

func TestBitsPerKey(t *testing.T) {
	// two partitions of 50k keys each come close to production amortization
	dir := buildTestTables(t, MinQq, MinQq+99_999, 1)
	ts, err := Open(dir)
	must.NoError(t, err)
	defer func() {
		check.NoError(t, ts.Close())
	}()

	for _, tb := range ts.All() {
		s := tb.Stats()
		check.Equal(t, s.Partitions, 2)
		check.Less(t, s.BitsPerKey, 4.5, tb.Type())
		check.Greater(t, s.BitsPerKey, 3.0, tb.Type())
	}
}

func TestSparsePartitions(t *testing.T) {
	// far fewer keys than partitions leaves most partitions empty
	const start, end = uint64(MinQq), uint64(MinQq + 49)
	dir := buildTestTables(t, start, end, 10)

	ts, err := Open(dir)
	must.NoError(t, err)
	defer func() {
		check.NoError(t, ts.Close())
	}()

	for _, tb := range ts.All() {
		check.Equal(t, tb.Stats().Partitions, 1024)
		must.NoError(t, tb.Verify(t.Context(), VerifyOptions{Coverage: true}))
	}
	for qq := start; qq <= end; qq++ {
		for _, typ := range Types {
			got, ok := ts.Lookup(hashOf(typ, fmt.Sprintf("%d@qq.com", qq)))
			must.True(t, ok)
			check.Equal(t, uint64(got), qq)
		}
	}
	_, ok := ts.Lookup(hashOf(TypeMD5, "10050@qq.com"))
	check.False(t, ok)
}

func TestSingleType(t *testing.T) {
	dir := t.TempDir()
	must.NoError(t, Build(t.Context(), BuildOptions{Dir: dir, Types: []string{TypeSHA256}, Start: MinQq, End: MinQq + 999, Workers: 2}))

	ts, err := Open(dir)
	must.NoError(t, err)
	defer func() {
		check.NoError(t, ts.Close())
	}()

	check.Nil(t, ts.Table(TypeMD5))
	must.NotNil(t, ts.Table(TypeSHA256))
	_, ok := ts.Lookup(hashOf(TypeMD5, "10000@qq.com"))
	check.False(t, ok)
	got, ok := ts.Lookup(hashOf(TypeSHA256, "10000@qq.com"))
	check.True(t, ok)
	check.Equal(t, got, uint32(10000))
}

func TestOpenMissing(t *testing.T) {
	ts, err := Open(t.TempDir())
	must.NoError(t, err)
	check.Empty(t, ts.All())
	_, ok := ts.Lookup(hashOf(TypeMD5, "10000@qq.com"))
	check.False(t, ok)
	check.NoError(t, ts.Close())

	var nilTables *Tables
	_, ok = nilTables.Lookup(hashOf(TypeMD5, "10000@qq.com"))
	check.False(t, ok)
	check.Nil(t, nilTables.All())
	check.NoError(t, nilTables.Close())
}

func TestOpenCorrupt(t *testing.T) {
	dir := buildTestTables(t, MinQq, MinQq+999, 4)
	idxPath, valPath := fileNames(dir, TypeMD5)

	idx, err := os.ReadFile(idxPath)
	must.NoError(t, err)
	_, parts, err := decodeIndexHead(idx)
	must.NoError(t, err)
	blob := parts[0]

	// structural damage fails Open
	corrupt := func(name string, mutate func([]byte) []byte) {
		t.Run(name, func(t *testing.T) {
			must.NoError(t, os.WriteFile(idxPath, mutate(append([]byte{}, idx...)), 0644))
			_, err := Open(dir)
			must.ErrorIs(t, err, ErrCorrupt)
		})
	}
	corrupt("key size", func(b []byte) []byte { b[offKeyBytes] ^= 0xff; return b })
	corrupt("header", func(b []byte) []byte { b[offKeyCount] ^= 0x01; return b })
	corrupt("table", func(b []byte) []byte { b[headerSize+8] ^= 0x01; return b })
	corrupt("truncated", func(b []byte) []byte { return b[:len(b)-64] })
	corrupt("blob key count", func(b []byte) []byte { b[blob.mphOffset] ^= 0xff; return b })

	// damaged partition data is only caught by Verify's checksum; Open skips it to start fast
	t.Run("blob data", func(t *testing.T) {
		b := append([]byte{}, idx...)
		b[blob.mphOffset+blob.mphLen-1] ^= 0xff
		must.NoError(t, os.WriteFile(idxPath, b, 0644))
		ts, err := Open(dir)
		must.NoError(t, err)
		defer func() {
			check.NoError(t, ts.Close())
		}()
		must.ErrorIs(t, ts.Table(TypeMD5).Verify(t.Context(), VerifyOptions{}), ErrCorrupt)
	})

	// restore idx so only the truncated val is wrong
	must.NoError(t, os.WriteFile(idxPath, idx, 0644))
	must.NoError(t, os.Truncate(valPath, 100))
	_, err = Open(dir)
	must.ErrorIs(t, err, ErrCorrupt)
}

func TestVerifyDetectsBadValue(t *testing.T) {
	dir := buildTestTables(t, MinQq, MinQq+9_999, 6)
	_, valPath := fileNames(dir, TypeMD5)

	val, err := os.ReadFile(valPath)
	must.NoError(t, err)
	val[4*17] ^= 0x01
	must.NoError(t, os.WriteFile(valPath, val, 0644))

	ts, err := Open(dir)
	must.NoError(t, err)
	defer func() {
		check.NoError(t, ts.Close())
	}()

	err = ts.Table(TypeMD5).Verify(t.Context(), VerifyOptions{Coverage: true})
	must.ErrorIs(t, err, ErrCorrupt)
	must.NoError(t, ts.Table(TypeSHA256).Verify(t.Context(), VerifyOptions{Coverage: true}))
}

func TestBuildOptionsDefaults(t *testing.T) {
	o, err := BuildOptions{Dir: t.TempDir()}.normalize()
	must.NoError(t, err)
	check.Equal(t, o.Start, uint64(MinQq))
	check.Equal(t, o.End, uint64(MaxQq))
	check.DeepEqual(t, o.Types, Types)
	check.Equal(t, o.PartBits, uint32(DefaultPartBits))
}

func TestBuildOptions(t *testing.T) {
	dir := t.TempDir()
	cases := map[string]BuildOptions{
		"no dir":      {Start: MinQq, End: MinQq + 10},
		"bad type":    {Dir: dir, Types: []string{"sha1"}, Start: MinQq, End: MinQq + 10},
		"dup type":    {Dir: dir, Types: []string{TypeMD5, TypeMD5}, Start: MinQq, End: MinQq + 10},
		"low start":   {Dir: dir, Start: 1, End: MinQq + 10},
		"end<start":   {Dir: dir, Start: MinQq + 10, End: MinQq},
		"end too big": {Dir: dir, Start: MinQq, End: MaxQq + 1},
		"part bits":   {Dir: dir, Start: MinQq, End: MinQq + 10, PartBits: 17},
	}
	for name, o := range cases {
		t.Run(name, func(t *testing.T) {
			check.Error(t, Build(t.Context(), o))
		})
	}
}

func BenchmarkLookup(b *testing.B) {
	const start, end = uint64(MinQq), uint64(MinQq + 200_000 - 1)
	dir := buildTestTables(b, start, end, DefaultPartBits)
	ts, err := Open(dir)
	must.NoError(b, err)
	defer func() {
		_ = ts.Close()
	}()

	for _, typ := range Types {
		hits := make([]string, 1024)
		misses := make([]string, 1024)
		for i := range hits {
			hits[i] = hashOf(typ, fmt.Sprintf("%d@qq.com", start+uint64(i)*97))
			misses[i] = hashOf(typ, fmt.Sprintf("user%d@example.com", i))
		}

		b.Run(typ+"/hit", func(b *testing.B) {
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				if _, ok := ts.Lookup(hits[i&1023]); !ok {
					b.Fatal("miss")
				}
				i++
			}
		})
		b.Run(typ+"/miss", func(b *testing.B) {
			b.ReportAllocs()
			i := 0
			for b.Loop() {
				if _, ok := ts.Lookup(misses[i&1023]); ok {
					b.Fatal("hit")
				}
				i++
			}
		})
	}
}
