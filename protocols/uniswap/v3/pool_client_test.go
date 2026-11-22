package v3

import (
	"context"
	"encoding/json/jsontext"
	"errors"
	"fmt"
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

func TestPoolReadsSendCallAndDecode(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	opts := bind.NewCallOptions().WithFrom(a).WithBlock("0x1b4")
	var key [32]byte
	key[0], key[31] = 0x11, 0x22

	factoryResult, err := abi.Pack(abi.NewTypes(abi.Address), a)
	require.NoError(t, err)
	token0Result, err := abi.Pack(abi.NewTypes(abi.Address), a)
	require.NoError(t, err)
	token1Result, err := abi.Pack(abi.NewTypes(abi.Address), b)
	require.NoError(t, err)
	feeResult, err := abi.Pack(abi.NewTypes(abi.Uint24), big.NewInt(3000))
	require.NoError(t, err)
	tickSpacingResult, err := abi.Pack(abi.NewTypes(abi.Int24), big.NewInt(60))
	require.NoError(t, err)
	maxLiquidityPerTickResult, err := abi.Pack(abi.NewTypes(abi.Uint128), new(big.Int).Lsh(big.NewInt(1), 100))
	require.NoError(t, err)
	slot0Result, err := abi.Pack(abi.NewTypes(abi.Uint160, abi.Int24, abi.Uint16, abi.Uint16, abi.Uint16, abi.Uint8, abi.Bool), new(big.Int).Lsh(big.NewInt(1), 96), big.NewInt(-887272), uint16(5), uint16(10), uint16(20), uint8(68), true)
	require.NoError(t, err)
	feeGrowthGlobal0X128Result, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(999))
	require.NoError(t, err)
	feeGrowthGlobal1X128Result, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(888))
	require.NoError(t, err)
	protocolFeesResult, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Uint128), big.NewInt(100), big.NewInt(200))
	require.NoError(t, err)
	liquidityResult, err := abi.Pack(abi.NewTypes(abi.Uint128), big.NewInt(42))
	require.NoError(t, err)
	ticksResult, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Int128, abi.Uint256, abi.Uint256, abi.Int56, abi.Uint160, abi.Uint32, abi.Bool), big.NewInt(1), big.NewInt(-2), big.NewInt(3), big.NewInt(4), big.NewInt(-5), big.NewInt(6), uint32(7), true)
	require.NoError(t, err)
	ticksData, err := EncodeTicks(big.NewInt(-60))
	require.NoError(t, err)
	tickBitmapResult, err := abi.Pack(abi.NewTypes(abi.Uint256), big.NewInt(255))
	require.NoError(t, err)
	tickBitmapData, err := EncodeTickBitmap(-3)
	require.NoError(t, err)
	positionsResult, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Uint256, abi.Uint256, abi.Uint128, abi.Uint128), big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5))
	require.NoError(t, err)
	positionsData, err := EncodePositions(key)
	require.NoError(t, err)
	observationsResult, err := abi.Pack(abi.NewTypes(abi.Uint32, abi.Int56, abi.Uint160, abi.Bool), uint32(111), big.NewInt(-222), big.NewInt(333), true)
	require.NoError(t, err)
	observationsData, err := EncodePoolObservations(big.NewInt(2))
	require.NoError(t, err)
	observeResult, err := abi.Pack(abi.NewTypes(abi.Slice(abi.Int56), abi.Slice(abi.Uint160)), []*big.Int{big.NewInt(-10), big.NewInt(20)}, []*big.Int{big.NewInt(30), big.NewInt(40)})
	require.NoError(t, err)
	observeData, err := EncodeObserve([]uint32{0, 60})
	require.NoError(t, err)
	snapshotCumulativesInsideResult, err := abi.Pack(abi.NewTypes(abi.Int56, abi.Uint160, abi.Uint32), big.NewInt(-1), big.NewInt(2), uint32(3))
	require.NoError(t, err)
	snapshotData, err := EncodeSnapshotCumulativesInside(big.NewInt(-60), big.NewInt(60))
	require.NoError(t, err)
	callMintResult, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(11), big.NewInt(22))
	require.NoError(t, err)
	mintData, err := EncodeMint(b, big.NewInt(-60), big.NewInt(60), big.NewInt(1000), []byte{0xca, 0xfe})
	require.NoError(t, err)
	callCollectResult, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Uint128), big.NewInt(5), big.NewInt(6))
	require.NoError(t, err)
	collectData, err := EncodeCollect(b, big.NewInt(-60), big.NewInt(60), big.NewInt(5), big.NewInt(6))
	require.NoError(t, err)
	callBurnResult, err := abi.Pack(abi.NewTypes(abi.Uint256, abi.Uint256), big.NewInt(33), big.NewInt(44))
	require.NoError(t, err)
	burnData, err := EncodeBurn(big.NewInt(-60), big.NewInt(60), big.NewInt(1000))
	require.NoError(t, err)
	callSwapResult, err := abi.Pack(abi.NewTypes(abi.Int256, abi.Int256), big.NewInt(510), big.NewInt(-500))
	require.NoError(t, err)
	swapData, err := EncodeSwap(b, true, big.NewInt(-500), big.NewInt(4295128740), nil)
	require.NoError(t, err)
	callCollectProtocolResult, err := abi.Pack(abi.NewTypes(abi.Uint128, abi.Uint128), big.NewInt(7), big.NewInt(8))
	require.NoError(t, err)
	collectProtocolData, err := EncodeCollectProtocol(b, big.NewInt(7), big.NewInt(8))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		result   []byte
		call     func(p *Pool) (any, error)
		want     any
	}{
		{
			"factory",
			EncodeFactory(),
			factoryResult,
			func(p *Pool) (any, error) {
				return p.Factory(t.Context(), opts, NewFactoryRequest(pool))
			},
			a,
		},
		{
			"token0",
			EncodeToken0(),
			token0Result,
			func(p *Pool) (any, error) {
				return p.Token0(t.Context(), opts, NewToken0Request(pool))
			},
			a,
		},
		{
			"token1",
			EncodeToken1(),
			token1Result,
			func(p *Pool) (any, error) {
				return p.Token1(t.Context(), opts, NewToken1Request(pool))
			},
			b,
		},
		{
			"fee",
			EncodePoolFee(),
			feeResult,
			func(p *Pool) (any, error) {
				return p.Fee(t.Context(), opts, NewFeeRequest(pool))
			},
			big.NewInt(3000),
		},
		{
			"tickSpacing",
			EncodePoolTickSpacing(),
			tickSpacingResult,
			func(p *Pool) (any, error) {
				return p.TickSpacing(t.Context(), opts, NewTickSpacingRequest(pool))
			},
			big.NewInt(60),
		},
		{
			"maxLiquidityPerTick",
			EncodeMaxLiquidityPerTick(),
			maxLiquidityPerTickResult,
			func(p *Pool) (any, error) {
				return p.MaxLiquidityPerTick(t.Context(), opts, NewMaxLiquidityPerTickRequest(pool))
			},
			new(big.Int).Lsh(big.NewInt(1), 100),
		},
		{
			"slot0",
			EncodeSlot0(),
			slot0Result,
			func(p *Pool) (any, error) {
				return p.Slot0(t.Context(), opts, NewSlot0Request(pool))
			},
			NewSlot0(new(big.Int).Lsh(big.NewInt(1), 96), big.NewInt(-887272), big.NewInt(5), big.NewInt(10), big.NewInt(20), big.NewInt(68), true),
		},
		{
			"feeGrowthGlobal0X128",
			EncodeFeeGrowthGlobal0X128(),
			feeGrowthGlobal0X128Result,
			func(p *Pool) (any, error) {
				return p.FeeGrowthGlobal0X128(t.Context(), opts, NewFeeGrowthGlobal0X128Request(pool))
			},
			big.NewInt(999),
		},
		{
			"feeGrowthGlobal1X128",
			EncodeFeeGrowthGlobal1X128(),
			feeGrowthGlobal1X128Result,
			func(p *Pool) (any, error) {
				return p.FeeGrowthGlobal1X128(t.Context(), opts, NewFeeGrowthGlobal1X128Request(pool))
			},
			big.NewInt(888),
		},
		{
			"protocolFees",
			EncodeProtocolFees(),
			protocolFeesResult,
			func(p *Pool) (any, error) {
				return p.ProtocolFees(t.Context(), opts, NewProtocolFeesRequest(pool))
			},
			NewProtocolFees(big.NewInt(100), big.NewInt(200)),
		},
		{
			"liquidity",
			EncodePoolLiquidity(),
			liquidityResult,
			func(p *Pool) (any, error) {
				return p.Liquidity(t.Context(), opts, NewLiquidityRequest(pool))
			},
			big.NewInt(42),
		},
		{
			"ticks",
			ticksData,
			ticksResult,
			func(p *Pool) (any, error) {
				return p.Ticks(t.Context(), opts, NewTicksRequest(pool, big.NewInt(-60)))
			},
			NewTickInfo(big.NewInt(1), big.NewInt(-2), big.NewInt(3), big.NewInt(4), big.NewInt(-5), big.NewInt(6), big.NewInt(7), true),
		},
		{
			"tickBitmap",
			tickBitmapData,
			tickBitmapResult,
			func(p *Pool) (any, error) {
				return p.TickBitmap(t.Context(), opts, NewTickBitmapRequest(pool, -3))
			},
			big.NewInt(255),
		},
		{
			"positions",
			positionsData,
			positionsResult,
			func(p *Pool) (any, error) {
				return p.Positions(t.Context(), opts, NewPositionsRequest(pool, key))
			},
			NewPositionInfo(big.NewInt(1), big.NewInt(2), big.NewInt(3), big.NewInt(4), big.NewInt(5)),
		},
		{
			"observations",
			observationsData,
			observationsResult,
			func(p *Pool) (any, error) {
				return p.Observations(t.Context(), opts, NewObservationsRequest(pool, big.NewInt(2)))
			},
			NewObservationInfo(big.NewInt(111), big.NewInt(-222), big.NewInt(333), true),
		},
		{
			"observe",
			observeData,
			observeResult,
			func(p *Pool) (any, error) {
				return p.Observe(t.Context(), opts, NewObserveRequest(pool, []uint32{0, 60}))
			},
			NewObserveResult([]*big.Int{big.NewInt(-10), big.NewInt(20)}, []*big.Int{big.NewInt(30), big.NewInt(40)}),
		},
		{
			"snapshotCumulativesInside",
			snapshotData,
			snapshotCumulativesInsideResult,
			func(p *Pool) (any, error) {
				return p.SnapshotCumulativesInside(t.Context(), opts, NewSnapshotCumulativesInsideRequest(pool, big.NewInt(-60), big.NewInt(60)))
			},
			NewSnapshotCumulativesInside(big.NewInt(-1), big.NewInt(2), big.NewInt(3)),
		},
		{
			"callMint",
			mintData,
			callMintResult,
			func(p *Pool) (any, error) {
				amount0, amount1, err := p.CallMint(t.Context(), opts, NewPoolMintRequest(pool, b, big.NewInt(-60), big.NewInt(60), big.NewInt(1000), []byte{0xca, 0xfe}))
				return []*big.Int{amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(11), big.NewInt(22)},
		},
		{
			"callCollect",
			collectData,
			callCollectResult,
			func(p *Pool) (any, error) {
				amount0, amount1, err := p.CallCollect(t.Context(), opts, NewPoolCollectRequest(pool, b, big.NewInt(-60), big.NewInt(60), big.NewInt(5), big.NewInt(6)))
				return []*big.Int{amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(5), big.NewInt(6)},
		},
		{
			"callBurn",
			burnData,
			callBurnResult,
			func(p *Pool) (any, error) {
				amount0, amount1, err := p.CallBurn(t.Context(), opts, NewBurnRequest(pool, big.NewInt(-60), big.NewInt(60), big.NewInt(1000)))
				return []*big.Int{amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(33), big.NewInt(44)},
		},
		{
			"callSwap",
			swapData,
			callSwapResult,
			func(p *Pool) (any, error) {
				amount0, amount1, err := p.CallSwap(t.Context(), opts, NewSwapRequest(pool, b, true, big.NewInt(-500), big.NewInt(4295128740), nil))
				return []*big.Int{amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(510), big.NewInt(-500)},
		},
		{
			"callCollectProtocol",
			collectProtocolData,
			callCollectProtocolResult,
			func(p *Pool) (any, error) {
				amount0, amount1, err := p.CallCollectProtocol(t.Context(), opts, NewCollectProtocolRequest(pool, b, big.NewInt(7), big.NewInt(8)))
				return []*big.Int{amount0, amount1}, err
			},
			[]*big.Int{big.NewInt(7), big.NewInt(8)},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHCall, e.Method)
				params, ok := e.Params.([]any)
				require.True(t, ok)
				require.Len(t, params, 2)
				call, ok := params[0].(map[string]any)
				require.True(t, ok)
				require.Equal(t, pool.String(), call["to"])
				require.Equal(t, a.String(), call["from"])
				require.Equal(t, encoding.Hex.EncodePrefixed(tc.calldata), call["data"])
				require.NotContains(t, call, "value")
				require.Equal(t, "0x1b4", params[1])
				*e.Result.(*string) = encoding.Hex.EncodePrefixed(tc.result)
				return nil
			}).Once()

			got, err := tc.call(NewPool(c))
			require.NoError(t, err)
			require.Equal(t, fmt.Sprint(tc.want), fmt.Sprint(got))
		})
	}
}

