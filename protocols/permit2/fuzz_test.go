package permit2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
)

// go test -run '^$' -fuzz '^FuzzDecodeAllowanceRoundTrip$' -fuzztime=10s ./protocols/permit2
func FuzzDecodeAllowanceRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03})
	f.Add(make([]byte, 20), make([]byte, 6), make([]byte, 6))

	f.Fuzz(func(t *testing.T, amountRaw, expirationRaw, nonceRaw []byte) {
		if len(amountRaw) > 20 {
			amountRaw = amountRaw[len(amountRaw)-20:]
		}
		if len(expirationRaw) > 6 {
			expirationRaw = expirationRaw[len(expirationRaw)-6:]
		}
		if len(nonceRaw) > 6 {
			nonceRaw = nonceRaw[len(nonceRaw)-6:]
		}
		amount := new(big.Int).SetBytes(amountRaw)
		expiration := new(big.Int).SetBytes(expirationRaw)
		nonce := new(big.Int).SetBytes(nonceRaw)

		data, err := abi.Pack(abi.Types{abi.Uint160, abi.Uint48, abi.Uint48}, amount, expiration, nonce)
		require.NoError(t, err)

		got, err := DecodeAllowance(data)
		require.NoError(t, err)
		require.Zero(t, amount.Cmp(got.Amount))
		require.Zero(t, expiration.Cmp(got.Expiration))
		require.Zero(t, nonce.Cmp(got.Nonce))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeApproveRoundTrip$' -fuzztime=10s ./protocols/permit2
func FuzzEncodeApproveRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 6))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), []byte{0x01}, []byte{0x02})

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, expirationRaw []byte) {
		token := types.NewAddressFromBytes(tokenRaw)
		spender := types.NewAddressFromBytes(spenderRaw)
		if len(amountRaw) > 20 {
			amountRaw = amountRaw[len(amountRaw)-20:]
		}
		if len(expirationRaw) > 6 {
			expirationRaw = expirationRaw[len(expirationRaw)-6:]
		}
		amount := new(big.Int).SetBytes(amountRaw)
		expiration := new(big.Int).SetBytes(expirationRaw)

		data, err := EncodeApprove(token, spender, amount, expiration)
		require.NoError(t, err)
		require.Equal(t, []byte{0x87, 0x51, 0x7c, 0x45}, data[:4])

		vals, err := approveFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(token))
		require.True(t, vals[1].(*types.Address).Equal(spender))
		require.Zero(t, amount.Cmp(vals[2].(*big.Int)))
		require.Zero(t, expiration.Cmp(vals[3].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodePermitSingleRoundTrip$' -fuzztime=10s ./protocols/permit2
func FuzzEncodePermitSingleRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 6), make([]byte, 6), make([]byte, 32))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, []byte{0x06})

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, expirationRaw, nonceRaw, sigDeadlineRaw []byte) {
		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
		require.NoError(t, err)

		token := types.NewAddressFromBytes(tokenRaw)
		spender := types.NewAddressFromBytes(spenderRaw)
		if len(amountRaw) > 20 {
			amountRaw = amountRaw[len(amountRaw)-20:]
		}
		if len(expirationRaw) > 6 {
			expirationRaw = expirationRaw[len(expirationRaw)-6:]
		}
		if len(nonceRaw) > 6 {
			nonceRaw = nonceRaw[len(nonceRaw)-6:]
		}
		if len(sigDeadlineRaw) > 32 {
			sigDeadlineRaw = sigDeadlineRaw[len(sigDeadlineRaw)-32:]
		}
		amount := new(big.Int).SetBytes(amountRaw)
		expiration := new(big.Int).SetBytes(expirationRaw)
		nonce := new(big.Int).SetBytes(nonceRaw)
		sigDeadline := new(big.Int).SetBytes(sigDeadlineRaw)

		details := NewPermitDetails(token, amount, expiration, nonce)
		permitSingle := NewPermitSingle(details, spender, sigDeadline)

		data, err := EncodePermitSingle(token, permitSingle, sig)
		require.NoError(t, err)
		require.Equal(t, []byte{0x2b, 0x67, 0xb5, 0x70}, data[:4])

		vals, err := permitSingleFn.DecodeCall(data)
		require.NoError(t, err)

		got := vals[1].([]any)
		gotDetails := got[0].([]any)
		require.True(t, gotDetails[0].(*types.Address).Equal(token))
		require.Zero(t, amount.Cmp(gotDetails[1].(*big.Int)))
		require.Zero(t, expiration.Cmp(gotDetails[2].(*big.Int)))
		require.Zero(t, nonce.Cmp(gotDetails[3].(*big.Int)))
		require.True(t, got[1].(*types.Address).Equal(spender))
		require.Zero(t, sigDeadline.Cmp(got[2].(*big.Int)))
		require.Equal(t, sig.LegacyBytes(), vals[2].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeTransferFromBatchRoundTrip$' -fuzztime=10s ./protocols/permit2
func FuzzEncodeTransferFromBatchRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 20))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04})

	f.Fuzz(func(t *testing.T, fromRaw, toRaw, amountRaw, tokenRaw []byte) {
		from := types.NewAddressFromBytes(fromRaw)
		to := types.NewAddressFromBytes(toRaw)
		if len(amountRaw) > 20 {
			amountRaw = amountRaw[len(amountRaw)-20:]
		}
		amount := new(big.Int).SetBytes(amountRaw)
		token := types.NewAddressFromBytes(tokenRaw)

		details := []AllowanceTransferDetails{*NewAllowanceTransferDetails(from, to, amount, token)}

		data, err := EncodeTransferFromBatch(details)
		require.NoError(t, err)
		require.Equal(t, []byte{0x0d, 0x58, 0xb1, 0xdb}, data[:4])

		vals, err := transferFromBatchFn.DecodeCall(data)
		require.NoError(t, err)
		got := vals[0].([]any)
		require.Len(t, got, 1)
		first := got[0].([]any)
		require.True(t, first[0].(*types.Address).Equal(from))
		require.True(t, first[1].(*types.Address).Equal(to))
		require.Zero(t, amount.Cmp(first[2].(*big.Int)))
		require.True(t, first[3].(*types.Address).Equal(token))
	})
}

