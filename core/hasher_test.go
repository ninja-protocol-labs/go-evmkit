package core

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestKeccak256Empty(t *testing.T) {
	got := Keccak256([]byte{})
	want, err := hex.DecodeString("c5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestKeccak256Abc(t *testing.T) {
	got := Keccak256([]byte("abc"))
	want, err := hex.DecodeString("4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestKeccak256Deterministic(t *testing.T) {
	data := []byte("the quick brown fox")
	require.Equal(t, Keccak256(data).Bytes(), Keccak256(data).Bytes())
}

func TestKeccak256DifferentInputsDifferentHashes(t *testing.T) {
	require.NotEqual(t, Keccak256([]byte("a")).Bytes(), Keccak256([]byte("b")).Bytes())
}

func TestEIP191HashHello(t *testing.T) {
	got := EIP191Hash([]byte("hello"))
	want, err := hex.DecodeString("50b2c43fd39106bafbba0da34fc430e1f91e3c96ea2acee2bc34119f92b37750")
	require.NoError(t, err)
	require.Equal(t, want, got.Bytes())
}

func TestEIP191HashEmpty(t *testing.T) {
	got := EIP191Hash([]byte{})
	want := Keccak256([]byte("\x19Ethereum Signed Message:\n0"))
	require.Equal(t, want.Bytes(), got.Bytes())
}

func TestEIP191HashDiffersFromRawKeccak256(t *testing.T) {
	msg := []byte("hello")
	require.NotEqual(t, Keccak256(msg).Bytes(), EIP191Hash(msg).Bytes())
}

func TestEIP1014AddressOfficialVectors(t *testing.T) {
	tests := []struct {
		deployer string
		salt     string
		initCode string
		want     string
	}{
		{
			"0x0000000000000000000000000000000000000000",
			"0000000000000000000000000000000000000000000000000000000000000000",
			"00",
			"0x4D1A2e2bB4F88F0250f26Ffff098B0b30B26BF38",
		},
		{
			"0xdeadbeef00000000000000000000000000000000",
			"0000000000000000000000000000000000000000000000000000000000000000",
			"00",
			"0xB928f69Bb1D91Cd65274e3c79d8986362984fDA3",
		},
		{
			"0xdeadbeef00000000000000000000000000000000",
			"000000000000000000000000feed000000000000000000000000000000000000",
			"00",
			"0xD04116cDd17beBE565EB2422F2497E06cC1C9833",
		},
		{
			"0x0000000000000000000000000000000000000000",
			"0000000000000000000000000000000000000000000000000000000000000000",
			"deadbeef",
			"0x70f2b2914A2a4b783FaEFb75f459A580616Fcb5e",
		},
		{
			"0x00000000000000000000000000000000deadbeef",
			"00000000000000000000000000000000000000000000000000000000cafebabe",
			"deadbeef",
			"0x60f3f640a8508fC6a86d45DF051962668E1e8AC7",
		},
		{
			"0x00000000000000000000000000000000deadbeef",
			"00000000000000000000000000000000000000000000000000000000cafebabe",
			"deadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeefdeadbeef",
			"0x1d8bfDC5D46DC4f61D6b6115972536eBE6A8854C",
		},
		{
			"0x0000000000000000000000000000000000000000",
			"0000000000000000000000000000000000000000000000000000000000000000",
			"",
			"0xE33C0C7F7df4809055C3ebA6c09CFe4BaF1BD9e0",
		},
	}

	for _, tt := range tests {
		deployer, err := types.NewAddressFromHex(tt.deployer)
		require.NoError(t, err)

		saltBytes, err := hex.DecodeString(tt.salt)
		require.NoError(t, err)
		require.Len(t, saltBytes, 32)
		var salt [32]byte
		copy(salt[:], saltBytes)

		initCode, err := hex.DecodeString(tt.initCode)
		require.NoError(t, err)

		want, err := types.NewAddressFromHex(tt.want)
		require.NoError(t, err)

		got := EIP1014Address(deployer, salt, initCode)
		require.Equal(t, want, got, tt.deployer)
	}
}

func TestCreateAddressKnownVectors(t *testing.T) {
	sender, err := types.NewAddressFromHex("0x6ac7ea33f8831ea9dcc53393aaa88b25a785dbf0")
	require.NoError(t, err)

	tests := []struct {
		nonce uint64
		want  string
	}{
		{0, "0xcd234a471b72ba2f1ccf0a70fcaba648a5eecd8d"},
		{1, "0x343c43a37d37dff08ae8c4a11544c718abb4fcf8"},
		{2, "0xf778b86fa74e846c4f0a1fbd1335fe81c00a0c91"},
		{3, "0xfffd933a0bc612844eaf0c6fe3e5b8e9b6c1d19c"},
	}

	for _, tt := range tests {
		want, err := types.NewAddressFromHex(tt.want)
		require.NoError(t, err)

		got, err := CreateAddress(sender, tt.nonce)
		require.NoError(t, err)
		require.Equal(t, want, got, tt.nonce)
	}
}

func TestCreateAddressDeterministic(t *testing.T) {
	sender, err := types.NewAddressFromHex("0x6ac7ea33f8831ea9dcc53393aaa88b25a785dbf0")
	require.NoError(t, err)

	got1, err := CreateAddress(sender, 7)
	require.NoError(t, err)
	got2, err := CreateAddress(sender, 7)
	require.NoError(t, err)
	require.Equal(t, got1, got2)
}

func TestCreateAddressDiffersByNonce(t *testing.T) {
	sender, err := types.NewAddressFromHex("0x6ac7ea33f8831ea9dcc53393aaa88b25a785dbf0")
	require.NoError(t, err)

	got0, err := CreateAddress(sender, 0)
	require.NoError(t, err)
	got1, err := CreateAddress(sender, 1)
	require.NoError(t, err)
	require.NotEqual(t, got0, got1)
}