func TestPoolReadsMapRevertToSentinel(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	errorSel := []byte{0x08, 0xc3, 0x79, 0xa0}
	all := []error{ErrLocked, ErrTickLowerGreaterThanOrEqual, ErrTickLowerTooLow, ErrTickUpperTooHigh, ErrAlreadyInitialized, ErrMintToken0Underpaid, ErrMintToken1Underpaid, ErrAmountSpecifiedZero, ErrInvalidSqrtPriceLimit, ErrSwapUnderpaid, ErrNoLiquidity, ErrFlashToken0Underpaid, ErrFlashToken1Underpaid, ErrLiquiditySubUnderflow, ErrLiquidityAddOverflow, ErrOracleNotInitialized, ErrOracleObservationTooOld, ErrNoPosition, ErrLiquidityOverflow, ErrTickOutOfBounds, ErrSqrtRatioOutOfBounds, ErrTransferFailed, ErrReverted}

	for _, tc := range []struct {
		name   string
		reason string
		want   error
	}{
		{"locked", reasonLocked, ErrLocked},
		{"tick lower >= upper", reasonTickLowerGreaterThanOrEqualUpper, ErrTickLowerGreaterThanOrEqual},
		{"tick lower too low", reasonTickLowerTooLow, ErrTickLowerTooLow},
		{"tick upper too high", reasonTickUpperTooHigh, ErrTickUpperTooHigh},
		{"already initialized", reasonAlreadyInitialized, ErrAlreadyInitialized},
		{"mint token0 underpaid", reasonMintToken0Underpaid, ErrMintToken0Underpaid},
		{"mint token1 underpaid", reasonMintToken1Underpaid, ErrMintToken1Underpaid},
		{"amount specified zero", reasonAmountSpecifiedZero, ErrAmountSpecifiedZero},
		{"invalid sqrt price limit", reasonInvalidSqrtPriceLimit, ErrInvalidSqrtPriceLimit},
		{"swap underpaid", reasonSwapUnderpaid, ErrSwapUnderpaid},
		{"no liquidity", reasonNoLiquidity, ErrNoLiquidity},
		{"flash token0 underpaid", reasonFlashToken0Underpaid, ErrFlashToken0Underpaid},
		{"flash token1 underpaid", reasonFlashToken1Underpaid, ErrFlashToken1Underpaid},
		{"liquidity sub underflow", reasonLiquiditySubUnderflow, ErrLiquiditySubUnderflow},
		{"liquidity add overflow", reasonLiquidityAddOverflow, ErrLiquidityAddOverflow},
		{"oracle not initialized", reasonOracleNotInitialized, ErrOracleNotInitialized},
		{"oracle observation too old", reasonOracleObservationTooOld, ErrOracleObservationTooOld},
		{"no position", reasonNoPosition, ErrNoPosition},
		{"liquidity overflow", reasonLiquidityOverflow, ErrLiquidityOverflow},
		{"tick out of bounds", reasonTickOutOfBounds, ErrTickOutOfBounds},
		{"sqrt ratio out of bounds", reasonSqrtRatioOutOfBounds, ErrSqrtRatioOutOfBounds},
		{"transfer failed", reasonTransferFailed, ErrTransferFailed},
		{"other reason", "XYZ", ErrReverted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			packed, err := abi.Pack(abi.NewTypes(abi.String), tc.reason)
			require.NoError(t, err)
			revert := append(append([]byte{}, errorSel...), packed...)

			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)}).Twice()
			p := NewPool(c)

			amount0, amount1, err := p.CallSwap(t.Context(), bind.NewCallOptions().WithFrom(a), NewSwapRequest(pool, b, true, big.NewInt(1), big.NewInt(4295128740), nil))
			require.Nil(t, amount0)
			require.Nil(t, amount1)
			require.ErrorIs(t, err, tc.want)
			for _, other := range all {
				if other != tc.want {
					require.NotErrorIs(t, err, other)
				}
			}
			var rpcErr *rpc.ResponseError
			require.ErrorAs(t, err, &rpcErr)

			slot0, err := p.Slot0(t.Context(), nil, NewSlot0Request(pool))
			require.Nil(t, slot0)
			require.ErrorIs(t, err, tc.want)
			require.ErrorContains(t, err, "uniswapv3: slot0: ")
		})
	}

	for _, data := range []jsontext.Value{nil, jsontext.Value(`"0x"`), jsontext.Value(`{"x":1}`), jsontext.Value(`"0xzz"`), jsontext.Value(`"0xdeadbeef"`)} {
		c := mocks.NewClient(t)
		c.EXPECT().Call(mock.Anything, mock.Anything).Return(&rpc.ResponseError{Code: 3, Message: "execution reverted", Data: data}).Once()

		snapshot, err := NewPool(c).SnapshotCumulativesInside(t.Context(), nil, NewSnapshotCumulativesInsideRequest(pool, big.NewInt(-60), big.NewInt(60)))
		require.Nil(t, snapshot)
		require.ErrorIs(t, err, ErrReverted)
		require.ErrorContains(t, err, "uniswapv3: snapshotCumulativesInside: reverted: ")
	}

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	liquidity, err := NewPool(c).Liquidity(t.Context(), nil, NewLiquidityRequest(pool))
	require.Nil(t, liquidity)
	require.EqualError(t, err, "uniswapv3: liquidity: dial tcp: connection refused")
	require.NotErrorIs(t, err, ErrReverted)
}