// go test -run '^$' -fuzz '^FuzzAllowanceTransferSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/permit2
func FuzzAllowanceTransferSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 128))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeDomainSeparator(data)
		_, _ = DecodeAllowance(data)
	})
}

// go test -run '^$' -fuzz '^FuzzPermitSingleHashDeterministic$' -fuzztime=10s ./protocols/permit2
func FuzzPermitSingleHashDeterministic(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 6), make([]byte, 6), make([]byte, 32), int64(1), make([]byte, 20))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, []byte{0x06}, int64(137), []byte{0x07})

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, expirationRaw, nonceRaw, sigDeadlineRaw []byte, chainIDRaw int64, contractRaw []byte) {
		if len(amountRaw) > 20 {
			amountRaw = amountRaw[len(amountRaw)-20:]
		}
		if len(expirationRaw) > 6 {
			expirationRaw = expirationRaw[len(expirationRaw)-6:]
		}
		if len(nonceRaw) > 6 {
			nonceRaw = nonceRaw[len(nonceRaw)-6:]
		}
		if len(sigDeadlineRaw) > 32 {
			sigDeadlineRaw = sigDeadlineRaw[len(sigDeadlineRaw)-32:]
		}
		if chainIDRaw < 0 {
			chainIDRaw = -chainIDRaw
		}

		details := NewPermitDetails(
			types.NewAddressFromBytes(tokenRaw),
			new(big.Int).SetBytes(amountRaw),
			new(big.Int).SetBytes(expirationRaw),
			new(big.Int).SetBytes(nonceRaw),
		)
		p := NewPermitSingle(details, types.NewAddressFromBytes(spenderRaw), new(big.Int).SetBytes(sigDeadlineRaw))
		domain := NewDomain(big.NewInt(chainIDRaw), types.NewAddressFromBytes(contractRaw))

		h1, err := p.Hash(domain)
		require.NoError(t, err)
		require.Len(t, h1.Bytes(), 32)

		h2, err := p.Hash(domain)
		require.NoError(t, err)
		require.Equal(t, h1.Bytes(), h2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzPermitBatchHashDeterministic$' -fuzztime=10s ./protocols/permit2
func FuzzPermitBatchHashDeterministic(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 6), make([]byte, 6), make([]byte, 32), int64(1), make([]byte, 20))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, []byte{0x06}, int64(137), []byte{0x07})

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, expirationRaw, nonceRaw, sigDeadlineRaw []byte, chainIDRaw int64, contractRaw []byte) {
		if len(amountRaw) > 20 {
			amountRaw = amountRaw[len(amountRaw)-20:]
		}
		if len(expirationRaw) > 6 {
			expirationRaw = expirationRaw[len(expirationRaw)-6:]
		}
		if len(nonceRaw) > 6 {
			nonceRaw = nonceRaw[len(nonceRaw)-6:]
		}
		if len(sigDeadlineRaw) > 32 {
			sigDeadlineRaw = sigDeadlineRaw[len(sigDeadlineRaw)-32:]
		}
		if chainIDRaw < 0 {
			chainIDRaw = -chainIDRaw
		}

		details := []PermitDetails{*NewPermitDetails(
			types.NewAddressFromBytes(tokenRaw),
			new(big.Int).SetBytes(amountRaw),
			new(big.Int).SetBytes(expirationRaw),
			new(big.Int).SetBytes(nonceRaw),
		)}
		p := NewPermitBatch(details, types.NewAddressFromBytes(spenderRaw), new(big.Int).SetBytes(sigDeadlineRaw))
		domain := NewDomain(big.NewInt(chainIDRaw), types.NewAddressFromBytes(contractRaw))

		h1, err := p.Hash(domain)
		require.NoError(t, err)
		require.Len(t, h1.Bytes(), 32)

		h2, err := p.Hash(domain)
		require.NoError(t, err)
		require.Equal(t, h1.Bytes(), h2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzPermitSingleSignVerifies$' -fuzztime=10s ./protocols/permit2
func FuzzPermitSingleSignVerifies(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 6), make([]byte, 6), make([]byte, 32), int64(1))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, []byte{0x06}, int64(137))

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, expirationRaw, nonceRaw, sigDeadlineRaw []byte, chainIDRaw int64) {
		if len(amountRaw) > 20 {
			amountRaw = amountRaw[len(amountRaw)-20:]
		}
		if len(expirationRaw) > 6 {
			expirationRaw = expirationRaw[len(expirationRaw)-6:]
		}
		if len(nonceRaw) > 6 {
			nonceRaw = nonceRaw[len(nonceRaw)-6:]
		}
		if len(sigDeadlineRaw) > 32 {
			sigDeadlineRaw = sigDeadlineRaw[len(sigDeadlineRaw)-32:]
		}
		if chainIDRaw < 0 {
			chainIDRaw = -chainIDRaw
		}

		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		signerAddr := core.PubkeyToAddress(key.PublicKey())

		details := NewPermitDetails(
			types.NewAddressFromBytes(tokenRaw),
			new(big.Int).SetBytes(amountRaw),
			new(big.Int).SetBytes(expirationRaw),
			new(big.Int).SetBytes(nonceRaw),
		)
		p := NewPermitSingle(details, types.NewAddressFromBytes(spenderRaw), new(big.Int).SetBytes(sigDeadlineRaw))
		domain := NewDomain(big.NewInt(chainIDRaw), types.NewAddressFromBytes(tokenRaw))

		sig, err := p.Sign(domain, key)
		require.NoError(t, err)

		digest, err := p.Hash(domain)
		require.NoError(t, err)

		ok, err := core.VerifyAddress(digest, sig, signerAddr)
		require.NoError(t, err)
		require.True(t, ok)
	})
}

