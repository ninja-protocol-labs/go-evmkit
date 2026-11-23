package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewPathKey(t *testing.T) {
	intermediateCurrency := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)
	hooks := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hookData := []byte{0x01, 0x02}

	p := NewPathKey(intermediateCurrency, fee, tickSpacing, hooks, hookData)
	require.Same(t, intermediateCurrency, p.IntermediateCurrency)
	require.Same(t, fee, p.Fee)
	require.Same(t, tickSpacing, p.TickSpacing)
	require.Same(t, hooks, p.Hooks)
	require.Equal(t, hookData, p.HookData)
}

func TestNewExactInputSingleParams(t *testing.T) {
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	poolKey := NewPoolKey(currency0, currency1, big.NewInt(3000), big.NewInt(60), hooks)
	amountIn := big.NewInt(1000)
	amountOutMinimum := big.NewInt(900)
	minHopPriceX36 := big.NewInt(123456789)
	hookData := []byte{0x03}

	p := NewExactInputSingleParams(poolKey, true, amountIn, amountOutMinimum, minHopPriceX36, hookData)
	require.Same(t, poolKey, p.PoolKey)
	require.True(t, p.ZeroForOne)
	require.Same(t, amountIn, p.AmountIn)
	require.Same(t, amountOutMinimum, p.AmountOutMinimum)
	require.Same(t, minHopPriceX36, p.MinHopPriceX36)
	require.Equal(t, hookData, p.HookData)
}

func TestNewExactInputParams(t *testing.T) {
	currencyIn := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	intermediateCurrency := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*PathKey{NewPathKey(intermediateCurrency, big.NewInt(3000), big.NewInt(60), hooks, nil)}
	minHopPriceX36 := []*big.Int{big.NewInt(123456789)}
	amountIn := big.NewInt(1000)
	amountOutMinimum := big.NewInt(900)

	p := NewExactInputParams(currencyIn, path, minHopPriceX36, amountIn, amountOutMinimum)
	require.Same(t, currencyIn, p.CurrencyIn)
	require.Equal(t, path, p.Path)
	require.Equal(t, minHopPriceX36, p.MinHopPriceX36)
	require.Same(t, amountIn, p.AmountIn)
	require.Same(t, amountOutMinimum, p.AmountOutMinimum)
}

func TestNewExactOutputSingleParams(t *testing.T) {
	currency0 := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	currency1 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	poolKey := NewPoolKey(currency0, currency1, big.NewInt(3000), big.NewInt(60), hooks)
	amountOut := big.NewInt(1000)
	amountInMaximum := big.NewInt(1100)
	minHopPriceX36 := big.NewInt(123456789)
	hookData := []byte{0x04}

	p := NewExactOutputSingleParams(poolKey, false, amountOut, amountInMaximum, minHopPriceX36, hookData)
	require.Same(t, poolKey, p.PoolKey)
	require.False(t, p.ZeroForOne)
	require.Same(t, amountOut, p.AmountOut)
	require.Same(t, amountInMaximum, p.AmountInMaximum)
	require.Same(t, minHopPriceX36, p.MinHopPriceX36)
	require.Equal(t, hookData, p.HookData)
}

func TestNewExactOutputParams(t *testing.T) {
	currencyOut := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	intermediateCurrency := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	hooks := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	path := []*PathKey{NewPathKey(intermediateCurrency, big.NewInt(3000), big.NewInt(60), hooks, nil)}
	minHopPriceX36 := []*big.Int{big.NewInt(123456789)}
	amountOut := big.NewInt(1000)
	amountInMaximum := big.NewInt(1100)

	p := NewExactOutputParams(currencyOut, path, minHopPriceX36, amountOut, amountInMaximum)
	require.Same(t, currencyOut, p.CurrencyOut)
	require.Equal(t, path, p.Path)
	require.Equal(t, minHopPriceX36, p.MinHopPriceX36)
	require.Same(t, amountOut, p.AmountOut)
	require.Same(t, amountInMaximum, p.AmountInMaximum)
}

func TestExecuteActionsRequestValidate(t *testing.T) {
	router := testHooks
	actions := []byte{ActionSwapExactInSingle, ActionSettleAll, ActionTakeAll}
	params := [][]byte{{0x01}, {0x02}, {}}
	overflow := new(big.Int).Lsh(big.NewInt(1), 256)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"nil", (*ExecuteActionsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"empty", (&ExecuteActionsRequest{}).Validate(), "invalid request: contract is nil"},
		{"nil contract", (&ExecuteActionsRequest{Actions: actions, Params: params}).Validate(), "invalid request: contract is nil"},
		{"empty actions", (&ExecuteActionsRequest{Contract: router, Params: params}).Validate(), "invalid request: actions is empty"},
		{"fewer params", (&ExecuteActionsRequest{Contract: router, Actions: actions, Params: params[:2]}).Validate(), "invalid request: params length 2 does not match actions length 3"},
		{"more params", (&ExecuteActionsRequest{Contract: router, Actions: actions[:1], Params: params}).Validate(), "invalid request: params length 3 does not match actions length 1"},
		{"negative native", (&ExecuteActionsRequest{Contract: router, Actions: actions, Params: params, NativeAmount: big.NewInt(-1)}).Validate(), "invalid request: nativeAmount is negative"},
		{"overflow native", (&ExecuteActionsRequest{Contract: router, Actions: actions, Params: params, NativeAmount: overflow}).Validate(), "invalid request: nativeAmount exceeds uint256"},
		{"nil native", (&ExecuteActionsRequest{Contract: router, Actions: actions, Params: params}).Validate(), ""},
		{"zero native", (&ExecuteActionsRequest{Contract: router, Actions: actions, Params: params, NativeAmount: big.NewInt(0)}).Validate(), ""},
		{"max native", (&ExecuteActionsRequest{Contract: router, Actions: actions, Params: params, NativeAmount: new(big.Int).Sub(overflow, big.NewInt(1))}).Validate(), ""},
		{"nil param entry", (&ExecuteActionsRequest{Contract: router, Actions: actions[:1], Params: [][]byte{nil}}).Validate(), ""},
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

func TestNewExecuteActionsRequestSetsFields(t *testing.T) {
	router := testHooks
	actions := []byte{ActionSettleAll}
	params := [][]byte{{0x09}}
	native := big.NewInt(77)

	req := NewExecuteActionsRequest(router, actions, params, native)
	require.Equal(t, &ExecuteActionsRequest{Contract: router, Actions: actions, Params: params, NativeAmount: native}, req)
	require.Same(t, router, req.Contract)
	require.Same(t, native, req.NativeAmount)
}