func TestPoolReadsRejectMalformedResult(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	for _, result := range []string{"0x", "0xzz", "0x01", "0x00", "not hex"} {
		t.Run(result, func(t *testing.T) {
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				*e.Result.(*string) = result
				return nil
			}).Times(5)
			p := NewPool(c)

			slot0, err := p.Slot0(t.Context(), nil, NewSlot0Request(pool))
			require.Nil(t, slot0)
			require.ErrorContains(t, err, "uniswapv3: slot0: ")

			tickInfo, err := p.Ticks(t.Context(), nil, NewTicksRequest(pool, big.NewInt(0)))
			require.Nil(t, tickInfo)
			require.ErrorContains(t, err, "uniswapv3: ticks: ")

			observed, err := p.Observe(t.Context(), nil, NewObserveRequest(pool, []uint32{0}))
			require.Nil(t, observed)
			require.ErrorContains(t, err, "uniswapv3: observe: ")

			token0, err := p.Token0(t.Context(), nil, NewToken0Request(pool))
			require.Nil(t, token0)
			require.ErrorContains(t, err, "uniswapv3: token0: ")

			amount0, amount1, err := p.CallSwap(t.Context(), nil, NewSwapRequest(pool, b, true, big.NewInt(1), big.NewInt(4295128740), nil))
			require.Nil(t, amount0)
			require.Nil(t, amount1)
			require.ErrorContains(t, err, "uniswapv3: callSwap: ")
		})
	}
}