// go test -run '^$' -fuzz '^FuzzExtractEventsDoNotPanic$' -fuzztime=10s ./protocols/permit2
func FuzzExtractEventsDoNotPanic(f *testing.F) {
	f.Add(uint8(0), []byte{}, []byte{})
	f.Add(uint8(4), make([]byte, 128), make([]byte, 64))
	f.Add(uint8(7), []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0x01})

	f.Fuzz(func(t *testing.T, topicCount uint8, topicsRaw, data []byte) {
		n := int(topicCount % 8)
		topics := make([]*types.Hash, n)
		for i := range topics {
			start := i * 32
			if start >= len(topicsRaw) {
				topics[i] = types.NewHashFromBytes(nil)
				continue
			}
			end := min(start+32, len(topicsRaw))
			topics[i] = types.NewHashFromBytes(topicsRaw[start:end])
		}
		logs := []tx.Log{{Address: testPermit2, Topics: topics, Data: data}}

		_, _ = ExtractNonceInvalidations(logs)
		_, _ = ExtractApprovals(logs)
		_, _ = ExtractPermits(logs)
		_, _ = ExtractLockdowns(logs)
		_, _ = ExtractUnorderedNonceInvalidations(logs)
	})
}

// go test -run '^$' -fuzz '^FuzzPermitTransferFromHashDeterministic$' -fuzztime=10s ./protocols/permit2
func FuzzPermitTransferFromHashDeterministic(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 32), int64(1), make([]byte, 20))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, int64(137), []byte{0x06})

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, nonceRaw, deadlineRaw []byte, chainIDRaw int64, contractRaw []byte) {
		if chainIDRaw < 0 {
			chainIDRaw = -chainIDRaw
		}
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		if len(nonceRaw) > 32 {
			nonceRaw = nonceRaw[len(nonceRaw)-32:]
		}
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}

		permitted := NewTokenPermissions(types.NewAddressFromBytes(tokenRaw), new(big.Int).SetBytes(amountRaw))
		p := NewPermitTransferFrom(permitted, new(big.Int).SetBytes(nonceRaw), new(big.Int).SetBytes(deadlineRaw))
		spender := types.NewAddressFromBytes(spenderRaw)
		domain := NewDomain(big.NewInt(chainIDRaw), types.NewAddressFromBytes(contractRaw))

		h1, err := p.Hash(domain, spender)
		require.NoError(t, err)
		require.Len(t, h1.Bytes(), 32)

		h2, err := p.Hash(domain, spender)
		require.NoError(t, err)
		require.Equal(t, h1.Bytes(), h2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzPermitBatchTransferFromHashDeterministic$' -fuzztime=10s ./protocols/permit2
func FuzzPermitBatchTransferFromHashDeterministic(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 32), int64(1), make([]byte, 20))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, int64(137), []byte{0x06})

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, nonceRaw, deadlineRaw []byte, chainIDRaw int64, contractRaw []byte) {
		if chainIDRaw < 0 {
			chainIDRaw = -chainIDRaw
		}
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		if len(nonceRaw) > 32 {
			nonceRaw = nonceRaw[len(nonceRaw)-32:]
		}
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}

		permitted := []TokenPermissions{*NewTokenPermissions(types.NewAddressFromBytes(tokenRaw), new(big.Int).SetBytes(amountRaw))}
		p := NewPermitBatchTransferFrom(permitted, new(big.Int).SetBytes(nonceRaw), new(big.Int).SetBytes(deadlineRaw))
		spender := types.NewAddressFromBytes(spenderRaw)
		domain := NewDomain(big.NewInt(chainIDRaw), types.NewAddressFromBytes(contractRaw))

		h1, err := p.Hash(domain, spender)
		require.NoError(t, err)
		require.Len(t, h1.Bytes(), 32)

		h2, err := p.Hash(domain, spender)
		require.NoError(t, err)
		require.Equal(t, h1.Bytes(), h2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzPermitTransferFromSignVerifies$' -fuzztime=10s ./protocols/permit2
func FuzzPermitTransferFromSignVerifies(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 32), int64(1))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, int64(137))

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, nonceRaw, deadlineRaw []byte, chainIDRaw int64) {
		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		signerAddr := core.PubkeyToAddress(key.PublicKey())

		if chainIDRaw < 0 {
			chainIDRaw = -chainIDRaw
		}
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		if len(nonceRaw) > 32 {
			nonceRaw = nonceRaw[len(nonceRaw)-32:]
		}
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}

		permitted := NewTokenPermissions(types.NewAddressFromBytes(tokenRaw), new(big.Int).SetBytes(amountRaw))
		p := NewPermitTransferFrom(permitted, new(big.Int).SetBytes(nonceRaw), new(big.Int).SetBytes(deadlineRaw))
		spender := types.NewAddressFromBytes(spenderRaw)
		domain := NewDomain(big.NewInt(chainIDRaw), types.NewAddressFromBytes(tokenRaw))

		sig, err := p.Sign(domain, spender, key)
		require.NoError(t, err)

		digest, err := p.Hash(domain, spender)
		require.NoError(t, err)

		ok, err := core.VerifyAddress(digest, sig, signerAddr)
		require.NoError(t, err)
		require.True(t, ok)
	})
}

