package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewQuoteExactSingleParams(t *testing.T) {
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	poolKey := NewPoolKey(currency0, currency1, big.NewInt(3000), big.NewInt(60), hooks)
	exactAmount := big.NewInt(1000)
	hookData := []byte{0xde, 0xad}

	p := NewQuoteExactSingleParams(poolKey, true, exactAmount, hookData)
	require.Same(t, poolKey, p.PoolKey)
	require.True(t, p.ZeroForOne)
	require.Same(t, exactAmount, p.ExactAmount)
	require.Equal(t, hookData, p.HookData)

	p2 := NewQuoteExactSingleParams(poolKey, false, exactAmount, nil)
	require.False(t, p2.ZeroForOne)
	require.Nil(t, p2.HookData)
}

func TestNewQuoteExactSingleParamsSharesPointers(t *testing.T) {
	poolKey := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(500), big.NewInt(10), testHooks)
	exactAmount := big.NewInt(1)
	hookData := []byte{0x01}

	p := NewQuoteExactSingleParams(poolKey, true, exactAmount, hookData)
	exactAmount.SetInt64(42)
	hookData[0] = 0xff
	require.Zero(t, big.NewInt(42).Cmp(p.ExactAmount))
	require.Equal(t, byte(0xff), p.HookData[0])
}

func TestNewQuoteExactParams(t *testing.T) {
	exactCurrency := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	intermediate0 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	intermediate1 := types.MustNewAddressFromHex("0x1231deb6f5749ef6ce6943a275a1d3e7486f4eae")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*PathKey{
		NewPathKey(intermediate0, big.NewInt(3000), big.NewInt(60), hooks, nil),
		NewPathKey(intermediate1, big.NewInt(500), big.NewInt(10), hooks, []byte{0x01}),
	}
	exactAmount := big.NewInt(1000)

	p := NewQuoteExactParams(exactCurrency, path, exactAmount)
	require.Same(t, exactCurrency, p.ExactCurrency)
	require.Len(t, p.Path, 2)
	require.Same(t, path[0], p.Path[0])
	require.Same(t, path[1], p.Path[1])
	require.Same(t, exactAmount, p.ExactAmount)
}

func TestNewQuoteExactParamsEmptyPath(t *testing.T) {
	p := NewQuoteExactParams(testCurrency0, nil, big.NewInt(0))
	require.Nil(t, p.Path)
	require.Zero(t, p.ExactAmount.Sign())

	p2 := NewQuoteExactParams(testCurrency0, []*PathKey{}, big.NewInt(0))
	require.NotNil(t, p2.Path)
	require.Empty(t, p2.Path)
}

func TestNewQuoteExactInputResult(t *testing.T) {
	amountOut := big.NewInt(997)
	gasEstimate := big.NewInt(85000)

	r := NewQuoteExactInputResult(amountOut, gasEstimate)
	require.Same(t, amountOut, r.AmountOut)
	require.Same(t, gasEstimate, r.GasEstimate)
}

func TestNewQuoteExactOutputResult(t *testing.T) {
	amountIn := big.NewInt(1003)
	gasEstimate := big.NewInt(91000)

	r := NewQuoteExactOutputResult(amountIn, gasEstimate)
	require.Same(t, amountIn, r.AmountIn)
	require.Same(t, gasEstimate, r.GasEstimate)
}