func TestPoolMethodsRejectInvalidRequestWithoutRPC(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	opts := bind.NewTransactOptions(key, tx.NewDynamicFeeBuilder())
	p := NewPool(mocks.NewClient(t))
	ctx := t.Context()

	_, err = p.Factory(ctx, nil, &FactoryRequest{})
	require.EqualError(t, err, "uniswapv3: factory: "+(&FactoryRequest{}).Validate().Error())
	_, err = p.Token0(ctx, nil, &Token0Request{})
	require.EqualError(t, err, "uniswapv3: token0: "+(&Token0Request{}).Validate().Error())
	_, err = p.Token1(ctx, nil, &Token1Request{})
	require.EqualError(t, err, "uniswapv3: token1: "+(&Token1Request{}).Validate().Error())
	_, err = p.Fee(ctx, nil, &FeeRequest{})
	require.EqualError(t, err, "uniswapv3: fee: "+(&FeeRequest{}).Validate().Error())
	_, err = p.TickSpacing(ctx, nil, &TickSpacingRequest{})
	require.EqualError(t, err, "uniswapv3: tickSpacing: "+(&TickSpacingRequest{}).Validate().Error())
	_, err = p.MaxLiquidityPerTick(ctx, nil, &MaxLiquidityPerTickRequest{})
	require.EqualError(t, err, "uniswapv3: maxLiquidityPerTick: "+(&MaxLiquidityPerTickRequest{}).Validate().Error())
	_, err = p.Slot0(ctx, nil, &Slot0Request{})
	require.EqualError(t, err, "uniswapv3: slot0: "+(&Slot0Request{}).Validate().Error())
	_, err = p.FeeGrowthGlobal0X128(ctx, nil, &FeeGrowthGlobal0X128Request{})
	require.EqualError(t, err, "uniswapv3: feeGrowthGlobal0X128: "+(&FeeGrowthGlobal0X128Request{}).Validate().Error())
	_, err = p.FeeGrowthGlobal1X128(ctx, nil, &FeeGrowthGlobal1X128Request{})
	require.EqualError(t, err, "uniswapv3: feeGrowthGlobal1X128: "+(&FeeGrowthGlobal1X128Request{}).Validate().Error())
	_, err = p.ProtocolFees(ctx, nil, &ProtocolFeesRequest{})
	require.EqualError(t, err, "uniswapv3: protocolFees: "+(&ProtocolFeesRequest{}).Validate().Error())
	_, err = p.Liquidity(ctx, nil, &LiquidityRequest{})
	require.EqualError(t, err, "uniswapv3: liquidity: "+(&LiquidityRequest{}).Validate().Error())
	_, err = p.Ticks(ctx, nil, &TicksRequest{})
	require.EqualError(t, err, "uniswapv3: ticks: "+(&TicksRequest{}).Validate().Error())
	_, err = p.TickBitmap(ctx, nil, &TickBitmapRequest{})
	require.EqualError(t, err, "uniswapv3: tickBitmap: "+(&TickBitmapRequest{}).Validate().Error())
	_, err = p.Positions(ctx, nil, &PositionsRequest{})
	require.EqualError(t, err, "uniswapv3: positions: "+(&PositionsRequest{}).Validate().Error())
	_, err = p.Observations(ctx, nil, &ObservationsRequest{})
	require.EqualError(t, err, "uniswapv3: observations: "+(&ObservationsRequest{}).Validate().Error())
	_, err = p.Observe(ctx, nil, &ObserveRequest{})
	require.EqualError(t, err, "uniswapv3: observe: "+(&ObserveRequest{}).Validate().Error())
	_, err = p.SnapshotCumulativesInside(ctx, nil, &SnapshotCumulativesInsideRequest{})
	require.EqualError(t, err, "uniswapv3: snapshotCumulativesInside: "+(&SnapshotCumulativesInsideRequest{}).Validate().Error())
	_, _, err = p.CallMint(ctx, nil, &PoolMintRequest{})
	require.EqualError(t, err, "uniswapv3: callMint: "+(&PoolMintRequest{}).Validate().Error())
	_, _, err = p.CallCollect(ctx, nil, &PoolCollectRequest{})
	require.EqualError(t, err, "uniswapv3: callCollect: "+(&PoolCollectRequest{}).Validate().Error())
	_, _, err = p.CallBurn(ctx, nil, &BurnRequest{})
	require.EqualError(t, err, "uniswapv3: callBurn: "+(&BurnRequest{}).Validate().Error())
	_, _, err = p.CallSwap(ctx, nil, &SwapRequest{})
	require.EqualError(t, err, "uniswapv3: callSwap: "+(&SwapRequest{}).Validate().Error())
	_, _, err = p.CallCollectProtocol(ctx, nil, &CollectProtocolRequest{})
	require.EqualError(t, err, "uniswapv3: callCollectProtocol: "+(&CollectProtocolRequest{}).Validate().Error())
	_, _, err = p.Initialize(ctx, opts, &InitializeRequest{})
	require.EqualError(t, err, "uniswapv3: initialize: "+(&InitializeRequest{}).Validate().Error())
	_, _, err = p.Mint(ctx, opts, &PoolMintRequest{})
	require.EqualError(t, err, "uniswapv3: mint: "+(&PoolMintRequest{}).Validate().Error())
	_, _, err = p.Collect(ctx, opts, &PoolCollectRequest{})
	require.EqualError(t, err, "uniswapv3: collect: "+(&PoolCollectRequest{}).Validate().Error())
	_, _, err = p.Burn(ctx, opts, &BurnRequest{})
	require.EqualError(t, err, "uniswapv3: burn: "+(&BurnRequest{}).Validate().Error())
	_, _, err = p.Swap(ctx, opts, &SwapRequest{})
	require.EqualError(t, err, "uniswapv3: swap: "+(&SwapRequest{}).Validate().Error())
	_, _, err = p.Flash(ctx, opts, &FlashRequest{})
	require.EqualError(t, err, "uniswapv3: flash: "+(&FlashRequest{}).Validate().Error())
	_, _, err = p.IncreaseObservationCardinalityNext(ctx, opts, &IncreaseObservationCardinalityNextRequest{})
	require.EqualError(t, err, "uniswapv3: increaseObservationCardinalityNext: "+(&IncreaseObservationCardinalityNextRequest{}).Validate().Error())
	_, _, err = p.SetFeeProtocol(ctx, opts, &SetFeeProtocolRequest{})
	require.EqualError(t, err, "uniswapv3: setFeeProtocol: "+(&SetFeeProtocolRequest{}).Validate().Error())
	_, _, err = p.CollectProtocol(ctx, opts, &CollectProtocolRequest{})
	require.EqualError(t, err, "uniswapv3: collectProtocol: "+(&CollectProtocolRequest{}).Validate().Error())
	require.ErrorIs(t, err, bind.ErrInvalidRequest)

	_, err = p.Ticks(ctx, nil, NewTicksRequest(pool, big.NewInt(1<<23)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = p.Burn(ctx, opts, NewBurnRequest(pool, big.NewInt(-60), big.NewInt(60), new(big.Int).Lsh(big.NewInt(1), 128)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = p.Initialize(ctx, opts, NewInitializeRequest(pool, new(big.Int).Lsh(big.NewInt(1), 160)))
	require.ErrorIs(t, err, abi.ErrIntegerOutOfRange)

	_, _, err = p.CollectProtocol(ctx, nil, NewCollectProtocolRequest(pool, a, big.NewInt(0), big.NewInt(0)))
	require.ErrorIs(t, err, bind.ErrInvalidOptions)
	require.EqualError(t, err, "uniswapv3: collectProtocol: bind: transact: invalid options: options is nil")

}

func TestPoolWritesSignAndSend(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	sender := core.PubkeyToAddress(key.PublicKey())

	initializeData, err := EncodeInitialize(new(big.Int).Lsh(big.NewInt(1), 96))
	require.NoError(t, err)
	mintData, err := EncodeMint(b, big.NewInt(-60), big.NewInt(60), big.NewInt(1000), []byte{0xca, 0xfe})
	require.NoError(t, err)
	collectData, err := EncodeCollect(b, big.NewInt(-60), big.NewInt(60), big.NewInt(5), big.NewInt(6))
	require.NoError(t, err)
	burnData, err := EncodeBurn(big.NewInt(-60), big.NewInt(60), big.NewInt(1000))
	require.NoError(t, err)
	swapData, err := EncodeSwap(b, false, big.NewInt(500), big.NewInt(1), []byte{0x01})
	require.NoError(t, err)
	flashData, err := EncodeFlash(b, big.NewInt(9), big.NewInt(10), []byte{0x02})
	require.NoError(t, err)
	increaseObservationCardinalityNextData, err := EncodeIncreaseObservationCardinalityNext(25)
	require.NoError(t, err)
	setFeeProtocolData, err := EncodeSetFeeProtocol(4, 10)
	require.NoError(t, err)
	collectProtocolData, err := EncodeCollectProtocol(b, big.NewInt(7), big.NewInt(8))
	require.NoError(t, err)

	for _, tc := range []struct {
		name     string
		calldata []byte
		send     func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error)
	}{
		{
			"initialize",
			initializeData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Initialize(t.Context(), opts, NewInitializeRequest(pool, new(big.Int).Lsh(big.NewInt(1), 96)))
			},
		},
		{
			"mint",
			mintData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Mint(t.Context(), opts, NewPoolMintRequest(pool, b, big.NewInt(-60), big.NewInt(60), big.NewInt(1000), []byte{0xca, 0xfe}))
			},
		},
		{
			"collect",
			collectData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Collect(t.Context(), opts, NewPoolCollectRequest(pool, b, big.NewInt(-60), big.NewInt(60), big.NewInt(5), big.NewInt(6)))
			},
		},
		{
			"burn",
			burnData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Burn(t.Context(), opts, NewBurnRequest(pool, big.NewInt(-60), big.NewInt(60), big.NewInt(1000)))
			},
		},
		{
			"swap",
			swapData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Swap(t.Context(), opts, NewSwapRequest(pool, b, false, big.NewInt(500), big.NewInt(1), []byte{0x01}))
			},
		},
		{
			"flash",
			flashData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.Flash(t.Context(), opts, NewFlashRequest(pool, b, big.NewInt(9), big.NewInt(10), []byte{0x02}))
			},
		},
		{
			"increaseObservationCardinalityNext",
			increaseObservationCardinalityNextData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.IncreaseObservationCardinalityNext(t.Context(), opts, NewIncreaseObservationCardinalityNextRequest(pool, 25))
			},
		},
		{
			"setFeeProtocol",
			setFeeProtocolData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.SetFeeProtocol(t.Context(), opts, NewSetFeeProtocolRequest(pool, 4, 10))
			},
		},
		{
			"collectProtocol",
			collectProtocolData,
			func(p *Pool, opts *bind.TransactOptions) (core.Transaction, *types.Hash, error) {
				return p.CollectProtocol(t.Context(), opts, NewCollectProtocolRequest(pool, b, big.NewInt(7), big.NewInt(8)))
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			builder := tx.NewDynamicFeeBuilder()
			builder.ChainID(big.NewInt(1))
			builder.Nonce(7)
			builder.Limit(3_000_000)
			builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

			var raw []byte
			c := mocks.NewClient(t)
			c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, e rpc.Element) error {
				require.Equal(t, rpc.MethodETHSendRawTransaction, e.Method)
				raw, err = encoding.Hex.Decode(e.Params.([]any)[0].(string))
				require.NoError(t, err)
				*e.Result.(*string) = core.Keccak256(raw).String()
				return nil
			}).Once()

			signed, hash, err := tc.send(NewPool(c), bind.NewTransactOptions(key, builder))
			require.NoError(t, err)
			require.True(t, hash.Equal(core.Keccak256(raw)))
			require.Equal(t, uint64(7), signed.AddressNonce())

			decoded, err := core.DecodeTransaction(raw)
			require.NoError(t, err)
			dyn, ok := decoded.(*core.DynamicFeeTx)
			require.True(t, ok)
			require.True(t, dyn.To.Equal(pool))
			require.Zero(t, dyn.Value.Sign())
			require.Equal(t, tc.calldata, dyn.Data)
			got, err := decoded.Sender()
			require.NoError(t, err)
			require.True(t, got.Equal(sender))
		})
	}
}

