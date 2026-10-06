package rlp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func FuzzBytesRoundTrip(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add([]byte{0x7f})
	f.Add([]byte{0x80})
	f.Add(make([]byte, 56))
	f.Add(make([]byte, 1024))

	f.Fuzz(func(t *testing.T, data []byte) {
		enc, err := Encode(data)
		require.NoError(t, err)

		var out []byte
		require.NoError(t, Decode(enc, &out))
		require.Equal(t, len(data), len(out))
		require.Equal(t, string(data), string(out))
	})
}

func FuzzDecodeCanonical(f *testing.F) {
	f.Add([]byte{0x80})
	f.Add([]byte{0x83, 'd', 'o', 'g'})
	f.Add([]byte{0x81, 0x05})
	f.Add([]byte{0xb8, 0x01, 0x00})

	f.Fuzz(func(t *testing.T, data []byte) {
		var out []byte
		if err := Decode(data, &out); err != nil {
			return
		}
		enc, err := Encode(out)
		require.NoError(t, err)
		require.Equal(t, data, enc)
	})
}
