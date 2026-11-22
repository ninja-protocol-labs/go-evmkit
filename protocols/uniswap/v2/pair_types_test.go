package v2

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewMintEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(200)

	m := NewMintEvent(contract, sender, amount0, amount1)
	require.Same(t, contract, m.Contract)
	require.Same(t, sender, m.Sender)
	require.Same(t, amount0, m.Amount0)
	require.Same(t, amount1, m.Amount1)
}

func TestNewBurnEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0 := big.NewInt(100)
	amount1 := big.NewInt(200)

	b := NewBurnEvent(contract, sender, amount0, amount1, to)
	require.Same(t, contract, b.Contract)
	require.Same(t, sender, b.Sender)
	require.Same(t, amount0, b.Amount0)
	require.Same(t, amount1, b.Amount1)
	require.Same(t, to, b.To)
}

func TestNewSwapEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	sender := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	to := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	amount0In := big.NewInt(100)
	amount1In := big.NewInt(0)
	amount0Out := big.NewInt(0)
	amount1Out := big.NewInt(95)

	s := NewSwapEvent(contract, sender, amount0In, amount1In, amount0Out, amount1Out, to)
	require.Same(t, contract, s.Contract)
	require.Same(t, sender, s.Sender)
	require.Same(t, amount0In, s.Amount0In)
	require.Same(t, amount1In, s.Amount1In)
	require.Same(t, amount0Out, s.Amount0Out)
	require.Same(t, amount1Out, s.Amount1Out)
	require.Same(t, to, s.To)
}

func TestNewSyncEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	reserve0 := big.NewInt(1000)
	reserve1 := big.NewInt(2000)

	s := NewSyncEvent(contract, reserve0, reserve1)
	require.Same(t, contract, s.Contract)
	require.Same(t, reserve0, s.Reserve0)
	require.Same(t, reserve1, s.Reserve1)
}

func TestNewReserves(t *testing.T) {
	reserve0 := big.NewInt(1000)
	reserve1 := big.NewInt(2000)
	blockTimestampLast := uint32(123456)

	r := NewReserves(reserve0, reserve1, blockTimestampLast)
	require.Same(t, reserve0, r.Reserve0)
	require.Same(t, reserve1, r.Reserve1)
	require.Equal(t, blockTimestampLast, r.BlockTimestampLast)
}