func TestQuoterRequestsValidate(t *testing.T) {
	quoter := testPoolManager
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(3000), big.NewInt(60), testHooks)
	single := NewQuoteExactSingleParams(key, true, big.NewInt(1), nil)
	hop := NewPathKey(testCurrency1, big.NewInt(500), big.NewInt(10), testHooks, nil)
	multi := NewQuoteExactParams(testCurrency0, []*PathKey{hop}, big.NewInt(1))

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"msgSender nil", (*MsgSenderRequest)(nil).Validate(), "invalid request: request is nil"},
		{"msgSender empty", (&MsgSenderRequest{}).Validate(), "invalid request: contract is nil"},
		{"msgSender valid", (&MsgSenderRequest{Contract: quoter}).Validate(), ""},
		{"exactInputSingle nil", (*QuoteExactInputSingleRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exactInputSingle empty", (&QuoteExactInputSingleRequest{}).Validate(), "invalid request: contract is nil"},
		{"exactInputSingle nil params", (&QuoteExactInputSingleRequest{Contract: quoter}).Validate(), "invalid request: params is nil"},
		{"exactInputSingle nil poolKey", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactSingleParams{ExactAmount: big.NewInt(1)}}).Validate(), "invalid request: params.poolKey is nil"},
		{"exactInputSingle nil currency0", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactSingleParams{PoolKey: &PoolKey{Currency1: testCurrency1, Fee: big.NewInt(1), TickSpacing: big.NewInt(1), Hooks: testHooks}, ExactAmount: big.NewInt(1)}}).Validate(), "invalid request: params.poolKey.currency0 is nil"},
		{"exactInputSingle nil currency1", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactSingleParams{PoolKey: &PoolKey{Currency0: testCurrency0, Fee: big.NewInt(1), TickSpacing: big.NewInt(1), Hooks: testHooks}, ExactAmount: big.NewInt(1)}}).Validate(), "invalid request: params.poolKey.currency1 is nil"},
		{"exactInputSingle nil fee", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactSingleParams{PoolKey: &PoolKey{Currency0: testCurrency0, Currency1: testCurrency1, TickSpacing: big.NewInt(1), Hooks: testHooks}, ExactAmount: big.NewInt(1)}}).Validate(), "invalid request: params.poolKey.fee is nil"},
		{"exactInputSingle nil tickSpacing", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactSingleParams{PoolKey: &PoolKey{Currency0: testCurrency0, Currency1: testCurrency1, Fee: big.NewInt(1), Hooks: testHooks}, ExactAmount: big.NewInt(1)}}).Validate(), "invalid request: params.poolKey.tickSpacing is nil"},
		{"exactInputSingle nil hooks", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactSingleParams{PoolKey: &PoolKey{Currency0: testCurrency0, Currency1: testCurrency1, Fee: big.NewInt(1), TickSpacing: big.NewInt(1)}, ExactAmount: big.NewInt(1)}}).Validate(), "invalid request: params.poolKey.hooks is nil"},
		{"exactInputSingle nil exactAmount", (&QuoteExactInputSingleRequest{Contract: quoter, Params: &QuoteExactSingleParams{PoolKey: key}}).Validate(), "invalid request: params.exactAmount is nil"},
		{"exactInputSingle valid", (&QuoteExactInputSingleRequest{Contract: quoter, Params: single}).Validate(), ""},
		{"exactInputSingle zero amount", (&QuoteExactInputSingleRequest{Contract: quoter, Params: NewQuoteExactSingleParams(key, false, big.NewInt(0), []byte{0x01})}).Validate(), ""},
		{"exactOutputSingle nil", (*QuoteExactOutputSingleRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exactOutputSingle empty", (&QuoteExactOutputSingleRequest{}).Validate(), "invalid request: contract is nil"},
		{"exactOutputSingle nil params", (&QuoteExactOutputSingleRequest{Contract: quoter}).Validate(), "invalid request: params is nil"},
		{"exactOutputSingle nil exactAmount", (&QuoteExactOutputSingleRequest{Contract: quoter, Params: &QuoteExactSingleParams{PoolKey: key}}).Validate(), "invalid request: params.exactAmount is nil"},
		{"exactOutputSingle valid", (&QuoteExactOutputSingleRequest{Contract: quoter, Params: single}).Validate(), ""},
		{"exactInput nil", (*QuoteExactInputRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exactInput empty", (&QuoteExactInputRequest{}).Validate(), "invalid request: contract is nil"},
		{"exactInput nil params", (&QuoteExactInputRequest{Contract: quoter}).Validate(), "invalid request: params is nil"},
		{"exactInput nil exactCurrency", (&QuoteExactInputRequest{Contract: quoter, Params: &QuoteExactParams{Path: []*PathKey{hop}, ExactAmount: big.NewInt(1)}}).Validate(), "invalid request: params.exactCurrency is nil"},
		{"exactInput nil path entry", (&QuoteExactInputRequest{Contract: quoter, Params: NewQuoteExactParams(testCurrency0, []*PathKey{hop, nil}, big.NewInt(1))}).Validate(), "invalid request: params.path[1] is nil"},
		{"exactInput nil intermediateCurrency", (&QuoteExactInputRequest{Contract: quoter, Params: NewQuoteExactParams(testCurrency0, []*PathKey{{Fee: big.NewInt(1), TickSpacing: big.NewInt(1), Hooks: testHooks}}, big.NewInt(1))}).Validate(), "invalid request: params.path[0].intermediateCurrency is nil"},
		{"exactInput nil path fee", (&QuoteExactInputRequest{Contract: quoter, Params: NewQuoteExactParams(testCurrency0, []*PathKey{hop, {IntermediateCurrency: testHooks, TickSpacing: big.NewInt(1), Hooks: testHooks}}, big.NewInt(1))}).Validate(), "invalid request: params.path[1].fee is nil"},
		{"exactInput nil path tickSpacing", (&QuoteExactInputRequest{Contract: quoter, Params: NewQuoteExactParams(testCurrency0, []*PathKey{{IntermediateCurrency: testHooks, Fee: big.NewInt(1), Hooks: testHooks}}, big.NewInt(1))}).Validate(), "invalid request: params.path[0].tickSpacing is nil"},
		{"exactInput nil path hooks", (&QuoteExactInputRequest{Contract: quoter, Params: NewQuoteExactParams(testCurrency0, []*PathKey{{IntermediateCurrency: testHooks, Fee: big.NewInt(1), TickSpacing: big.NewInt(1)}}, big.NewInt(1))}).Validate(), "invalid request: params.path[0].hooks is nil"},
		{"exactInput nil exactAmount", (&QuoteExactInputRequest{Contract: quoter, Params: NewQuoteExactParams(testCurrency0, []*PathKey{hop}, nil)}).Validate(), "invalid request: params.exactAmount is nil"},
		{"exactInput empty path", (&QuoteExactInputRequest{Contract: quoter, Params: NewQuoteExactParams(testCurrency0, nil, big.NewInt(1))}).Validate(), ""},
		{"exactInput valid", (&QuoteExactInputRequest{Contract: quoter, Params: multi}).Validate(), ""},
		{"exactOutput nil", (*QuoteExactOutputRequest)(nil).Validate(), "invalid request: request is nil"},
		{"exactOutput empty", (&QuoteExactOutputRequest{}).Validate(), "invalid request: contract is nil"},
		{"exactOutput nil params", (&QuoteExactOutputRequest{Contract: quoter}).Validate(), "invalid request: params is nil"},
		{"exactOutput nil path entry", (&QuoteExactOutputRequest{Contract: quoter, Params: NewQuoteExactParams(testCurrency0, []*PathKey{nil}, big.NewInt(1))}).Validate(), "invalid request: params.path[0] is nil"},
		{"exactOutput valid", (&QuoteExactOutputRequest{Contract: quoter, Params: multi}).Validate(), ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if tc.want == "" {
				require.NoError(t, tc.err)
				return
			}
			require.ErrorIs(t, tc.err, bind.ErrInvalidRequest)
			require.EqualError(t, tc.err, tc.want)
		})
	}
}

func TestNewQuoterRequestsSetFields(t *testing.T) {
	quoter := testPoolManager
	key := NewPoolKey(testCurrency0, testCurrency1, big.NewInt(500), big.NewInt(10), testHooks)
	single := NewQuoteExactSingleParams(key, true, big.NewInt(7), []byte{0x01})
	multi := NewQuoteExactParams(testCurrency0, []*PathKey{NewPathKey(testCurrency1, big.NewInt(500), big.NewInt(10), testHooks, nil)}, big.NewInt(8))

	require.Equal(t, &MsgSenderRequest{Contract: quoter}, NewMsgSenderRequest(quoter))
	require.Equal(t, &QuoteExactInputSingleRequest{Contract: quoter, Params: single}, NewQuoteExactInputSingleRequest(quoter, single))
	require.Equal(t, &QuoteExactInputRequest{Contract: quoter, Params: multi}, NewQuoteExactInputRequest(quoter, multi))
	require.Equal(t, &QuoteExactOutputSingleRequest{Contract: quoter, Params: single}, NewQuoteExactOutputSingleRequest(quoter, single))
	require.Equal(t, &QuoteExactOutputRequest{Contract: quoter, Params: multi}, NewQuoteExactOutputRequest(quoter, multi))

	require.Same(t, single, NewQuoteExactInputSingleRequest(quoter, single).Params)
	require.Same(t, multi, NewQuoteExactOutputRequest(quoter, multi).Params)
	require.Same(t, quoter, NewMsgSenderRequest(quoter).Contract)
}
