package mphf

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand/v2"
	"testing"

	"github.com/libtnb/assert/check"
	"github.com/libtnb/assert/must"
)

func TestBuildFind(t *testing.T) {
	const n, keySize = 200_000, 16
	keys := randomKeys(t, n, keySize, 1)

	m, slots, err := Build(keys, keySize, Options{})
	must.NoError(t, err)
	assertBijection(t, m, keys, keySize, slots)

	check.Greater(t, m.Levels(), 1)
	check.Less(t, m.BitsPerKey(), 4.5)
	check.Greater(t, m.BitsPerKey(), 3.0)
	t.Logf("n=%d levels=%d bits/key=%.2f size=%d", n, m.Levels(), m.BitsPerKey(), m.Size())
}

func TestDigestKeys(t *testing.T) {
	const n = 100_000
	keys := digestKeys(n)

	m, slots, err := Build(keys, md5.Size, Options{Gamma: 2})
	must.NoError(t, err)
	assertBijection(t, m, keys, md5.Size, slots)
}

func TestKeySizes(t *testing.T) {
	for _, keySize := range []int{7, 8, 16, 20, 32, 33} {
		t.Run(fmt.Sprintf("size%d", keySize), func(t *testing.T) {
			keys := randomKeys(t, 20_000, keySize, uint64(keySize))
			m, slots, err := Build(keys, keySize, Options{})
			must.NoError(t, err)
			assertBijection(t, m, keys, keySize, slots)
		})
	}

	// one-byte keys have only 256 values; use them all
	t.Run("size1", func(t *testing.T) {
		keys := make([]byte, 256)
		for i := range keys {
			keys[i] = byte(i)
		}
		m, slots, err := Build(keys, 1, Options{})
		must.NoError(t, err)
		assertBijection(t, m, keys, 1, slots)
	})
}

func TestGamma(t *testing.T) {
	keys := randomKeys(t, 50_000, 16, 7)
	var prevBits float64
	for _, gamma := range []float64{1, 1.5, 2, 3} {
		m, slots, err := Build(keys, 16, Options{Gamma: gamma})
		must.NoError(t, err)
		assertBijection(t, m, keys, 16, slots)
		check.Greater(t, m.BitsPerKey(), prevBits, check.Msgf("gamma=%v", gamma))
		prevBits = m.BitsPerKey()
	}
}

func TestSmall(t *testing.T) {
	for _, n := range []int{0, 1, 2, 3, 10, 447, 448, 449} {
		keys := randomKeys(t, n, 16, uint64(n)+100)
		m, slots, err := Build(keys, 16, Options{})
		must.NoError(t, err, must.Msgf("n=%d", n))
		assertBijection(t, m, keys, 16, slots)

		loaded, err := Load(m.Bytes())
		must.NoError(t, err, must.Msgf("n=%d", n))
		assertBijection(t, loaded, keys, 16, nil)
	}

	m, _, err := Build(nil, 16, Options{})
	must.NoError(t, err)
	_, ok := m.Find(make([]byte, 16))
	check.False(t, ok)
}

func TestInvalidKeySize(t *testing.T) {
	_, _, err := Build(make([]byte, 10), 0, Options{})
	check.Error(t, err)
	_, _, err = Build(make([]byte, 10), 3, Options{})
	check.Error(t, err)
}

func TestDuplicateKey(t *testing.T) {
	keys := randomKeys(t, 1000, 16, 3)
	copy(keys[500*16:], keys[7*16:8*16])

	_, _, err := Build(keys, 16, Options{})
	must.ErrorIs(t, err, ErrDuplicateKey)
}

