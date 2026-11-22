package v2

import (
	"context"
	"math/big"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core"
	"github.com/ninja-protocol-labs/go-evmkit/core/abi"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-evmkit/rpc"
	"github.com/ninja-protocol-labs/go-evmkit/rpc/mocks"
	"github.com/ninja-protocol-labs/go-evmkit/tx"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

// go test -run '^$' -fuzz '^FuzzDecodeGetAmountsOutRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzDecodeGetAmountsOutRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03})
	f.Add(make([]byte, 32), make([]byte, 32), make([]byte, 32))

	f.Fuzz(func(t *testing.T, a, b, c []byte) {
		amounts := make([]*big.Int, 0, 3)
		for _, raw := range [][]byte{a, b, c} {
			if len(raw) > 32 {
				raw = raw[len(raw)-32:]
			}
			amounts = append(amounts, new(big.Int).SetBytes(raw))
		}

		data, err := abi.Pack(abi.Types{abi.Slice(abi.Uint256)}, amounts)
		require.NoError(t, err)

		got, err := DecodeGetAmountsOut(data)
		require.NoError(t, err)
		require.Len(t, got, len(amounts))
		for i, want := range amounts {
			require.Zero(t, want.Cmp(got[i]))
		}
	})
}

// go test -run '^$' -fuzz '^FuzzEncodeSwapExactTokensForTokensRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzEncodeSwapExactTokensForTokensRoundTrip(f *testing.F) {
	f.Add([]byte{}, []byte{}, make([]byte, 20), make([]byte, 20), make([]byte, 20))
	f.Add(make([]byte, 32), make([]byte, 32), []byte{0xde, 0xad, 0xbe, 0xef}, make([]byte, 20), make([]byte, 20))

	f.Fuzz(func(t *testing.T, amountInRaw, amountOutMinRaw, token0Raw, token1Raw, toRaw []byte) {
		if len(amountInRaw) > 32 {
			amountInRaw = amountInRaw[len(amountInRaw)-32:]
		}
		if len(amountOutMinRaw) > 32 {
			amountOutMinRaw = amountOutMinRaw[len(amountOutMinRaw)-32:]
		}
		amountIn := new(big.Int).SetBytes(amountInRaw)
		amountOutMin := new(big.Int).SetBytes(amountOutMinRaw)
		path := []*types.Address{types.NewAddressFromBytes(token0Raw), types.NewAddressFromBytes(token1Raw)}
		to := types.NewAddressFromBytes(toRaw)
		deadline := big.NewInt(9999999999)

		data, err := EncodeSwapExactTokensForTokens(amountIn, amountOutMin, path, to, deadline)
		require.NoError(t, err)
		require.Equal(t, []byte{0x38, 0xed, 0x17, 0x39}, data[:4])

		vals, err := swapExactTokensForTokensFn.DecodeCall(data)
		require.NoError(t, err)
		require.Zero(t, amountIn.Cmp(vals[0].(*big.Int)))
		require.Zero(t, amountOutMin.Cmp(vals[1].(*big.Int)))
		gotPath, ok := vals[2].([]any)
		require.True(t, ok)
		require.Len(t, gotPath, 2)
		require.True(t, gotPath[0].(*types.Address).Equal(path[0]))
		require.True(t, gotPath[1].(*types.Address).Equal(path[1]))
		require.True(t, vals[3].(*types.Address).Equal(to))
		require.Zero(t, deadline.Cmp(vals[4].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzDecodeAddLiquidityRoundTrip$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzDecodeAddLiquidityRoundTrip(f *testing.F) {
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03})
	f.Add(make([]byte, 32), make([]byte, 32), make([]byte, 32))

	f.Fuzz(func(t *testing.T, aRaw, bRaw, lRaw []byte) {
		if len(aRaw) > 32 {
			aRaw = aRaw[len(aRaw)-32:]
		}
		if len(bRaw) > 32 {
			bRaw = bRaw[len(bRaw)-32:]
		}
		if len(lRaw) > 32 {
			lRaw = lRaw[len(lRaw)-32:]
		}
		amountA := new(big.Int).SetBytes(aRaw)
		amountB := new(big.Int).SetBytes(bRaw)
		liquidity := new(big.Int).SetBytes(lRaw)

		data, err := abi.Pack(abi.Types{abi.Uint256, abi.Uint256, abi.Uint256}, amountA, amountB, liquidity)
		require.NoError(t, err)

		got, err := DecodeAddLiquidity(data)
		require.NoError(t, err)
		require.Zero(t, amountA.Cmp(got.AmountA))
		require.Zero(t, amountB.Cmp(got.AmountB))
		require.Zero(t, liquidity.Cmp(got.Liquidity))
	})
}

// go test -run '^$' -fuzz '^FuzzRouterSimpleDecodersDoNotPanic$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzRouterSimpleDecodersDoNotPanic(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 33))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))
	f.Add(make([]byte, 128))

	f.Fuzz(func(t *testing.T, data []byte) {
		_, _ = DecodeRouterFactory(data)
		_, _ = DecodeWETH(data)
		_, _ = DecodeQuote(data)
		_, _ = DecodeGetAmountOut(data)
		_, _ = DecodeGetAmountIn(data)
		_, _ = DecodeGetAmountsOut(data)
		_, _ = DecodeGetAmountsIn(data)
		_, _ = DecodeAddLiquidity(data)
		_, _ = DecodeAddLiquidityETH(data)
		_, _ = DecodeRemoveLiquidity(data)
		_, _ = DecodeRemoveLiquidityETH(data)
		_, _ = DecodeRemoveLiquidityWithPermit(data)
		_, _ = DecodeRemoveLiquidityETHWithPermit(data)
		_, _ = DecodeSwapExactTokensForTokens(data)
		_, _ = DecodeSwapTokensForExactTokens(data)
		_, _ = DecodeSwapExactETHForTokens(data)
		_, _ = DecodeSwapTokensForExactETH(data)
		_, _ = DecodeSwapExactTokensForETH(data)
		_, _ = DecodeSwapETHForExactTokens(data)
		_, _ = DecodeRemoveLiquidityETHSupportingFeeOnTransferTokens(data)
		_, _ = DecodeRemoveLiquidityETHWithPermitSupportingFeeOnTransferTokens(data)
	})
}

// go test -run '^$' -fuzz '^FuzzRouterSwapExactETHForTokensThroughClient$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzRouterSwapExactETHForTokensThroughClient(f *testing.F) {
	f.Add([]byte{}, []byte{}, []byte{}, []byte{}, []byte{}, uint8(2), uint64(0), true)
	f.Add([]byte{0x01}, []byte{0x02}, []byte{0x03}, []byte{0x04}, []byte{0x05}, uint8(3), uint64(42), false)
	f.Add(make([]byte, 20), make([]byte, 32), make([]byte, 32), make([]byte, 20), make([]byte, 32), uint8(7), ^uint64(0)-1, true)

	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(f, err)
	signer := core.PubkeyToAddress(key.PublicKey())

	f.Fuzz(func(t *testing.T, routerRaw, amountOutMinRaw, amountInRaw, toRaw, deadlineRaw []byte, hops uint8, nonce uint64, dynamic bool) {
		if len(amountOutMinRaw) > 32 {
			amountOutMinRaw = amountOutMinRaw[len(amountOutMinRaw)-32:]
		}
		if len(amountInRaw) > 16 {
			amountInRaw = amountInRaw[len(amountInRaw)-16:]
		}
		if len(deadlineRaw) > 32 {
			deadlineRaw = deadlineRaw[len(deadlineRaw)-32:]
		}
		router := types.NewAddressFromBytes(routerRaw)
		amountOutMin := new(big.Int).SetBytes(amountOutMinRaw)
		amountIn := new(big.Int).SetBytes(amountInRaw)
		to := types.NewAddressFromBytes(toRaw)
		deadline := new(big.Int).SetBytes(deadlineRaw)
		path := make([]*types.Address, int(hops%6)+1)
		for i := range path {
			path[i] = types.NewAddressFromBytes([]byte{byte(i + 1)})
		}

		var builder tx.Builder
		if dynamic {
			builder = tx.NewDynamicFeeBuilder()
		} else {
			builder = tx.NewLegacyBuilder()
		}
		builder.ChainID(big.NewInt(1))
		builder.Nonce(nonce)
		builder.Limit(3_000_000)
		builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

		var raw []byte
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
			require.NoError(t, err)
			*e.Result.(*string) = core.Keccak256(raw).String()
			return nil
		}).Once()

		opts := bind.NewTransactOptions(key, builder)
		signed, hash, err := NewRouter(c).SwapExactETHForTokens(t.Context(), opts, NewSwapExactETHForTokensRequest(router, amountOutMin, path, to, deadline, amountIn))
		require.NoError(t, err)
		require.True(t, hash.Equal(core.Keccak256(raw)))
		require.Equal(t, nonce, signed.AddressNonce())
		require.Nil(t, opts.Value)

		decoded, err := core.DecodeTransaction(raw)
		require.NoError(t, err)
		from, err := decoded.Sender()
		require.NoError(t, err)
		require.True(t, from.Equal(signer))

		var gotTo *types.Address
		var gotValue *big.Int
		var gotData []byte
		switch v := decoded.(type) {
		case *core.DynamicFeeTx:
			require.True(t, dynamic)
			gotTo, gotValue, gotData = v.To, v.Value, v.Data
		case *core.LegacyTx:
			require.False(t, dynamic)
			gotTo, gotValue, gotData = v.To, v.Value, v.Data
		default:
			t.Fatalf("unexpected tx type %T", decoded)
		}
		require.True(t, gotTo.Equal(router))
		require.Zero(t, amountIn.Cmp(gotValue))

		vals, err := swapExactETHForTokensFn.DecodeCall(gotData)
		require.NoError(t, err)
		require.Zero(t, amountOutMin.Cmp(vals[0].(*big.Int)))
		gotPath := vals[1].([]any)
		require.Len(t, gotPath, len(path))
		for i, addr := range path {
			require.True(t, addr.Equal(gotPath[i].(*types.Address)))
		}
		require.True(t, to.Equal(vals[2].(*types.Address)))
		require.Zero(t, deadline.Cmp(vals[3].(*big.Int)))
	})
}

