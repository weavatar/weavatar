package qqhash

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func buildTestTables(t testing.TB, start, end uint64, partBits uint32) string {
	t.Helper()
	dir := t.TempDir()
	err := Build(t.Context(), BuildOptions{
		Dir:      dir,
		Start:    start,
		End:      end,
		PartBits: partBits,
		Workers:  4,
	})
	require.NoError(t, err)
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

	// 中间文件应当已被清理
	_, err := os.Stat(dir + "/tmp")
	assert.True(t, os.IsNotExist(err))

	ts, err := Open(dir)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, ts.Close())
	}()
	require.Len(t, ts.All(), 2)

	for _, typ := range Types {
		tb := ts.Table(typ)
		require.NotNil(t, tb, typ)
		s := tb.Stats()
		assert.Equal(t, typ, s.Type)
		assert.Equal(t, start, s.Start)
		assert.Equal(t, end, s.End)
		assert.Equal(t, end-start+1, s.KeyCount)
		assert.Equal(t, 256, s.Partitions)
		assert.Equal(t, s.KeyCount*4, s.ValSize)
		assert.False(t, s.BuildTime.IsZero())
		// 每区只有几百个键，MPHF 的固定开销摊不薄，bits/key 只做记录不做断言，见 TestBitsPerKey
		t.Logf("%s: idx=%d val=%d bits/key=%.2f levels=%d", typ, s.IdxSize, s.ValSize, s.BitsPerKey, s.MaxLevels)
	}

	// 范围内的每个 QQ 号两种哈希都必须命中
	for qq := start; qq <= end; qq++ {
		email := fmt.Sprintf("%d@qq.com", qq)
		for _, typ := range Types {
			got, ok := ts.Lookup(hashOf(typ, email))
			require.True(t, ok, "%s %d", typ, qq)
			require.Equal(t, uint64(got), qq, "%s %d", typ, qq)
		}
	}

	// 大写十六进制同样接受
	got, ok := ts.Lookup(strings.ToUpper(hashOf(TypeMD5, "12345@qq.com")))
	assert.True(t, ok)
	assert.Equal(t, uint32(12345), got)

	// 范围外的 QQ 号与非 QQ 邮箱都不能命中
	for _, email := range []string{
		fmt.Sprintf("%d@qq.com", start-1),
		fmt.Sprintf("%d@qq.com", end+1),
		"12345@gmail.com",
		"user@example.com",
		"",
	} {
		for _, typ := range Types {
			_, ok := ts.Lookup(hashOf(typ, email))
			assert.False(t, ok, "%s %q", typ, email)
		}
	}
	for i := range 2000 {
		_, ok := ts.Lookup(hashOf(TypeMD5, fmt.Sprintf("user%d@example.com", i)))
		assert.False(t, ok)
	}

	// 非法输入
	for _, hash := range []string{"", "abc", strings.Repeat("g", 32), strings.Repeat("0", 33), strings.Repeat("0", 64) + "0"} {
		_, ok := ts.Lookup(hash)
		assert.False(t, ok, hash)
	}

	for _, tb := range ts.All() {
		require.NoError(t, tb.Verify(t.Context(), VerifyOptions{Workers: 4, Coverage: true, Logf: t.Logf}))
		require.NoError(t, tb.Sample(t.Context(), 1000, 1))
	}
}

func TestBitsPerKey(t *testing.T) {
	// 两个分区各 5 万键，接近生产环境单分区的摊销水平
	dir := buildTestTables(t, MinQq, MinQq+99_999, 1)
	ts, err := Open(dir)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, ts.Close())
	}()

	for _, tb := range ts.All() {
		s := tb.Stats()
		assert.Equal(t, 2, s.Partitions)
		assert.Less(t, s.BitsPerKey, 4.5, tb.Type())
		assert.Greater(t, s.BitsPerKey, 3.0, tb.Type())
	}
}

func TestSparsePartitions(t *testing.T) {
	// 键数远少于分区数，大量分区为空
	const start, end = uint64(MinQq), uint64(MinQq + 49)
	dir := buildTestTables(t, start, end, 10)

	ts, err := Open(dir)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, ts.Close())
	}()

	for _, tb := range ts.All() {
		assert.Equal(t, 1024, tb.Stats().Partitions)
		require.NoError(t, tb.Verify(t.Context(), VerifyOptions{Coverage: true}))
	}
	for qq := start; qq <= end; qq++ {
		for _, typ := range Types {
			got, ok := ts.Lookup(hashOf(typ, fmt.Sprintf("%d@qq.com", qq)))
			require.True(t, ok)
			assert.Equal(t, uint64(got), qq)
		}
	}
	_, ok := ts.Lookup(hashOf(TypeMD5, "10050@qq.com"))
	assert.False(t, ok)
}

