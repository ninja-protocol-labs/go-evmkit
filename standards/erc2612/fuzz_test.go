package erc2612

import (
	"math/big"
	"testing"

	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/standards/eip712"
)

// go test -run '^$' -fuzz '^FuzzPermitHashCrashSafety$' -fuzztime=10s ./standards/erc2612
func FuzzPermitHashCrashSafety(f *testing.F) {
	f.Add(
		make([]byte, 20), make([]byte, 20), make([]byte, 20),
		[]byte{}, []byte{}, []byte{},
		"evmkit test token", DefaultVersion, []byte{1},
	)
	f.Add(
		[]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), make([]byte, 20),
		make([]byte, 32), make([]byte, 32), make([]byte, 32),
		"", "", make([]byte, 32),
	)

	f.Fuzz(func(t *testing.T, ownerRaw, spenderRaw, contractRaw, valueRaw, nonceRaw, deadlineRaw []byte, name, version string, chainIDRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		spender := types.NewAddressFromBytes(spenderRaw)
		contract := types.NewAddressFromBytes(contractRaw)
		if len(valueRaw) > 32 {
			valueRaw = valueRaw[len(valueRaw)-32:]
		}
		value := new(big.Int).SetBytes(valueRaw)
		if len(nonceRaw) > 32 {
			nonceRaw = nonceRaw[len(nonceRaw)-32:]
		}
		nonce := new(big.Int).SetBytes(nonceRaw)
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}
		deadline := new(big.Int).SetBytes(deadlineRaw)
		if len(chainIDRaw) > 32 {
			chainIDRaw = chainIDRaw[len(chainIDRaw)-32:]
		}
		chainID := new(big.Int).SetBytes(chainIDRaw)
		if name == "" {
			name = "x"
		}
		if version == "" {
			version = DefaultVersion
		}

		p := NewPermit(owner, spender, value, nonce, deadline)
		domain := eip712.NewDomain(name, version, chainID, contract, nil)

		got, err := p.Hash(domain)
		require.NoError(t, err)
		require.Len(t, got.Bytes(), types.HashLength)

		got2, err := p.Hash(domain)
		require.NoError(t, err)
		require.Equal(t, got.Bytes(), got2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzPermitHashMatchesManualEip712Hash$' -fuzztime=10s ./standards/erc2612
func FuzzPermitHashMatchesManualEip712Hash(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{1}, []byte{}, []byte{}, "evmkit", DefaultVersion, []byte{1})

	f.Fuzz(func(t *testing.T, ownerRaw, spenderRaw, contractRaw, valueRaw, nonceRaw, deadlineRaw []byte, name, version string, chainIDRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		spender := types.NewAddressFromBytes(spenderRaw)
		contract := types.NewAddressFromBytes(contractRaw)
		if len(valueRaw) > 32 {
			valueRaw = valueRaw[len(valueRaw)-32:]
		}
		value := new(big.Int).SetBytes(valueRaw)
		if len(nonceRaw) > 32 {
			nonceRaw = nonceRaw[len(nonceRaw)-32:]
		}
		nonce := new(big.Int).SetBytes(nonceRaw)
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}
		deadline := new(big.Int).SetBytes(deadlineRaw)
		if len(chainIDRaw) > 32 {
			chainIDRaw = chainIDRaw[len(chainIDRaw)-32:]
		}
		chainID := new(big.Int).SetBytes(chainIDRaw)
		if name == "" {
			name = "x"
		}
		if version == "" {
			version = DefaultVersion
		}

		p := NewPermit(owner, spender, value, nonce, deadline)
		domain := eip712.NewDomain(name, version, chainID, contract, nil)

		got, err := p.Hash(domain)
		require.NoError(t, err)

		want, err := eip712.Hash(domain, primaryType, eip712.NewTypes(primaryType,
			eip712.NewField(ownerName, abi.Address.String()),
			eip712.NewField(spenderName, abi.Address.String()),
			eip712.NewField(valueName, abi.Uint256.String()),
			eip712.NewField(nonceName, abi.Uint256.String()),
			eip712.NewField(deadlineName, abi.Uint256.String()),
		), map[string]any{
			ownerName:    owner,
			spenderName:  spender,
			valueName:    value,
			nonceName:    nonce,
			deadlineName: deadline,
		})
		require.NoError(t, err)
		require.Equal(t, want.Bytes(), got.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzPermitHashSensitiveToEachField$' -fuzztime=10s ./standards/erc2612
func FuzzPermitHashSensitiveToEachField(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), uint64(100), uint64(0), uint64(9999999999), uint8(0))
	f.Add(make([]byte, 20), make([]byte, 20), uint64(100), uint64(0), uint64(9999999999), uint8(4))

	contract := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	domain := eip712.NewDomain("evmkit test token", DefaultVersion, big.NewInt(1), contract, nil)

	f.Fuzz(func(t *testing.T, ownerRaw, spenderRaw []byte, value, nonce, deadline uint64, which uint8) {
		owner := types.NewAddressFromBytes(ownerRaw)
		spender := types.NewAddressFromBytes(spenderRaw)
		if owner.Equal(spender) {
			spender = types.NewAddressFromBytes(append(append([]byte{}, spenderRaw...), 0x01))
		}

		base := NewPermit(owner, spender, new(big.Int).SetUint64(value), new(big.Int).SetUint64(nonce), new(big.Int).SetUint64(deadline))
		baseHash, err := base.Hash(domain)
		require.NoError(t, err)

		var variant *Permit
		switch which % 5 {
		case 0:
			variant = NewPermit(spender, spender, base.Value, base.Nonce, base.Deadline)
		case 1:
			variant = NewPermit(owner, owner, base.Value, base.Nonce, base.Deadline)
		case 2:
			variant = NewPermit(owner, spender, new(big.Int).Add(base.Value, big.NewInt(1)), base.Nonce, base.Deadline)
		case 3:
			variant = NewPermit(owner, spender, base.Value, new(big.Int).Add(base.Nonce, big.NewInt(1)), base.Deadline)
		case 4:
			variant = NewPermit(owner, spender, base.Value, base.Nonce, new(big.Int).Add(base.Deadline, big.NewInt(1)))
		}

		variantHash, err := variant.Hash(domain)
		require.NoError(t, err)
		require.NotEqual(t, baseHash.Bytes(), variantHash.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzEncodePermitRoundTrip$' -fuzztime=10s ./standards/erc2612
func FuzzEncodePermitRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{}, []byte{}, []byte("digest"))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), make([]byte, 32), make([]byte, 32), []byte{1, 2, 3})

	f.Fuzz(func(t *testing.T, ownerRaw, spenderRaw, valueRaw, deadlineRaw, digestRaw []byte) {
		owner := types.NewAddressFromBytes(ownerRaw)
		spender := types.NewAddressFromBytes(spenderRaw)
		if len(valueRaw) > 32 {
			valueRaw = valueRaw[len(valueRaw)-32:]
		}
		value := new(big.Int).SetBytes(valueRaw)
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}
		deadline := new(big.Int).SetBytes(deadlineRaw)

		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		sig, err := key.Sign(types.NewHashFromBytes(digestRaw))
		require.NoError(t, err)

		data, err := EncodePermit(owner, spender, value, deadline, sig)
		require.NoError(t, err)
		require.Equal(t, []byte{0xd5, 0x05, 0xac, 0xcf}, data[:4])

		vals, err := permitFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(owner))
		require.True(t, vals[1].(*types.Address).Equal(spender))
		require.Zero(t, value.Cmp(vals[2].(*big.Int)))
		require.Zero(t, deadline.Cmp(vals[3].(*big.Int)))

		wantV := uint8(sig.LegacyV().Uint64())
		require.Equal(t, wantV, vals[4].(uint8))
		require.True(t, wantV == 27 || wantV == 28)

		wantR := make([]byte, 32)
		sig.R().FillBytes(wantR)
		require.Equal(t, wantR, vals[5].([]byte))

		wantS := make([]byte, 32)
		sig.S().FillBytes(wantS)
		require.Equal(t, wantS, vals[6].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeNoncesRoundTrip$' -fuzztime=10s ./standards/erc2612
func FuzzDecodeNoncesRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add([]byte{})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, raw []byte) {
		if len(raw) > 32 {
			raw = raw[len(raw)-32:]
		}
		want := new(big.Int).SetBytes(raw)

		data, err := abi.Pack(abi.Types{abi.Uint256}, want)
		require.NoError(t, err)

		got, err := DecodeNonces(data)
		require.NoError(t, err)
		require.Zero(t, want.Cmp(got))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeDomainSeparatorRoundTrip$' -fuzztime=10s ./standards/erc2612
func FuzzDecodeDomainSeparatorRoundTrip(f *testing.F) {
	f.Add(make([]byte, 32))
	f.Add([]byte{0xff})

	f.Fuzz(func(t *testing.T, raw []byte) {
		want := make([]byte, 32)
		if len(raw) > 32 {
			raw = raw[:32]
		}
		copy(want[32-len(raw):], raw)

		data, err := abi.Pack(abi.Types{abi.Bytes32}, want)
		require.NoError(t, err)

		got, err := DecodeDomainSeparator(data)
		require.NoError(t, err)
		require.Equal(t, want, got)
	})
}

// go test -run '^$' -fuzz '^FuzzSimpleDecodersDoNotPanic$' -fuzztime=10s ./standards/erc2612
func FuzzSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeNonces(data)
		_, _ = DecodeDomainSeparator(data)
	})
}
