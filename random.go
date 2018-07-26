package sp2p

import (
	crand "crypto/rand"
	mrand "math/rand"
	"sync"
	"time"
)

const (
	strChars = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz" // 62 characters
)

// pseudo random number generator.
// seeded with OS randomness (crand)
var prng struct {
	sync.Mutex
	*mrand.Rand
}

func reset() {
	b := cRandBytes(8)
	var seed uint64
	for i := 0; i < 8; i++ {
		seed |= uint64(b[i])
		seed <<= 8
	}
	prng.Lock()
	prng.Rand = mrand.New(mrand.NewSource(int64(seed)))
	prng.Unlock()
}

func init() {
	reset()
}

// Constructs an alphanumeric string of given length.
// It is not safe for cryptographic usage.
func randStr(length int) string {
	chars := []byte{}
MAIN_LOOP:
	for {
		val := randInt63()
		for i := 0; i < 10; i++ {
			v := int(val & 0x3f) // rightmost 6 bits
			if v >= 62 { // only 62 characters in strChars
				val >>= 6
				continue
			} else {
				chars = append(chars, strChars[v])
				if len(chars) == length {
					break MAIN_LOOP
				}
				val >>= 6
			}
		}
	}

	return string(chars)
}

// It is not safe for cryptographic usage.
func randUint16() uint16 {
	return uint16(randUint32() & (1<<16 - 1))
}

// It is not safe for cryptographic usage.
func randUint32() uint32 {
	prng.Lock()
	u32 := prng.Uint32()
	prng.Unlock()
	return u32
}

// It is not safe for cryptographic usage.
func randUint64() uint64 {
	return uint64(randUint32())<<32 + uint64(randUint32())
}

// It is not safe for cryptographic usage.
func randUint() uint {
	prng.Lock()
	i := prng.Int()
	prng.Unlock()
	return uint(i)
}

// It is not safe for cryptographic usage.
func randInt16() int16 {
	return int16(randUint32() & (1<<16 - 1))
}

// It is not safe for cryptographic usage.
func randInt32() int32 {
	return int32(randUint32())
}

// It is not safe for cryptographic usage.
func randInt64() int64 {
	return int64(randUint64())
}

// It is not safe for cryptographic usage.
func randInt() int {
	prng.Lock()
	i := prng.Int()
	prng.Unlock()
	return i
}

// It is not safe for cryptographic usage.
func randInt31() int32 {
	prng.Lock()
	i31 := prng.Int31()
	prng.Unlock()
	return i31
}

// It is not safe for cryptographic usage.
func randInt63() int64 {
	prng.Lock()
	i63 := prng.Int63()
	prng.Unlock()
	return i63
}

// Distributed pseudo-exponentially to test for various cases
// It is not safe for cryptographic usage.
func randUint16Exp() uint16 {
	bits := randUint32() % 16
	if bits == 0 {
		return 0
	}
	n := uint16(1 << (bits - 1))
	n += uint16(randInt31()) & ((1 << (bits - 1)) - 1)
	return n
}

// Distributed pseudo-exponentially to test for various cases
// It is not safe for cryptographic usage.
func randUint32Exp() uint32 {
	bits := randUint32() % 32
	if bits == 0 {
		return 0
	}
	n := uint32(1 << (bits - 1))
	n += uint32(randInt31()) & ((1 << (bits - 1)) - 1)
	return n
}

// Distributed pseudo-exponentially to test for various cases
// It is not safe for cryptographic usage.
func randUint64Exp() uint64 {
	bits := randUint32() % 64
	if bits == 0 {
		return 0
	}
	n := uint64(1 << (bits - 1))
	n += uint64(randInt63()) & ((1 << (bits - 1)) - 1)
	return n
}

// It is not safe for cryptographic usage.
func randFloat32() float32 {
	prng.Lock()
	f32 := prng.Float32()
	prng.Unlock()
	return f32
}

// It is not safe for cryptographic usage.
func randTime() time.Time {
	return time.Unix(int64(randUint64Exp()), 0)
}

// RandBytes returns n random bytes from the OS's source of entropy ie. via crypto/rand.
// It is not safe for cryptographic usage.
func randBytes(n int) []byte {
	// cRandBytes isn't guaranteed to be fast so instead
	// use random bytes generated from the internal PRNG
	bs := make([]byte, n)
	for i := 0; i < len(bs); i++ {
		bs[i] = byte(randInt() & 0xFF)
	}
	return bs
}

// RandIntn returns, as an int, a non-negative pseudo-random number in [0, n).
// It panics if n <= 0.
// It is not safe for cryptographic usage.
func randIntn(n int) int {
	prng.Lock()
	i := prng.Intn(n)
	prng.Unlock()
	return i
}

// RandPerm returns a pseudo-random permutation of n integers in [0, n).
// It is not safe for cryptographic usage.
func randPerm(n int) []int {
	prng.Lock()
	perm := prng.Perm(n)
	prng.Unlock()
	return perm
}

// NOTE: This relies on the os's random number generator.
// For real security, we should salt that with some seed.
// See github.com/tendermint/go-crypto for a more secure reader.
func cRandBytes(numBytes int) []byte {
	b := make([]byte, numBytes)
	_, err := crand.Read(b)
	if err != nil {
		mustNotErr(err)
	}
	return b
}

func rand32(max uint32) uint32 {
	if max == 0 {
		return 0
	}
	mrand.Seed(time.Now().Unix())
	return mrand.Uint32() % max
}