// go test -run '^$' -fuzz '^FuzzRouterReadsSurviveArbitraryResult$' -fuzztime=10s ./protocols/uniswap/v2
func FuzzRouterReadsSurviveArbitraryResult(f *testing.F) {
	f.Add([]byte{})
	f.Add(make([]byte, 31))
	f.Add(make([]byte, 32))
	f.Add(make([]byte, 64))
	f.Add(make([]byte, 96))
	f.Add(append(append(make([]byte, 31), 0x20), append(make([]byte, 31), 0x02)...))

	router := types.MustNewAddressFromHex("0x7a250d5630B4cF539739dF2C5dAcb4c659F2488D")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	f.Fuzz(func(t *testing.T, result []byte) {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(result)
			return nil
		}).Times(6)
		r := NewRouter(c)

		weth, err := r.WETH(t.Context(), nil, NewWETHRequest(router))
		require.True(t, (err == nil) != (weth == nil))
		quote, err := r.Quote(t.Context(), nil, NewQuoteRequest(router, big.NewInt(1), big.NewInt(1), big.NewInt(1)))
		require.True(t, (err == nil) != (quote == nil))
		amounts, err := r.GetAmountsOut(t.Context(), nil, NewGetAmountsOutRequest(router, big.NewInt(1), []*types.Address{a, b}))
		require.True(t, err == nil || amounts == nil)
		added, err := r.CallAddLiquidity(t.Context(), nil, NewAddLiquidityRequest(router, a, b, big.NewInt(1), big.NewInt(1), big.NewInt(0), big.NewInt(0), a, big.NewInt(1)))
		require.True(t, (err == nil) != (added == nil))
		removed, err := r.CallRemoveLiquidityETH(t.Context(), nil, NewRemoveLiquidityETHRequest(router, a, big.NewInt(1), big.NewInt(0), big.NewInt(0), a, big.NewInt(1)))
		require.True(t, (err == nil) != (removed == nil))
		amountETH, err := r.CallRemoveLiquidityETHSupportingFeeOnTransferTokens(t.Context(), nil, NewRemoveLiquidityETHSupportingFeeOnTransferTokensRequest(router, a, big.NewInt(1), big.NewInt(0), big.NewInt(0), a, big.NewInt(1)))
		require.True(t, (err == nil) != (amountETH == nil))
	})
}