// go test -run '^$' -fuzz '^FuzzPermitTransferFromHashWithWitnessSignVerifies$' -fuzztime=10s ./protocols/permit2
func FuzzPermitTransferFromHashWithWitnessSignVerifies(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 32), make([]byte, 32), "Witness witness)Witness(uint256 x)", int64(1))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, []byte{0x06}, "", int64(137))

	f.Fuzz(func(t *testing.T, tokenRaw, spenderRaw, amountRaw, nonceRaw, deadlineRaw, witnessRaw []byte, witnessTypeString string, chainIDRaw int64) {
		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		signerAddr := core.PubkeyToAddress(key.PublicKey())

		if chainIDRaw < 0 {
			chainIDRaw = -chainIDRaw
		}
		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		if len(nonceRaw) > 32 {
			nonceRaw = nonceRaw[len(nonceRaw)-32:]
		}
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}

		permitted := NewTokenPermissions(types.NewAddressFromBytes(tokenRaw), new(big.Int).SetBytes(amountRaw))
		p := NewPermitTransferFrom(permitted, new(big.Int).SetBytes(nonceRaw), new(big.Int).SetBytes(deadlineRaw))
		spender := types.NewAddressFromBytes(spenderRaw)
		witness := types.NewHashFromBytes(witnessRaw)
		domain := NewDomain(big.NewInt(chainIDRaw), types.NewAddressFromBytes(tokenRaw))

		h1, err := p.HashWithWitness(domain, spender, witness, witnessTypeString)
		require.NoError(t, err)
		require.Len(t, h1.Bytes(), 32)
		h2, err := p.HashWithWitness(domain, spender, witness, witnessTypeString)
		require.NoError(t, err)
		require.Equal(t, h1.Bytes(), h2.Bytes())

		sig, err := p.SignWithWitness(domain, spender, witness, witnessTypeString, key)
		require.NoError(t, err)

		ok, err := core.VerifyAddress(h1, sig, signerAddr)
		require.NoError(t, err)
		require.True(t, ok)
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeNonceBitmapRoundTrip$' -fuzztime=10s ./protocols/permit2
func FuzzEncodeNonceBitmapRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 32))
	f.Add([]byte{0x01}, []byte{0x02})

	f.Fuzz(func(t *testing.T, ownerRaw, wordPosRaw []byte) {
		if len(wordPosRaw) > 32 {
			wordPosRaw = wordPosRaw[len(wordPosRaw)-32:]
		}
		owner := types.NewAddressFromBytes(ownerRaw)
		wordPos := new(big.Int).SetBytes(wordPosRaw)

		data, err := EncodeNonceBitmap(owner, wordPos)
		require.NoError(t, err)
		require.Equal(t, []byte{0x4f, 0xe0, 0x2b, 0x44}, data[:4])

		vals, err := nonceBitmapFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(owner))
		require.Zero(t, wordPos.Cmp(vals[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodePermitTransferFromRoundTrip$' -fuzztime=10s ./protocols/permit2
func FuzzEncodePermitTransferFromRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), make([]byte, 32), make([]byte, 32), make([]byte, 32), make([]byte, 20), make([]byte, 32))
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, []byte{0x06}, []byte{0x07})

	f.Fuzz(func(t *testing.T, tokenRaw, ownerRaw, amountRaw, nonceRaw, deadlineRaw, toRaw, requestedAmountRaw []byte) {
		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
		require.NoError(t, err)

		if len(amountRaw) > 32 {
			amountRaw = amountRaw[len(amountRaw)-32:]
		}
		if len(nonceRaw) > 32 {
			nonceRaw = nonceRaw[len(nonceRaw)-32:]
		}
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}
		if len(requestedAmountRaw) > 32 {
			requestedAmountRaw = requestedAmountRaw[len(requestedAmountRaw)-32:]
		}

		permitted := NewTokenPermissions(types.NewAddressFromBytes(tokenRaw), new(big.Int).SetBytes(amountRaw))
		permit := NewPermitTransferFrom(permitted, new(big.Int).SetBytes(nonceRaw), new(big.Int).SetBytes(deadlineRaw))
		details := NewSignatureTransferDetails(types.NewAddressFromBytes(toRaw), new(big.Int).SetBytes(requestedAmountRaw))
		owner := types.NewAddressFromBytes(ownerRaw)

		data, err := EncodePermitTransferFrom(permit, details, owner, sig)
		require.NoError(t, err)
		require.Equal(t, []byte{0x30, 0xf2, 0x8b, 0x7a}, data[:4])

		vals, err := permitTransferFromFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[2].(*types.Address).Equal(owner))
		require.Equal(t, sig.LegacyBytes(), vals[3].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzSignatureTransferSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/permit2
func FuzzSignatureTransferSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeNonceBitmap(data)
	})
}
