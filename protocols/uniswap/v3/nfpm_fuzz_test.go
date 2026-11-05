package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

// go test -run '^$' -fuzz '^FuzzEncodeNfpmMintRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeNfpmMintRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x01, 0xf4}, []byte{0x00, 0x64}, []byte{0xff, 0x9c}, []byte{0x00}, []byte{0x00}, make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xbe, 0xef}, []byte{0xff, 0xff, 0xff}, []byte{0x00}, []byte{0x00}, []byte{0xff}, []byte{0xff}, []byte{0xaa})

	f.Fuzz(func(t *testing.T, token0Raw, token1Raw, feeRaw, tickLowerRaw, tickUpperRaw, amount0Raw, amount1Raw, recipientRaw []byte) {
		token0 := types.NewAddressFromBytes(token0Raw)
		token1 := types.NewAddressFromBytes(token1Raw)
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		tickLower := new(big.Int).Mod(new(big.Int).SetBytes(tickLowerRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickLower.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickLower.Sub(tickLower, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		tickUpper := new(big.Int).Mod(new(big.Int).SetBytes(tickUpperRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickUpper.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickUpper.Sub(tickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		if len(amount0Raw) > 32 {
			amount0Raw = amount0Raw[len(amount0Raw)-32:]
		}
		amount0Desired := new(big.Int).SetBytes(amount0Raw)
		if len(amount1Raw) > 32 {
			amount1Raw = amount1Raw[len(amount1Raw)-32:]
		}
		amount1Desired := new(big.Int).SetBytes(amount1Raw)
		recipient := types.NewAddressFromBytes(recipientRaw)

		data, err := EncodeNfpmMint(token0, token1, fee, tickLower, tickUpper, amount0Desired, amount1Desired, big.NewInt(0), big.NewInt(0), recipient, big.NewInt(1700000000))
		require.NoError(t, err)

		vals, err := nfpmMintFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.True(t, fields[0].(*types.Address).Equal(token0))
		require.True(t, fields[1].(*types.Address).Equal(token1))
		require.Zero(t, fee.Cmp(fields[2].(*big.Int)))
		require.Zero(t, tickLower.Cmp(fields[3].(*big.Int)))
		require.Zero(t, tickUpper.Cmp(fields[4].(*big.Int)))
		require.Zero(t, amount0Desired.Cmp(fields[5].(*big.Int)))
		require.Zero(t, amount1Desired.Cmp(fields[6].(*big.Int)))
		require.True(t, fields[9].(*types.Address).Equal(recipient))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeIncreaseLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeIncreaseLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x00}, []byte{0x00})
	f.Add([]byte{0xff}, []byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, tokenIdRaw, amount0Raw, amount1Raw []byte) {
		if len(tokenIdRaw) > 32 {
			tokenIdRaw = tokenIdRaw[len(tokenIdRaw)-32:]
		}
		tokenId := new(big.Int).SetBytes(tokenIdRaw)
		if len(amount0Raw) > 32 {
			amount0Raw = amount0Raw[len(amount0Raw)-32:]
		}
		amount0Desired := new(big.Int).SetBytes(amount0Raw)
		if len(amount1Raw) > 32 {
			amount1Raw = amount1Raw[len(amount1Raw)-32:]
		}
		amount1Desired := new(big.Int).SetBytes(amount1Raw)

		data, err := EncodeIncreaseLiquidity(tokenId, amount0Desired, amount1Desired, big.NewInt(0), big.NewInt(0), big.NewInt(1700000000))
		require.NoError(t, err)

		vals, err := increaseLiquidityFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.Zero(t, tokenId.Cmp(fields[0].(*big.Int)))
		require.Zero(t, amount0Desired.Cmp(fields[1].(*big.Int)))
		require.Zero(t, amount1Desired.Cmp(fields[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeDecreaseLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeDecreaseLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x02})
	f.Add([]byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, tokenIdRaw, liquidityRaw []byte) {
		if len(tokenIdRaw) > 32 {
			tokenIdRaw = tokenIdRaw[len(tokenIdRaw)-32:]
		}
		tokenId := new(big.Int).SetBytes(tokenIdRaw)
		if len(liquidityRaw) > 16 {
			liquidityRaw = liquidityRaw[len(liquidityRaw)-16:]
		}
		liquidity := new(big.Int).SetBytes(liquidityRaw)

		data, err := EncodeDecreaseLiquidity(tokenId, liquidity, big.NewInt(0), big.NewInt(0), big.NewInt(1700000000))
		require.NoError(t, err)

		vals, err := decreaseLiquidityFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.Zero(t, tokenId.Cmp(fields[0].(*big.Int)))
		require.Zero(t, liquidity.Cmp(fields[1].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeNfpmCollectRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeNfpmCollectRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, make([]byte, 20), []byte{0x02}, []byte{0x03})
	f.Add([]byte{0xff}, []byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, tokenIdRaw, recipientRaw, amount0MaxRaw, amount1MaxRaw []byte) {
		if len(tokenIdRaw) > 32 {
			tokenIdRaw = tokenIdRaw[len(tokenIdRaw)-32:]
		}
		tokenId := new(big.Int).SetBytes(tokenIdRaw)
		recipient := types.NewAddressFromBytes(recipientRaw)
		if len(amount0MaxRaw) > 16 {
			amount0MaxRaw = amount0MaxRaw[len(amount0MaxRaw)-16:]
		}
		amount0Max := new(big.Int).SetBytes(amount0MaxRaw)
		if len(amount1MaxRaw) > 16 {
			amount1MaxRaw = amount1MaxRaw[len(amount1MaxRaw)-16:]
		}
		amount1Max := new(big.Int).SetBytes(amount1MaxRaw)

		data, err := EncodeNfpmCollect(tokenId, recipient, amount0Max, amount1Max)
		require.NoError(t, err)

		vals, err := nfpmCollectFn.DecodeCall(data)
		require.NoError(t, err)
		fields := vals[0].([]any)
		require.Zero(t, tokenId.Cmp(fields[0].(*big.Int)))
		require.True(t, fields[1].(*types.Address).Equal(recipient))
		require.Zero(t, amount0Max.Cmp(fields[2].(*big.Int)))
		require.Zero(t, amount1Max.Cmp(fields[3].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeNfpmBurnRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeNfpmBurnRoundTrip(f *testing.F) {
	f.Add([]byte{0x01})
	f.Add(make([]byte, 32))

	f.Fuzz(func(t *testing.T, tokenIdRaw []byte) {
		if len(tokenIdRaw) > 32 {
			tokenIdRaw = tokenIdRaw[len(tokenIdRaw)-32:]
		}
		tokenId := new(big.Int).SetBytes(tokenIdRaw)

		data, err := EncodeNfpmBurn(tokenId)
		require.NoError(t, err)

		vals, err := nfpmBurnFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, tokenId.Cmp(vals[0].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeCreateAndInitializePoolIfNecessaryRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeCreateAndInitializePoolIfNecessaryRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), make([]byte, 20), []byte{0x01, 0xf4}, []byte{0x00})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xbe, 0xef}, []byte{0xff, 0xff, 0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, token0Raw, token1Raw, feeRaw, sqrtPriceX96Raw []byte) {
		token0 := types.NewAddressFromBytes(token0Raw)
		token1 := types.NewAddressFromBytes(token1Raw)
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if len(sqrtPriceX96Raw) > 20 {
			sqrtPriceX96Raw = sqrtPriceX96Raw[len(sqrtPriceX96Raw)-20:]
		}
		sqrtPriceX96 := new(big.Int).SetBytes(sqrtPriceX96Raw)

		data, err := EncodeCreateAndInitializePoolIfNecessary(token0, token1, fee, sqrtPriceX96)
		require.NoError(t, err)

		vals, err := createAndInitializePoolIfNecessaryFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(token0))
		require.True(t, vals[1].(*types.Address).Equal(token1))
		require.Zero(t, fee.Cmp(vals[2].(*big.Int)))
		require.Zero(t, sqrtPriceX96.Cmp(vals[3].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeNfpmMulticallRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeNfpmMulticallRoundTrip(f *testing.F) {
	f.Add([]byte{0xde, 0xad}, []byte{0xbe, 0xef})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, a, b []byte) {
		calls := [][]byte{a, b}

		data, err := EncodeMulticall(calls)
		require.NoError(t, err)

		vals, err := nfpmMulticallFn.DecodeCall(data)
		require.NoError(t, err)
		elems := vals[0].([]any)
		require.Len(t, elems, 2)
		require.Equal(t, a, elems[0].([]byte))
		require.Equal(t, b, elems[1].([]byte))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeNfpmMulticallRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzDecodeNfpmMulticallRoundTrip(f *testing.F) {
	f.Add([]byte{0xde, 0xad}, []byte{0xbe, 0xef})
	f.Add([]byte{}, []byte{})

	f.Fuzz(func(t *testing.T, a, b []byte) {
		results := [][]byte{a, b}

		data, err := nfpmMulticallFn.EncodeReturn(results)
		require.NoError(t, err)

		got, err := DecodeMulticall(data)
		require.NoError(t, err)
		require.Len(t, got, 2)
		require.Equal(t, a, got[0])
		require.Equal(t, b, got[1])
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeNFTPermitRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzEncodeNFTPermitRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x01}, []byte{0x00})
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xff}, []byte{0xff})

	f.Fuzz(func(t *testing.T, spenderRaw, tokenIdRaw, deadlineRaw []byte) {
		spender := types.NewAddressFromBytes(spenderRaw)
		if len(tokenIdRaw) > 32 {
			tokenIdRaw = tokenIdRaw[len(tokenIdRaw)-32:]
		}
		tokenId := new(big.Int).SetBytes(tokenIdRaw)
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}
		deadline := new(big.Int).SetBytes(deadlineRaw)
		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		sig, err := key.Sign(types.NewHashFromBytes([]byte("some 32 byte digest padded out!")))
		require.NoError(t, err)

		data, err := EncodeNFTPermit(spender, tokenId, deadline, sig)
		require.NoError(t, err)

		vals, err := nftPermitFn.DecodeCall(data)
		require.NoError(t, err)
		require.True(t, vals[0].(*types.Address).Equal(spender))
		require.Zero(t, tokenId.Cmp(vals[1].(*big.Int)))
		require.Zero(t, deadline.Cmp(vals[2].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeNfpmPositionsRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzDecodeNfpmPositionsRoundTrip(f *testing.F) {
	f.Add(make([]byte, 12), make([]byte, 20), make([]byte, 20), make([]byte, 20), []byte{0x01, 0xf4}, []byte{0x00, 0x64}, []byte{0xff, 0x9c})
	f.Add([]byte{0xaa}, []byte{0xbb}, []byte{0xcc}, []byte{0xdd}, []byte{0xff, 0xff, 0xff}, []byte{0x00}, []byte{0xff})

	f.Fuzz(func(t *testing.T, nonceRaw, operatorRaw, token0Raw, token1Raw, feeRaw, tickLowerRaw, tickUpperRaw []byte) {
		if len(nonceRaw) > 12 {
			nonceRaw = nonceRaw[len(nonceRaw)-12:]
		}
		nonce := new(big.Int).SetBytes(nonceRaw)
		operator := types.NewAddressFromBytes(operatorRaw)
		token0 := types.NewAddressFromBytes(token0Raw)
		token1 := types.NewAddressFromBytes(token1Raw)
		fee := new(big.Int).Mod(new(big.Int).SetBytes(feeRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		tickLower := new(big.Int).Mod(new(big.Int).SetBytes(tickLowerRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickLower.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickLower.Sub(tickLower, new(big.Int).Lsh(big.NewInt(1), 24))
		}
		tickUpper := new(big.Int).Mod(new(big.Int).SetBytes(tickUpperRaw), new(big.Int).Lsh(big.NewInt(1), 24))
		if tickUpper.Cmp(new(big.Int).Lsh(big.NewInt(1), 23)) >= 0 {
			tickUpper.Sub(tickUpper, new(big.Int).Lsh(big.NewInt(1), 24))
		}

		data, err := nfpmPositionsFn.EncodeReturn(nonce, operator, token0, token1, fee, tickLower, tickUpper, big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0), big.NewInt(0))
		require.NoError(t, err)

		got, err := DecodeNfpmPositions(data)
		require.NoError(t, err)
		require.Zero(t, nonce.Cmp(got.Nonce))
		require.True(t, got.Operator.Equal(operator))
		require.True(t, got.Token0.Equal(token0))
		require.True(t, got.Token1.Equal(token1))
		require.Zero(t, fee.Cmp(got.Fee))
		require.Zero(t, tickLower.Cmp(got.TickLower))
		require.Zero(t, tickUpper.Cmp(got.TickUpper))
	})
}

// go test -run '^$' -fuzz '^FuzzNfpmSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzNfpmSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte{0x00})
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeNfpmFactory(data)
		_, _ = DecodeWETH9(data)
		_, _ = DecodeCreateAndInitializePoolIfNecessary(data)
		_, _ = DecodeNfpmPositions(data)
		_, _, _, _, _ = DecodeNfpmMint(data)
		_, _, _, _ = DecodeIncreaseLiquidity(data)
		_, _, _ = DecodeDecreaseLiquidity(data)
		_, _, _ = DecodeNfpmCollect(data)
		_, _ = DecodeMulticall(data)
		_, _ = DecodeTotalSupply(data)
		_, _ = DecodeTokenOfOwnerByIndex(data)
		_, _ = DecodeTokenByIndex(data)
		_, _ = DecodePermitTypehash(data)
		_, _ = DecodeDomainSeparator(data)
	})
}

// go test -run '^$' -fuzz '^FuzzNFTPermitHashDeterministic$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzNFTPermitHashDeterministic(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x01}, []byte{0x00}, []byte{0xff}, []byte{0x01}, make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xff}, []byte{0xff}, []byte{0xff}, []byte{0xff}, []byte{0xbe, 0xef})

	f.Fuzz(func(t *testing.T, spenderRaw, tokenIdRaw, nonceRaw, deadlineRaw, chainIDRaw, verifyingContractRaw []byte) {
		spender := types.NewAddressFromBytes(spenderRaw)
		if len(tokenIdRaw) > 32 {
			tokenIdRaw = tokenIdRaw[len(tokenIdRaw)-32:]
		}
		tokenId := new(big.Int).SetBytes(tokenIdRaw)
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
		if chainID.Sign() == 0 {
			chainID = big.NewInt(1)
		}
		verifyingContract := types.NewAddressFromBytes(verifyingContractRaw)

		domain := NewNFTPermitDomain(chainID, verifyingContract)
		p := NewNFTPermit(spender, tokenId, nonce, deadline)

		h1, err := p.Hash(domain)
		require.NoError(t, err)
		require.Len(t, h1.Bytes(), 32)

		h2, err := p.Hash(domain)
		require.NoError(t, err)
		require.Equal(t, h1.Bytes(), h2.Bytes())
	})
}

// go test -run '^$' -fuzz '^FuzzNFTPermitSignRoundTrip$' -fuzztime=10s ./protocols/uniswap/v3
func FuzzNFTPermitSignRoundTrip(f *testing.F) {
	f.Add(make([]byte, 20), []byte{0x01}, []byte{0x00}, []byte{0xff}, []byte{0x01}, make([]byte, 20))
	f.Add([]byte{0xde, 0xad, 0xbe, 0xef}, []byte{0xff}, []byte{0xff}, []byte{0xff}, []byte{0xff}, []byte{0xbe, 0xef})

	f.Fuzz(func(t *testing.T, spenderRaw, tokenIdRaw, nonceRaw, deadlineRaw, chainIDRaw, verifyingContractRaw []byte) {
		spender := types.NewAddressFromBytes(spenderRaw)
		if len(tokenIdRaw) > 32 {
			tokenIdRaw = tokenIdRaw[len(tokenIdRaw)-32:]
		}
		tokenId := new(big.Int).SetBytes(tokenIdRaw)
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
		if chainID.Sign() == 0 {
			chainID = big.NewInt(1)
		}
		verifyingContract := types.NewAddressFromBytes(verifyingContractRaw)

		key, err := types.GeneratePrivateKey()
		require.NoError(t, err)
		signerAddr := core.PubkeyToAddress(key.PublicKey())

		domain := NewNFTPermitDomain(chainID, verifyingContract)
		p := NewNFTPermit(spender, tokenId, nonce, deadline)

		sig, err := p.Sign(domain, key)
		require.NoError(t, err)

		digest, err := p.Hash(domain)
		require.NoError(t, err)

		ok, err := core.VerifyAddress(digest, sig, signerAddr)
		require.NoError(t, err)
		require.True(t, ok)
	})
}
