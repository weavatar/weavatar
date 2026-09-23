package mphf

import (
	"crypto/md5"
	"encoding/binary"
	"errors"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randomKeys(t testing.TB, n, keySize int, seed uint64) []byte {
	t.Helper()
	r := rand.New(rand.NewPCG(seed, seed^0xABCDEF))
	keys := make([]byte, n*keySize)
	for i := range n {
		key := keys[i*keySize : (i+1)*keySize]
		for j := 0; j+8 <= keySize; j += 8 {
			binary.LittleEndian.PutUint64(key[j:], r.Uint64())
		}
		for j := keySize &^ 7; j < keySize; j++ {
			key[j] = byte(r.Uint32())
		}
	}
	return keys
}

// 用连续整数的 MD5 摘要模拟生产环境的键
func digestKeys(n int) []byte {
	keys := make([]byte, n*md5.Size)
	var buf [8]byte
	for i := range n {
		binary.LittleEndian.PutUint64(buf[:], uint64(i))
		sum := md5.Sum(buf[:])
		copy(keys[i*md5.Size:], sum[:])
	}
	return keys
}

func assertBijection(t *testing.T, m *MPHF, keys []byte, keySize int) {
	t.Helper()
	n := len(keys) / keySize
	require.Equal(t, uint64(n), m.KeyCount())

	seen := make([]bool, n)
	for i := range n {
		key := keys[i*keySize : (i+1)*keySize]
		slot, ok := m.Find(key)
		require.True(t, ok, "key %d not found", i)
		require.Less(t, slot, uint64(n), "key %d slot out of range", i)
		require.False(t, seen[slot], "key %d collides at slot %d", i, slot)
		seen[slot] = true
	}
}

func TestBuildFind(t *testing.T) {
	const n, keySize = 200_000, 16
	keys := randomKeys(t, n, keySize, 1)

	m, err := Build(keys, keySize, Options{})
	require.NoError(t, err)
	assertBijection(t, m, keys, keySize)

	assert.Greater(t, m.Levels(), 1)
	assert.Less(t, m.BitsPerKey(), 4.5)
	assert.Greater(t, m.BitsPerKey(), 3.0)
	t.Logf("n=%d levels=%d bits/key=%.2f size=%d", n, m.Levels(), m.BitsPerKey(), m.Size())
}

func TestDigestKeys(t *testing.T) {
	const n = 100_000
	keys := digestKeys(n)

	m, err := Build(keys, md5.Size, Options{Gamma: 2})
	require.NoError(t, err)
	assertBijection(t, m, keys, md5.Size)
}

func TestKeySizes(t *testing.T) {
	for _, keySize := range []int{7, 8, 16, 20, 32, 33} {
		t.Run(fmt.Sprintf("size%d", keySize), func(t *testing.T) {
			keys := randomKeys(t, 20_000, keySize, uint64(keySize))
			m, err := Build(keys, keySize, Options{})
			require.NoError(t, err)
			assertBijection(t, m, keys, keySize)
		})
	}

	// 单字节键只有 256 种取值，全部用上
	t.Run("size1", func(t *testing.T) {
		keys := make([]byte, 256)
		for i := range keys {
			keys[i] = byte(i)
		}
		m, err := Build(keys, 1, Options{})
		require.NoError(t, err)
		assertBijection(t, m, keys, 1)
	})
}

func TestGamma(t *testing.T) {
	keys := randomKeys(t, 50_000, 16, 7)
	var prevBits float64
	for _, gamma := range []float64{1, 1.5, 2, 3} {
		m, err := Build(keys, 16, Options{Gamma: gamma})
		require.NoError(t, err)
		assertBijection(t, m, keys, 16)
		assert.Greater(t, m.BitsPerKey(), prevBits, "gamma=%v", gamma)
		prevBits = m.BitsPerKey()
	}
}

func TestSmall(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 10, 447, 448, 449} {
		keys := randomKeys(t, n, 16, uint64(n)+100)
		m, err := Build(keys, 16, Options{})
		require.NoError(t, err, "n=%d", n)
		assertBijection(t, m, keys, 16)

		loaded, err := Load(m.Bytes())
		require.NoError(t, err, "n=%d", n)
		assertBijection(t, loaded, keys, 16)
	}

	m, err := Build(nil, 16, Options{})
	require.NoError(t, err)
	_, ok := m.Find(make([]byte, 16))
	assert.False(t, ok)
}