func TestPoolWritesMapEstimateRevertToSentinel(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	packed, err := abi.Pack(abi.NewTypes(abi.String), reasonAlreadyInitialized)
	require.NoError(t, err)
	revert := append([]byte{0x08, 0xc3, 0x79, 0xa0}, packed...)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(7)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Batch(mock.Anything, mock.Anything).RunAndReturn(func(_ context.Context, elems rpc.Elements) error {
		require.Equal(t, 1, elems.Len())
		require.Equal(t, rpc.MethodETHEstimateGas, elems.GetMethod(0))
		return fmt.Errorf("rpc: %s: %w", rpc.MethodETHEstimateGas, &rpc.ResponseError{Code: 3, Message: "execution reverted", Data: jsontext.Value(`"` + encoding.Hex.EncodePrefixed(revert) + `"`)})
	}).Once()

	signed, hash, err := NewPool(c).Initialize(t.Context(), bind.NewTransactOptions(key, builder), NewInitializeRequest(pool, new(big.Int).Lsh(big.NewInt(1), 96)))
	require.ErrorIs(t, err, ErrAlreadyInitialized)
	require.NotErrorIs(t, err, ErrReverted)
	require.ErrorContains(t, err, "uniswapv3: initialize: ")
	require.Nil(t, signed)
	require.Nil(t, hash)
}