func TestLoadRoundTrip(t *testing.T) {
	const n, keySize = 50_000, 32
	keys := randomKeys(t, n, keySize, 5)
	m, _, err := Build(keys, keySize, Options{})
	must.NoError(t, err)

	data := m.Bytes()
	must.Len(t, data, m.Size())

	aligned, err := Load(data)
	must.NoError(t, err)
	check.Equal(t, aligned.Levels(), m.Levels())
	check.Equal(t, aligned.Size(), m.Size())

	// offset by one byte to take the copying path
	shifted := make([]byte, len(data)+1)
	copy(shifted[1:], data)
	misaligned, err := Load(shifted[1:])
	must.NoError(t, err)

	for i := range n {
		key := keys[i*keySize : (i+1)*keySize]
		want, ok := m.Find(key)
		must.True(t, ok)
		got, ok := aligned.Find(key)
		must.True(t, ok)
		check.Equal(t, got, want)
		got, ok = misaligned.Find(key)
		must.True(t, ok)
		check.Equal(t, got, want)
	}
}

func TestLoadCorrupt(t *testing.T) {
	keys := randomKeys(t, 1000, 16, 9)
	m, _, err := Build(keys, 16, Options{})
	must.NoError(t, err)
	data := m.Bytes()

	cases := map[string][]byte{
		"empty":     {},
		"short":     data[:16],
		"unaligned": data[:len(data)-3],
		"truncated": data[:len(data)-8],
		"extra":     append(append([]byte{}, data...), make([]byte, 8)...),
	}
	badLevels := append([]byte{}, data...)
	binary.LittleEndian.PutUint64(badLevels[16:], 1<<40)
	cases["levels"] = badLevels
	badBlocks := append([]byte{}, data...)
	binary.LittleEndian.PutUint64(badBlocks[24:], 0)
	cases["blocks"] = badBlocks

	for name, c := range cases {
		_, err := Load(c)
		check.ErrorIs(t, err, ErrCorrupt, name)
	}
}

func TestNonMember(t *testing.T) {
	const n = 100_000
	keys := randomKeys(t, n, 16, 11)
	m, _, err := Build(keys, 16, Options{})
	must.NoError(t, err)

	others := randomKeys(t, 10_000, 16, 12)
	found := 0
	for i := range 10_000 {
		slot, ok := m.Find(others[i*16 : (i+1)*16])
		if ok {
			found++
			check.Less(t, slot, uint64(n))
		}
	}
	// non-members usually land on some slot; only the bounds are guaranteed
	t.Logf("non-member keys reported found: %d/10000", found)
}

func TestSeedRetry(t *testing.T) {
	keys := randomKeys(t, 10_000, 16, 13)
	// a single level can never place every key, so the retries run out
	_, _, err := Build(keys, 16, Options{MaxLevels: 1})
	must.ErrorIs(t, err, ErrBuildFailed)

	m, slots, err := Build(keys, 16, Options{MaxLevels: 64, Seed: 42})
	must.NoError(t, err)
	assertBijection(t, m, keys, 16, slots)
}

func randomKeys(tb testing.TB, n, keySize int, seed uint64) []byte {
	tb.Helper()
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

// digestKeys mimics production keys with MD5 digests of consecutive integers.
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

func assertBijection(t *testing.T, m *MPHF, keys []byte, keySize int, slots []uint32) {
	t.Helper()
	n := len(keys) / keySize
	must.Equal(t, m.KeyCount(), uint64(n))

	seen := make([]bool, n)
	for i := range n {
		key := keys[i*keySize : (i+1)*keySize]
		slot, ok := m.Find(key)
		must.True(t, ok, must.Msgf("key %d not found", i))
		must.Less(t, slot, uint64(n), must.Msgf("key %d slot out of range", i))
		must.False(t, seen[slot], must.Msgf("key %d collides at slot %d", i, slot))
		seen[slot] = true
		if slots != nil {
			must.Equal(t, uint64(slots[i]), slot, must.Msgf("key %d: Build slot differs from Find", i))
		}
	}
}

func BenchmarkFind(b *testing.B) {
	const n, keySize = 1_000_000, 16
	keys := randomKeys(b, n, keySize, 21)
	m, _, err := Build(keys, keySize, Options{})
	must.NoError(b, err)
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
		_, _, err := Build(keys, keySize, Options{})
		must.NoError(b, err)
	}
}