func TestPairRequestsValidate(t *testing.T) {
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"minimumLiquidity nil", (*MinimumLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"minimumLiquidity empty", (&MinimumLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"minimumLiquidity valid", (&MinimumLiquidityRequest{Contract: pair}).Validate(), ""},
		{"factory nil", (*FactoryRequest)(nil).Validate(), "invalid request: request is nil"},
		{"factory empty", (&FactoryRequest{}).Validate(), "invalid request: contract is nil"},
		{"factory valid", (&FactoryRequest{Contract: pair}).Validate(), ""},
		{"token0 nil", (*Token0Request)(nil).Validate(), "invalid request: request is nil"},
		{"token0 empty", (&Token0Request{}).Validate(), "invalid request: contract is nil"},
		{"token0 valid", (&Token0Request{Contract: pair}).Validate(), ""},
		{"token1 nil", (*Token1Request)(nil).Validate(), "invalid request: request is nil"},
		{"token1 empty", (&Token1Request{}).Validate(), "invalid request: contract is nil"},
		{"token1 valid", (&Token1Request{Contract: pair}).Validate(), ""},
		{"getReserves nil", (*GetReservesRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getReserves empty", (&GetReservesRequest{}).Validate(), "invalid request: contract is nil"},
		{"getReserves valid", (&GetReservesRequest{Contract: pair}).Validate(), ""},
		{"price0CumulativeLast nil", (*Price0CumulativeLastRequest)(nil).Validate(), "invalid request: request is nil"},
		{"price0CumulativeLast empty", (&Price0CumulativeLastRequest{}).Validate(), "invalid request: contract is nil"},
		{"price0CumulativeLast valid", (&Price0CumulativeLastRequest{Contract: pair}).Validate(), ""},
		{"price1CumulativeLast nil", (*Price1CumulativeLastRequest)(nil).Validate(), "invalid request: request is nil"},
		{"price1CumulativeLast empty", (&Price1CumulativeLastRequest{}).Validate(), "invalid request: contract is nil"},
		{"price1CumulativeLast valid", (&Price1CumulativeLastRequest{Contract: pair}).Validate(), ""},
		{"kLast nil", (*KLastRequest)(nil).Validate(), "invalid request: request is nil"},
		{"kLast empty", (&KLastRequest{}).Validate(), "invalid request: contract is nil"},
		{"kLast valid", (&KLastRequest{Contract: pair}).Validate(), ""},
		{"permitTypehash nil", (*PermitTypehashRequest)(nil).Validate(), "invalid request: request is nil"},
		{"permitTypehash empty", (&PermitTypehashRequest{}).Validate(), "invalid request: contract is nil"},
		{"permitTypehash valid", (&PermitTypehashRequest{Contract: pair}).Validate(), ""},
		{"initialize nil", (*InitializeRequest)(nil).Validate(), "invalid request: request is nil"},
		{"initialize empty", (&InitializeRequest{}).Validate(), "invalid request: contract is nil"},
		{"initialize nil contract", (&InitializeRequest{Token0: a, Token1: b}).Validate(), "invalid request: contract is nil"},
		{"initialize nil token0", (&InitializeRequest{Contract: pair, Token1: b}).Validate(), "invalid request: token0 is nil"},
		{"initialize nil token1", (&InitializeRequest{Contract: pair, Token0: a}).Validate(), "invalid request: token1 is nil"},
		{"initialize valid", (&InitializeRequest{Contract: pair, Token0: a, Token1: b}).Validate(), ""},
		{"mint nil", (*MintRequest)(nil).Validate(), "invalid request: request is nil"},
		{"mint empty", (&MintRequest{}).Validate(), "invalid request: contract is nil"},
		{"mint nil contract", (&MintRequest{To: a}).Validate(), "invalid request: contract is nil"},
		{"mint nil to", (&MintRequest{Contract: pair}).Validate(), "invalid request: to is nil"},
		{"mint valid", (&MintRequest{Contract: pair, To: a}).Validate(), ""},
		{"burn nil", (*BurnRequest)(nil).Validate(), "invalid request: request is nil"},
		{"burn empty", (&BurnRequest{}).Validate(), "invalid request: contract is nil"},
		{"burn nil contract", (&BurnRequest{To: a}).Validate(), "invalid request: contract is nil"},
		{"burn nil to", (&BurnRequest{Contract: pair}).Validate(), "invalid request: to is nil"},
		{"burn valid", (&BurnRequest{Contract: pair, To: a}).Validate(), ""},
		{"swap nil", (*SwapRequest)(nil).Validate(), "invalid request: request is nil"},
		{"swap empty", (&SwapRequest{}).Validate(), "invalid request: contract is nil"},
		{"swap nil contract", (&SwapRequest{Amount0Out: big.NewInt(0), Amount1Out: big.NewInt(1), To: a}).Validate(), "invalid request: contract is nil"},
		{"swap nil amount0Out", (&SwapRequest{Contract: pair, Amount1Out: big.NewInt(1), To: a}).Validate(), "invalid request: amount0Out is nil"},
		{"swap nil amount1Out", (&SwapRequest{Contract: pair, Amount0Out: big.NewInt(0), To: a}).Validate(), "invalid request: amount1Out is nil"},
		{"swap nil to", (&SwapRequest{Contract: pair, Amount0Out: big.NewInt(0), Amount1Out: big.NewInt(1)}).Validate(), "invalid request: to is nil"},
		{"swap nil data", (&SwapRequest{Contract: pair, Amount0Out: big.NewInt(0), Amount1Out: big.NewInt(1), To: a}).Validate(), ""},
		{"swap valid", (&SwapRequest{Contract: pair, Amount0Out: big.NewInt(0), Amount1Out: big.NewInt(1), To: a, Data: []byte{0x01}}).Validate(), ""},
		{"swap zero amounts", (&SwapRequest{Contract: pair, Amount0Out: big.NewInt(0), Amount1Out: big.NewInt(0), To: a}).Validate(), ""},
		{"skim nil", (*SkimRequest)(nil).Validate(), "invalid request: request is nil"},
		{"skim empty", (&SkimRequest{}).Validate(), "invalid request: contract is nil"},
		{"skim nil contract", (&SkimRequest{To: a}).Validate(), "invalid request: contract is nil"},
		{"skim nil to", (&SkimRequest{Contract: pair}).Validate(), "invalid request: to is nil"},
		{"skim valid", (&SkimRequest{Contract: pair, To: a}).Validate(), ""},
		{"sync nil", (*SyncRequest)(nil).Validate(), "invalid request: request is nil"},
		{"sync empty", (&SyncRequest{}).Validate(), "invalid request: contract is nil"},
		{"sync valid", (&SyncRequest{Contract: pair}).Validate(), ""},
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

func TestNewPairRequestsSetFields(t *testing.T) {
	pair := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	require.Equal(t, &MinimumLiquidityRequest{Contract: pair}, NewMinimumLiquidityRequest(pair))
	require.Equal(t, &FactoryRequest{Contract: pair}, NewFactoryRequest(pair))
	require.Equal(t, &Token0Request{Contract: pair}, NewToken0Request(pair))
	require.Equal(t, &Token1Request{Contract: pair}, NewToken1Request(pair))
	require.Equal(t, &GetReservesRequest{Contract: pair}, NewGetReservesRequest(pair))
	require.Equal(t, &Price0CumulativeLastRequest{Contract: pair}, NewPrice0CumulativeLastRequest(pair))
	require.Equal(t, &Price1CumulativeLastRequest{Contract: pair}, NewPrice1CumulativeLastRequest(pair))
	require.Equal(t, &KLastRequest{Contract: pair}, NewKLastRequest(pair))
	require.Equal(t, &PermitTypehashRequest{Contract: pair}, NewPermitTypehashRequest(pair))
	require.Equal(t, &InitializeRequest{Contract: pair, Token0: a, Token1: b}, NewInitializeRequest(pair, a, b))
	require.Equal(t, &MintRequest{Contract: pair, To: a}, NewMintRequest(pair, a))
	require.Equal(t, &BurnRequest{Contract: pair, To: b}, NewBurnRequest(pair, b))
	require.Equal(t, &SwapRequest{Contract: pair, Amount0Out: big.NewInt(0), Amount1Out: big.NewInt(95), To: a, Data: []byte{0xab}}, NewSwapRequest(pair, big.NewInt(0), big.NewInt(95), a, []byte{0xab}))
	require.Equal(t, &SkimRequest{Contract: pair, To: b}, NewSkimRequest(pair, b))
	require.Equal(t, &SyncRequest{Contract: pair}, NewSyncRequest(pair))
}
