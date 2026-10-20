package eip712

import (
	"fmt"
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzEIP712EncodeArrayOfUint8$' -fuzztime=10s ./standards/eip712
func FuzzEIP712EncodeArrayOfUint8(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{1, 2, 3})
	f.Add(make([]byte, 256))

	schema := NewTypes("Simple", NewField("a", abi.Slice(abi.Uint8).String()))

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

	schema := NewTypes("Simple", NewField("a", abi.Address.String()))

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
	schema := NewTypes("Person",
		NewField("name", abi.String.String()),
		NewField("wallet", abi.Address.String()),
	)
	schema.Append("Mail",
		NewField("from", "Person"),
		NewField("to", "Person"),
		NewField("contents", abi.String.String()),
		NewField("nums", abi.Slice(abi.Uint256).String()),
	)

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
		var chainID *big.Int
		if hasChainID {
			if len(chainIDRaw) > 32 {
				chainIDRaw = chainIDRaw[len(chainIDRaw)-32:]
			}
			chainID = new(big.Int).SetBytes(chainIDRaw)
		}
		var contract *types.Address
		if hasContract {
			contract = types.NewAddressFromBytes(contractRaw)
		}
		var salt *types.Hash
		if hasSalt {
			salt = types.NewHashFromBytes(saltRaw)
		}
		domain := NewDomain(name, version, chainID, contract, salt)

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

	schema := NewTypes("Simple",
		NewField("s", abi.String.String()),
		NewField("a", abi.Address.String()),
		NewField("ok", abi.Bool.String()),
	)

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
	schema := NewTypes("Matrix", NewField("rows", abi.Slice(abi.Slice(abi.Uint256)).String()))

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
		schema := NewTypes("Bad", NewField("x", typ))
		values := []any{raw, string(raw), nil, big.NewInt(int64(len(raw))), []any{raw}, map[string]any{"x": raw}}
		for _, v := range values {
			_, _ = schema.EncodeData("Bad", map[string]any{"x": v})
		}
	})
}

// go test -run '^$' -fuzz '^FuzzFinalizeHashDeterministic$' -fuzztime=10s ./standards/eip712
func FuzzFinalizeHashDeterministic(f *testing.F) {
	f.Add("Ether Mail", "1", []byte{1}, make([]byte, 20), make([]byte, 32))
	f.Add("x", "", make([]byte, 64), make([]byte, 64), make([]byte, 64))

	f.Fuzz(func(t *testing.T, name, version string, chainIDRaw, contractRaw, structHashRaw []byte) {
		if len(chainIDRaw) > 32 {
			chainIDRaw = chainIDRaw[len(chainIDRaw)-32:]
		}
		chainID := new(big.Int).SetBytes(chainIDRaw)
		contract := types.NewAddressFromBytes(contractRaw)
		domain := NewDomain(name, version, chainID, contract, nil)
		structHash := types.NewHashFromBytes(structHashRaw)

		got1, err := FinalizeHash(domain, structHash)
		require.NoError(t, err)
		require.Len(t, got1.Bytes(), 32)

		got2, err := FinalizeHash(domain, structHash)
		require.NoError(t, err)
		require.Equal(t, got1.Bytes(), got2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzHashStructWithTypeHashMatchesHashStruct$' -fuzztime=10s ./standards/eip712
func FuzzHashStructWithTypeHashMatchesHashStruct(f *testing.F) {
	f.Add("hello", []byte{1, 2, 3}, true)
	f.Add("", []byte{}, false)

	schema := NewTypes("Simple",
		NewField("s", abi.String.String()),
		NewField("a", abi.Address.String()),
		NewField("ok", abi.Bool.String()),
	)

	f.Fuzz(func(t *testing.T, s string, addrRaw []byte, ok bool) {
		data := Message{
			"s":  s,
			"a":  types.NewAddressFromBytes(addrRaw),
			"ok": ok,
		}

		want, err := schema.HashStruct("Simple", data)
		require.NoError(t, err)

		got, err := schema.HashStructWithTypeHash(schema.TypeHash("Simple"), "Simple", data)
		require.NoError(t, err)
		require.Equal(t, want.Bytes(), got.Bytes())
	})
}
