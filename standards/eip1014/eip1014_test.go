package eip1014

import (
	"encoding/hex"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestAddressOfficialVectors(t *testing.T) {
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

		got := Address(deployer, salt, initCode)
		require.Equal(t, want, got, tt.deployer)
	}
}
