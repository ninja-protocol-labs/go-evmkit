package eip712

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzEIP712EncodeArrayOfUint8$' -fuzztime=10s ./standards/eip712
func FuzzEIP712EncodeArrayOfUint8(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3})
	f.Add(make([]byte, 256))

	schema := Types{"Simple": {{Name: "a", Type: "uint8[]"}}}

	f.Fuzz(func(t *testing.T, raw []byte) {
		_, err := schema.EncodeData("Simple", map[string]any{"a": raw})
		require.NoError(t, err)
	})
}

// go test -run '^$' -fuzz '^FuzzEIP712EncodeDataAddressField$' -fuzztime=10s ./standards/eip712
func FuzzEIP712EncodeDataAddressField(f *testing.F) {
	f.Add([]byte{}, false)
	f.Add(make([]byte, 20), false)
	f.Add(make([]byte, 20), true)
	f.Add(make([]byte, 5), true)

	schema := Types{"Simple": {{Name: "a", Type: "address"}}}

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

// go test -run '^$' -fuzz '^FuzzEIP712EncodeDataCrashSafety$' -fuzztime=10s ./standards/eip712
func FuzzEIP712EncodeDataCrashSafety(f *testing.F) {
	schema := Types{
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

// go test -run '^$' -fuzz '^FuzzBaseTypeName$' -fuzztime=10s ./standards/eip712
func FuzzBaseTypeName(f *testing.F) {
	f.Add("")
	f.Add("Person")
	f.Add("Person[]")
	f.Add("Person[3]")
	f.Add("Person[3][]")
	f.Add("[")
	f.Add("]")
	f.Add("[[[[[")
	f.Add("Person[3][][][][][][][][][][][][][][][][][][][][]")

	f.Fuzz(func(t *testing.T, s string) {
		got := baseTypeName(s)
		require.NotContains(t, got, "[")
		require.LessOrEqual(t, len(got), len(s))
		require.Equal(t, got, baseTypeName(got))
	})
}

// go test -run '^$' -fuzz '^FuzzDomainSeparatorCrashSafety$' -fuzztime=10s ./standards/eip712
func FuzzDomainSeparatorCrashSafety(f *testing.F) {
	f.Add("", "", []byte{}, []byte{}, []byte{}, false, false, false)
	f.Add("Ether Mail", "1", []byte{1}, make([]byte, 20), []byte("salt"), true, true, true)
	f.Add("x", "", make([]byte, 64), make([]byte, 64), make([]byte, 64), true, true, true)

	f.Fuzz(func(t *testing.T, name, version string, chainIDRaw, contractRaw, saltRaw []byte, hasChainID, hasContract, hasSalt bool) {
		domain := Domain{Name: name, Version: version}
		if hasChainID {
			if len(chainIDRaw) > 32 {
				chainIDRaw = chainIDRaw[len(chainIDRaw)-32:]
			}
			domain.ChainID = new(big.Int).SetBytes(chainIDRaw)
		}
		if hasContract {
			domain.VerifyingContract = types.NewAddressFromBytes(contractRaw)
		}
		if hasSalt {
			domain.Salt = types.NewHashFromBytes(saltRaw)
		}

		if name == "" && version == "" && !hasChainID && !hasContract && !hasSalt {
			_, err := DomainSeparator(domain)
			require.Error(t, err)
			return
		}

		got, err := DomainSeparator(domain)
		require.NoError(t, err)
		require.Len(t, got.Bytes(), 32)

		got2, err := DomainSeparator(domain)
		require.NoError(t, err)
		require.Equal(t, got.Bytes(), got2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeTypeCyclicGraphNeverPanics$' -fuzztime=10s ./standards/eip712
func FuzzEncodeTypeCyclicGraphNeverPanics(f *testing.F) {
	f.Add([]byte{0, 1, 2, 3, 4})
	f.Add([]byte{0, 0, 0, 0, 0})
	f.Add([]byte{4, 3, 2, 1, 0})
	f.Add([]byte{})

	const n = 5
	names := make([]string, n)
	for i := range names {
		names[i] = fmt.Sprintf("T%d", i)
	}
	scalars := []string{"uint256", "address", "string", "bool"}

	f.Fuzz(func(t *testing.T, seed []byte) {
		ts := make(Types, n)
		for i, name := range names {
			fieldType := scalars[0]
			if len(seed) > 0 {
				b := seed[i%len(seed)]
				if b%2 == 0 {
					fieldType = names[int(b)%n]
				} else {
					fieldType = scalars[int(b)%len(scalars)]
				}
			}
			ts[name] = []Field{{Name: "f", Type: fieldType}}
		}

		got1 := ts.EncodeType("T0")
		got2 := ts.EncodeType("T0")
		require.Equal(t, got1, got2)
	})
}

// go test -run '^$' -fuzz '^FuzzHashStructDeterministic$' -fuzztime=10s ./standards/eip712
func FuzzHashStructDeterministic(f *testing.F) {
	f.Add("hello", []byte{1, 2, 3}, true)
	f.Add("", []byte{}, false)

	schema := Types{"Simple": {
		{Name: "s", Type: "string"},
		{Name: "a", Type: "address"},
		{Name: "ok", Type: "bool"},
	}}

	f.Fuzz(func(t *testing.T, s string, addrRaw []byte, ok bool) {
		data := map[string]any{
			"s":  s,
			"a":  types.NewAddressFromBytes(addrRaw),
			"ok": ok,
		}

		got1, err := schema.HashStruct("Simple", data)
		require.NoError(t, err)
		require.Len(t, got1.Bytes(), 32)

		got2, err := schema.HashStruct("Simple", data)
		require.NoError(t, err)
		require.Equal(t, got1.Bytes(), got2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeArrayValueNestedCrashSafety$' -fuzztime=10s ./standards/eip712
func FuzzEncodeArrayValueNestedCrashSafety(f *testing.F) {
	schema := Types{"Matrix": {{Name: "rows", Type: "uint256[][]"}}}

	f.Add(uint8(0), uint8(0), uint64(1))
	f.Add(uint8(3), uint8(3), uint64(7))
	f.Add(uint8(255), uint8(0), uint64(0))
	f.Add(uint8(0), uint8(255), uint64(0))

	f.Fuzz(func(t *testing.T, outerLen, innerLen uint8, v uint64) {
		outer := make([][]*big.Int, int(outerLen)%16)
		for i := range outer {
			inner := make([]*big.Int, int(innerLen)%16)
			for j := range inner {
				inner[j] = new(big.Int).SetUint64(v)
			}
			outer[i] = inner
		}

		_, err := schema.EncodeData("Matrix", map[string]any{"rows": outer})
		require.NoError(t, err)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeFieldValueGarbageTypeCrashSafety$' -fuzztime=10s ./standards/eip712
func FuzzEncodeFieldValueGarbageTypeCrashSafety(f *testing.F) {
	f.Add("uint256", []byte{1, 2, 3})
	f.Add("bytes32[]", []byte{})
	f.Add("notarealtype", []byte{0xff})
	f.Add("", []byte(nil))
	f.Add("Bad", []byte{0x01})

	f.Fuzz(func(t *testing.T, typ string, raw []byte) {
		schema := Types{"Bad": {{Name: "x", Type: typ}}}
		values := []any{raw, string(raw), nil, big.NewInt(int64(len(raw))), []any{raw}, map[string]any{"x": raw}}
		for _, v := range values {
			_, _ = schema.EncodeData("Bad", map[string]any{"x": v})
		}
	})
}
