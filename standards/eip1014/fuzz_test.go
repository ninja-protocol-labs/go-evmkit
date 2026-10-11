package eip1014

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzAddress$' -fuzztime=10s ./standards/eip1014
func FuzzAddress(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 32), []byte{})
	f.Add(make([]byte, 20), make([]byte, 32), []byte{0xde, 0xad, 0xbe, 0xef})

	f.Fuzz(func(t *testing.T, deployerBytes []byte, saltBytes []byte, initCode []byte) {
		deployer := types.NewAddressFromBytes(deployerBytes)

		var salt [32]byte
		copy(salt[:], saltBytes)

		got1 := Address(deployer, salt, initCode)
		got2 := Address(deployer, salt, initCode)
		require.Equal(t, got1, got2)
	})
}

// go test -run '^$' -fuzz '^FuzzAddressMatchesManualFormula$' -fuzztime=10s ./standards/eip1014
//
// Differential test: Address must always equal the EIP-1014 formula
// computed independently, byte for byte, not just be internally consistent.
func FuzzAddressMatchesManualFormula(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 32), []byte{})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 32), []byte{0x60, 0x80})
	f.Add(make([]byte, 20), []byte{0xff}, make([]byte, 1024))

	f.Fuzz(func(t *testing.T, deployerBytes []byte, saltBytes []byte, initCode []byte) {
		deployer := types.NewAddressFromBytes(deployerBytes)

		var salt [32]byte
		copy(salt[:], saltBytes)

		got := Address(deployer, salt, initCode)

		initCodeHash := core.Keccak256(initCode)
		buf := make([]byte, 0, 1+types.AddressLength+32+types.HashLength)
		buf = append(buf, 0xff)
		buf = append(buf, deployer.Bytes()...)
		buf = append(buf, salt[:]...)
		buf = append(buf, initCodeHash.Bytes()...)
		want := types.NewAddressFromBytes(core.Keccak256(buf).Bytes()[12:])

		require.Equal(t, want, got)
		require.Len(t, got.Bytes(), types.AddressLength)
	})
}

// go test -run '^$' -fuzz '^FuzzAddressSensitiveToEveryInput$' -fuzztime=10s ./standards/eip1014
//
// Flipping any single byte of deployer, salt, or initCode must change the
// resulting address (Keccak-256 avalanche property, exercised rather than
// assumed).
func FuzzAddressSensitiveToEveryInput(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 32), []byte{0x60, 0x80}, uint8(0))
	f.Add(make([]byte, 20), make([]byte, 32), []byte{0x60, 0x80}, uint8(1))

	f.Fuzz(func(t *testing.T, deployerBytes []byte, saltBytes []byte, initCode []byte, which uint8) {
		if len(deployerBytes) == 0 {
			deployerBytes = []byte{0}
		}
		deployer := types.NewAddressFromBytes(deployerBytes)
		var salt [32]byte
		copy(salt[:], saltBytes)

		base := Address(deployer, salt, initCode)

		switch which % 3 {
		case 0:
			deployerBytes2 := append([]byte{}, deployerBytes...)
			deployerBytes2[len(deployerBytes2)-1] ^= 0x01
			got := Address(types.NewAddressFromBytes(deployerBytes2), salt, initCode)
			require.NotEqual(t, base, got)
		case 1:
			salt2 := salt
			salt2[31] ^= 0x01
			got := Address(deployer, salt2, initCode)
			require.NotEqual(t, base, got)
		case 2:
			initCode2 := append(append([]byte{}, initCode...), 0x00)
			got := Address(deployer, salt, initCode2)
			require.NotEqual(t, base, got)
		}
	})
}
