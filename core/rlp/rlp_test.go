package rlp

import (
	"encoding/hex"
	"math/big"
	"testing"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/stretchr/testify/require"
)

func mustHex(t *testing.T, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	require.NoError(t, err)
	return b
}

func TestEncodeVectors(t *testing.T) {
	lorem := "Lorem ipsum dolor sit amet, consectetur adipisicing elit"

	tests := []struct {
		name string
		in   any
		want string
	}{
		{"empty string", []byte{}, "80"},
		{"dog", []byte("dog"), "83646f67"},
		{"single byte below 0x80", []byte{0x0f}, "0f"},
		{"single byte 0x80", []byte{0x80}, "8180"},
		{"zero as uint64", uint64(0), "80"},
		{"15 as uint64", uint64(15), "0f"},
		{"1024 as uint64", uint64(1024), "820400"},
		{"zero as big.Int", big.NewInt(0), "80"},
		{"1024 as big.Int", big.NewInt(1024), "820400"},
		{"long string", []byte(lorem), "b838" + hex.EncodeToString([]byte(lorem))},
		{"empty list", []any{}, "c0"},
		{"cat dog", []any{[]byte("cat"), []byte("dog")}, "c88363617483646f67"},
		{"nested empty lists", []any{[]any{}, []any{[]any{}}, []any{[]any{}, []any{[]any{}}}}, "c7c0c1c0c3c0c1c0"},
		{"byte array", [4]byte{0xde, 0xad, 0xbe, 0xef}, "84deadbeef"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Encode(tt.in)
			require.NoError(t, err)
			require.Equal(t, mustHex(t, tt.want), got)
		})
	}
}

func TestEncodeNegativeBigInt(t *testing.T) {
	_, err := Encode(big.NewInt(-1))
	require.Error(t, err)
	require.Contains(t, err.Error(), "rlp: encode")
}

type sample struct {
	Nonce uint64
	Value *big.Int
	To    *[20]byte `rlp:"nil"`
	Data  []byte
	Keys  [][32]byte
}

func TestStructRoundTrip(t *testing.T) {
	to := [20]byte{1, 2, 3}
	in := sample{
		Nonce: 7,
		Value: big.NewInt(1_000_000),
		To:    &to,
		Data:  []byte{0xaa, 0xbb},
		Keys:  [][32]byte{{1}, {2}},
	}

	enc, err := Encode(&in)
	require.NoError(t, err)

	var out sample
	require.NoError(t, Decode(enc, &out))
	require.Equal(t, in, out)
}

func TestStructNilPointerRoundTrip(t *testing.T) {
	in := sample{Value: big.NewInt(0), Data: []byte{}, Keys: [][32]byte{}}

	enc, err := Encode(&in)
	require.NoError(t, err)

	var out sample
	require.NoError(t, Decode(enc, &out))
	require.Nil(t, out.To)
}

func TestDecodeRejectsTrailingBytes(t *testing.T) {
	var out []byte
	err := Decode(mustHex(t, "83646f6700"), &out)
	require.Error(t, err)
	require.Contains(t, err.Error(), "rlp: decode")
}

func TestDecodeRejectsNonCanonical(t *testing.T) {
	tests := []struct {
		name string
		in   string
		into func() any
	}{
		{"single byte with string prefix", "8105", func() any { return new([]byte) }},
		{"integer with leading zero", "820001", func() any { return new(*big.Int) }},
		{"uint64 with leading zero", "820001", func() any { return new(uint64) }},
		{"zero encoded as 0x00", "00", func() any { return new(uint64) }},
		{"long form for short string", "b80100", func() any { return new([]byte) }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Error(t, Decode(mustHex(t, tt.in), tt.into()))
		})
	}
}

func TestDecodeRejectsTruncated(t *testing.T) {
	var out []byte
	require.Error(t, Decode(mustHex(t, "83646f"), &out))
	require.Error(t, Decode(nil, &out))
}

func TestDecodeTypeMismatch(t *testing.T) {
	var out []byte
	require.Error(t, Decode(mustHex(t, "c0"), &out))
}

func TestAddressRoundTrip(t *testing.T) {
	a, err := types.NewAddressFromHex("0x3535353535353535353535353535353535353535")
	require.NoError(t, err)

	arr := FromAddress(a)
	require.NotNil(t, arr)
	require.True(t, a.Equal(ToAddress(arr)))
}

func TestAddressNil(t *testing.T) {
	require.Nil(t, FromAddress(nil))
	require.Nil(t, ToAddress(nil))
}

func TestAddressDoesNotAliasInput(t *testing.T) {
	a, err := types.NewAddressFromHex("0x3535353535353535353535353535353535353535")
	require.NoError(t, err)

	arr := FromAddress(a)
	arr[0] = 0xff
	require.NotEqual(t, byte(0xff), a.Bytes()[0])
}

func TestBigOrZero(t *testing.T) {
	require.Equal(t, 0, BigOrZero(nil).Sign())

	n := big.NewInt(5)
	require.Same(t, n, BigOrZero(n))
}

func TestSignatureFromRS(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex("4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes(make([]byte, 32)))
	require.NoError(t, err)

	got, err := SignatureFromRS(sig.R(), sig.S(), byte(sig.V().Uint64()))
	require.NoError(t, err)
	require.True(t, sig.Equal(got))
}

func TestSignatureFromRSRejectsOversizedValues(t *testing.T) {
	tooBig := new(big.Int).Lsh(big.NewInt(1), 256)
	_, err := SignatureFromRS(tooBig, big.NewInt(1), 0)
	require.Error(t, err)
	_, err = SignatureFromRS(big.NewInt(1), tooBig, 0)
	require.Error(t, err)
}

func TestRecoveryID(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex("4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes(make([]byte, 32)))
	require.NoError(t, err)

	got, err := RecoveryID(sig)
	require.NoError(t, err)
	require.Equal(t, byte(sig.V().Uint64()), got)
}

func TestRecoveryIDRejectsAboveOne(t *testing.T) {
	key, err := types.NewPrivateKeyFromHex("4c0883a69102937d6231471b5dbb6204fe5129617082792ae468d01a3f362318")
	require.NoError(t, err)
	sig, err := key.Sign(types.NewHashFromBytes(make([]byte, 32)))
	require.NoError(t, err)

	high, err := SignatureFromRS(sig.R(), sig.S(), 2)
	require.NoError(t, err)
	_, err = RecoveryID(high)
	require.Error(t, err)
}