func TestPoolWritesReturnSignedTxOnBroadcastFailure(t *testing.T) {
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(9)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(30_000_000_000), big.NewInt(1_000_000_000))

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).Return(errors.New("dial tcp: connection refused")).Once()

	signed, hash, err := NewPool(c).Collect(t.Context(), bind.NewTransactOptions(key, builder), NewPoolCollectRequest(pool, b, big.NewInt(-60), big.NewInt(60), big.NewInt(1), big.NewInt(1)))
	require.ErrorContains(t, err, "uniswapv3: collect: ")
	require.ErrorContains(t, err, "connection refused")
	require.NotErrorIs(t, err, ErrReverted)
	require.Nil(t, hash)
	require.NotNil(t, signed)
	require.Equal(t, uint64(9), signed.AddressNonce())
}

func TestPoolCallsPropagateContext(t *testing.T) {
	type marker struct{}
	ctx := context.WithValue(t.Context(), marker{}, "pool")
	pool := types.MustNewAddressFromHex("0x8ad599c3a0ff1de082011efddc58f1908eb6e6d8")
	key, err := types.NewPrivateKeyFromBig(big.NewInt(0xb1d))
	require.NoError(t, err)
	word, err := abi.Pack(abi.NewTypes(abi.Uint128), big.NewInt(1))
	require.NoError(t, err)

	c := mocks.NewClient(t)
	c.EXPECT().Call(mock.Anything, mock.Anything).RunAndReturn(func(got context.Context, e rpc.Element) error {
		require.Equal(t, "pool", got.Value(marker{}))
		if e.Method == rpc.MethodETHCall {
			*e.Result.(*string) = encoding.Hex.EncodePrefixed(word)
		}
		return nil
	}).Twice()

	_, err = NewPool(c).Liquidity(ctx, nil, NewLiquidityRequest(pool))
	require.NoError(t, err)

	builder := tx.NewDynamicFeeBuilder()
	builder.ChainID(big.NewInt(1))
	builder.Nonce(1)
	builder.Limit(3_000_000)
	builder.Price(big.NewInt(2), big.NewInt(1))
	_, _, err = NewPool(c).IncreaseObservationCardinalityNext(ctx, bind.NewTransactOptions(key, builder), NewIncreaseObservationCardinalityNextRequest(pool, 2))
	require.NoError(t, err)
}
