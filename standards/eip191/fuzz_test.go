package eip191

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
)

// go test -run '^$' -fuzz '^FuzzHash$' -fuzztime=10s ./standards/eip191
func FuzzHash(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("hello"))
	f.Add(make([]byte, 1024))

	f.Fuzz(func(t *testing.T, msg []byte) {
		got := Hash(msg)
		require.Len(t, got.Bytes(), 32)
	})
}

// go test -run '^$' -fuzz '^FuzzHashMatchesManualFormula$' -fuzztime=10s ./standards/eip191
//
// Differential test: Hash must always equal
// keccak256("\x19Ethereum Signed Message:\n" + strconv.Itoa(len(msg)) + msg)
// computed independently, not just be internally consistent.
func FuzzHashMatchesManualFormula(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("hello"))
	f.Add(make([]byte, 9999)) // pushes the decimal length prefix to 4 digits

	f.Fuzz(func(t *testing.T, msg []byte) {
		got := Hash(msg)

		prefixed := append([]byte("\x19Ethereum Signed Message:\n"), strconv.Itoa(len(msg))...)
		prefixed = append(prefixed, msg...)
		want := core.Keccak256(prefixed)

		require.Equal(t, want.Bytes(), got.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzHashDeterministicAndSensitive$' -fuzztime=10s ./standards/eip191
//
// Hashing the same message twice always agrees; appending any extra byte
// always changes the digest (the length prefix means a longer message can
// never collide with a shorter one's framing).
func FuzzHashDeterministicAndSensitive(f *testing.F) {
	f.Add([]byte{}, uint8(0))
	f.Add([]byte("hello"), uint8(1))

	f.Fuzz(func(t *testing.T, msg []byte, extra uint8) {
		got1 := Hash(msg)
		got2 := Hash(msg)
		require.Equal(t, got1.Bytes(), got2.Bytes())

		longer := append(append([]byte{}, msg...), extra)
		gotLonger := Hash(longer)
		require.NotEqual(t, got1.Bytes(), gotLonger.Bytes())
	})
}