func TestSingleType(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, Build(t.Context(), BuildOptions{Dir: dir, Types: []string{TypeSHA256}, Start: MinQq, End: MinQq + 999, Workers: 2}))

	ts, err := Open(dir)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, ts.Close())
	}()

	assert.Nil(t, ts.Table(TypeMD5))
	require.NotNil(t, ts.Table(TypeSHA256))
	_, ok := ts.Lookup(hashOf(TypeMD5, "10000@qq.com"))
	assert.False(t, ok)
	got, ok := ts.Lookup(hashOf(TypeSHA256, "10000@qq.com"))
	assert.True(t, ok)
	assert.Equal(t, uint32(10000), got)
}

func TestOpenMissing(t *testing.T) {
	ts, err := Open(t.TempDir())
	require.NoError(t, err)
	assert.Empty(t, ts.All())
	_, ok := ts.Lookup(hashOf(TypeMD5, "10000@qq.com"))
	assert.False(t, ok)
	assert.NoError(t, ts.Close())

	var nilTables *Tables
	_, ok = nilTables.Lookup(hashOf(TypeMD5, "10000@qq.com"))
	assert.False(t, ok)
	assert.Nil(t, nilTables.All())
	assert.NoError(t, nilTables.Close())
}

func TestOpenCorrupt(t *testing.T) {
	dir := buildTestTables(t, MinQq, MinQq+999, 4)
	idxPath, valPath := fileNames(dir, TypeMD5)

	idx, err := os.ReadFile(idxPath)
	require.NoError(t, err)
	_, parts, err := decodeIndexHead(idx)
	require.NoError(t, err)
	blob := parts[0]

	// 结构性损坏在 Open 时就会发现
	corrupt := func(name string, mutate func([]byte) []byte) {
		t.Run(name, func(t *testing.T) {
			require.NoError(t, os.WriteFile(idxPath, mutate(append([]byte{}, idx...)), 0644))
			_, err := Open(dir)
			require.ErrorIs(t, err, ErrCorrupt)
		})
	}
	corrupt("magic", func(b []byte) []byte { b[0] ^= 0xff; return b })
	corrupt("header", func(b []byte) []byte { b[offKeyCount] ^= 0x01; return b })
	corrupt("table", func(b []byte) []byte { b[headerSize+8] ^= 0x01; return b })
	corrupt("truncated", func(b []byte) []byte { return b[:len(b)-64] })
	corrupt("blob magic", func(b []byte) []byte { b[blob.mphOffset] ^= 0xff; return b })

	// 分区数据损坏为了启动速度不在 Open 时检查，由 Verify 的校验和发现
	t.Run("blob data", func(t *testing.T) {
		b := append([]byte{}, idx...)
		b[blob.mphOffset+blob.mphLen-1] ^= 0xff
		require.NoError(t, os.WriteFile(idxPath, b, 0644))
		ts, err := Open(dir)
		require.NoError(t, err)
		defer func() {
			assert.NoError(t, ts.Close())
		}()
		require.ErrorIs(t, ts.Table(TypeMD5).Verify(t.Context(), VerifyOptions{}), ErrCorrupt)
	})

	// 恢复 idx，破坏 val 的长度
	require.NoError(t, os.WriteFile(idxPath, idx, 0644))
	require.NoError(t, os.Truncate(valPath, 100))
	_, err = Open(dir)
	require.ErrorIs(t, err, ErrCorrupt)
}

func TestVerifyDetectsBadValue(t *testing.T) {
	dir := buildTestTables(t, MinQq, MinQq+9_999, 6)
	_, valPath := fileNames(dir, TypeMD5)

	val, err := os.ReadFile(valPath)
	require.NoError(t, err)
	val[4*17] ^= 0x01
	require.NoError(t, os.WriteFile(valPath, val, 0644))

	ts, err := Open(dir)
	require.NoError(t, err)
	defer func() {
		assert.NoError(t, ts.Close())
	}()

	err = ts.Table(TypeMD5).Verify(t.Context(), VerifyOptions{Coverage: true})
	require.ErrorIs(t, err, ErrCorrupt)
	require.NoError(t, ts.Table(TypeSHA256).Verify(t.Context(), VerifyOptions{Coverage: true}))
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
			assert.Error(t, Build(t.Context(), o))
		})
	}
}

func BenchmarkLookup(b *testing.B) {
	const start, end = uint64(MinQq), uint64(MinQq + 200_000 - 1)
	dir := buildTestTables(b, start, end, DefaultPartBits)
	ts, err := Open(dir)
	require.NoError(b, err)
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
