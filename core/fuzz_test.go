package core

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzKeccak256$' -fuzztime=10s ./core
func FuzzKeccak256(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("abc"))
	f.Add(make([]byte, 1024))

	f.Fuzz(func(t *testing.T, data []byte) {
		got1 := Keccak256(data)
		got2 := Keccak256(data)
		require.Len(t, got1.Bytes(), 32)
		require.Equal(t, got1.Bytes(), got2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzEIP191Hash$' -fuzztime=10s ./core
func FuzzEIP191Hash(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("hello"))
	f.Add(make([]byte, 1024))

	f.Fuzz(func(t *testing.T, msg []byte) {
		got := EIP191Hash(msg)
		require.Len(t, got.Bytes(), 32)
	})
}

// go test -run '^$' -fuzz '^FuzzEIP712EncodeArrayOfUint8$' -fuzztime=10s ./core
func FuzzEIP712EncodeArrayOfUint8(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3})
	f.Add(make([]byte, 256))

	schema := EIP712Types{"Simple": {{Name: "a", Type: "uint8[]"}}}

	f.Fuzz(func(t *testing.T, raw []byte) {
		_, err := schema.EncodeData("Simple", map[string]any{"a": raw})
		require.NoError(t, err)
	})
}

// go test -run '^$' -fuzz '^FuzzEIP712EncodeDataAddressField$' -fuzztime=10s ./core
func FuzzEIP712EncodeDataAddressField(f *testing.F) {
	f.Add([]byte{}, false)
	f.Add(make([]byte, 20), false)
	f.Add(make([]byte, 20), true)
	f.Add(make([]byte, 5), true)

	schema := EIP712Types{"Simple": {{Name: "a", Type: "address"}}}

	f.Fuzz(func(t *testing.T, raw []byte, wrongType bool) {
		var value any
		if wrongType {
			value = raw
		} else {
			value = types.NewAddressFromBytes(raw)
		}
		_, _ = schema.EncodeData("Simple", map[string]any{"a": value})
	})
}

// go test -run '^$' -fuzz '^FuzzEIP712EncodeDataCrashSafety$' -fuzztime=10s ./core
func FuzzEIP712EncodeDataCrashSafety(f *testing.F) {
	schema := EIP712Types{
		"Person": {{Name: "name", Type: "string"}, {Name: "wallet", Type: "address"}},
		"Mail": {
			{Name: "from", Type: "Person"},
			{Name: "to", Type: "Person"},
			{Name: "contents", Type: "string"},
			{Name: "nums", Type: "uint256[]"},
		},
	}

	f.Add("hello", []byte{1, 2, 3}, int8(5))
	f.Add("", []byte(nil), int8(0))

	f.Fuzz(func(t *testing.T, s string, b []byte, n int8) {
		values := []any{
			s, b, n, nil,
			map[string]any{"name": s, "wallet": b},
			[]any{s, b, n},
		}
		for _, v := range values {
			_, _ = schema.EncodeData("Mail", map[string]any{
				"from":     v,
				"to":       v,
				"contents": v,
				"nums":     v,
			})
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEIP1014Address$' -fuzztime=10s ./core
func FuzzEIP1014Address(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 32), []byte{})
	f.Add(make([]byte, 20), make([]byte, 32), []byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, deployerBytes []byte, saltBytes []byte, initCode []byte) {
		deployer := types.NewAddressFromBytes(deployerBytes)

		var salt [32]byte
		copy(salt[:], saltBytes)

		got1 := EIP1014Address(deployer, salt, initCode)
		got2 := EIP1014Address(deployer, salt, initCode)
		require.Equal(t, got1, got2)
	})
}