func TestInvalidKeySize(t *testing.T) {
	_, err := Build(make([]byte, 10), 0, Options{})
	assert.Error(t, err)
	_, err = Build(make([]byte, 10), 3, Options{})
	assert.Error(t, err)
}

func TestDuplicateKey(t *testing.T) {
	keys := randomKeys(t, 1000, 16, 3)
	copy(keys[500*16:], keys[7*16:8*16])

	_, err := Build(keys, 16, Options{})
	require.ErrorIs(t, err, ErrDuplicateKey)
}

func TestLoadRoundTrip(t *testing.T) {
	const n, keySize = 50_000, 32
	keys := randomKeys(t, n, keySize, 5)
	m, err := Build(keys, keySize, Options{})
	require.NoError(t, err)

	data := m.Bytes()
	require.Len(t, data, m.Size())

	aligned, err := Load(data)
	require.NoError(t, err)
	assert.Equal(t, m.Levels(), aligned.Levels())
	assert.Equal(t, m.Size(), aligned.Size())

	// 故意错开一个字节，走拷贝路径
	shifted := make([]byte, len(data)+1)
	copy(shifted[1:], data)
	misaligned, err := Load(shifted[1:])
	require.NoError(t, err)

	for i := range n {
		key := keys[i*keySize : (i+1)*keySize]
		want, ok := m.Find(key)
		require.True(t, ok)
		got, ok := aligned.Find(key)
		require.True(t, ok)
		assert.Equal(t, want, got)
		got, ok = misaligned.Find(key)
		require.True(t, ok)
		assert.Equal(t, want, got)
	}
}

func TestLoadCorrupt(t *testing.T) {
	keys := randomKeys(t, 1000, 16, 9)
	m, err := Build(keys, 16, Options{})
	require.NoError(t, err)
	data := m.Bytes()

	cases := map[string][]byte{
		"empty":     {},
		"short":     data[:16],
		"unaligned": data[:len(data)-3],
		"truncated": data[:len(data)-8],
		"extra":     append(append([]byte{}, data...), make([]byte, 8)...),
	}
	badMagic := append([]byte{}, data...)
	badMagic[0] ^= 0xff
	cases["magic"] = badMagic
	badLevels := append([]byte{}, data...)
	binary.LittleEndian.PutUint64(badLevels[24:], 1<<40)
	cases["levels"] = badLevels
	badBlocks := append([]byte{}, data...)
	binary.LittleEndian.PutUint64(badBlocks[32:], 0)
	cases["blocks"] = badBlocks

	for name, c := range cases {
		_, err := Load(c)
		assert.True(t, errors.Is(err, ErrCorrupt), "%s: %v", name, err)
	}
}

func TestNonMember(t *testing.T) {
	const n = 100_000
	keys := randomKeys(t, n, 16, 11)
	m, err := Build(keys, 16, Options{})
	require.NoError(t, err)

	others := randomKeys(t, 10_000, 16, 12)
	found := 0
	for i := range 10_000 {
		slot, ok := m.Find(others[i*16 : (i+1)*16])
		if ok {
			found++
			assert.Less(t, slot, uint64(n))
		}
	}
	// 非成员键大概率会命中某个槽，只验证不越界
	t.Logf("non-member keys reported found: %d/10000", found)
}

func TestSeedRetry(t *testing.T) {
	keys := randomKeys(t, 10_000, 16, 13)
	// 只允许一层必然放不完，重试耗尽后应失败
	_, err := Build(keys, 16, Options{MaxLevels: 1})
	require.ErrorIs(t, err, ErrBuildFailed)

	m, err := Build(keys, 16, Options{MaxLevels: 64, Seed: 42})
	require.NoError(t, err)
	assertBijection(t, m, keys, 16)
}

func BenchmarkFind(b *testing.B) {
	const n, keySize = 1_000_000, 16
	keys := randomKeys(b, n, keySize, 21)
	m, err := Build(keys, keySize, Options{})
	require.NoError(b, err)
	b.Logf("levels=%d bits/key=%.2f", m.Levels(), m.BitsPerKey())

	b.ReportAllocs()
	var sink uint64
	i := 0
	for b.Loop() {
		idx := (i * 7919) % n
		slot, _ := m.Find(keys[idx*keySize : (idx+1)*keySize])
		sink += slot
		i++
	}
	_ = sink
}

func BenchmarkBuild(b *testing.B) {
	const n, keySize = 1_000_000, 16
	keys := randomKeys(b, n, keySize, 22)

	b.ReportAllocs()
	for b.Loop() {
		_, err := Build(keys, keySize, Options{})
		require.NoError(b, err)
	}
}
