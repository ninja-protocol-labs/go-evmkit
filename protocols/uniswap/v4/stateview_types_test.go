package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
	"github.com/ninja-protocol-labs/go-lib-cryptography/encoding"
)

func TestNewSlot0(t *testing.T) {
	sqrtPriceX96 := new(big.Int).Lsh(big.NewInt(1), 96)
	tick := big.NewInt(-20)
	protocolFee := big.NewInt(0)
	lpFee := big.NewInt(3000)

	s := NewSlot0(sqrtPriceX96, tick, protocolFee, lpFee)
	require.Same(t, sqrtPriceX96, s.SqrtPriceX96)
	require.Same(t, tick, s.Tick)
	require.Same(t, protocolFee, s.ProtocolFee)
	require.Same(t, lpFee, s.LPFee)
}

func TestNewTickInfo(t *testing.T) {
	liquidityGross := big.NewInt(1000)
	liquidityNet := big.NewInt(-1000)
	feeGrowthOutside0X128 := big.NewInt(11)
	feeGrowthOutside1X128 := big.NewInt(22)

	info := NewTickInfo(liquidityGross, liquidityNet, feeGrowthOutside0X128, feeGrowthOutside1X128)
	require.Same(t, liquidityGross, info.LiquidityGross)
	require.Same(t, liquidityNet, info.LiquidityNet)
	require.Same(t, feeGrowthOutside0X128, info.FeeGrowthOutside0X128)
	require.Same(t, feeGrowthOutside1X128, info.FeeGrowthOutside1X128)
}

func TestNewTickLiquidity(t *testing.T) {
	liquidityGross := big.NewInt(1000)
	liquidityNet := big.NewInt(-1000)

	l := NewTickLiquidity(liquidityGross, liquidityNet)
	require.Same(t, liquidityGross, l.LiquidityGross)
	require.Same(t, liquidityNet, l.LiquidityNet)
}

func TestNewTickFeeGrowthOutside(t *testing.T) {
	feeGrowthOutside0X128 := big.NewInt(11)
	feeGrowthOutside1X128 := big.NewInt(22)

	o := NewTickFeeGrowthOutside(feeGrowthOutside0X128, feeGrowthOutside1X128)
	require.Same(t, feeGrowthOutside0X128, o.FeeGrowthOutside0X128)
	require.Same(t, feeGrowthOutside1X128, o.FeeGrowthOutside1X128)
}

func TestNewFeeGrowthGlobals(t *testing.T) {
	feeGrowthGlobal0 := big.NewInt(33)
	feeGrowthGlobal1 := big.NewInt(44)

	g := NewFeeGrowthGlobals(feeGrowthGlobal0, feeGrowthGlobal1)
	require.Same(t, feeGrowthGlobal0, g.FeeGrowthGlobal0)
	require.Same(t, feeGrowthGlobal1, g.FeeGrowthGlobal1)
}

func TestNewPositionState(t *testing.T) {
	liquidity := big.NewInt(1000000)
	feeGrowthInside0LastX128 := big.NewInt(55)
	feeGrowthInside1LastX128 := big.NewInt(66)

	p := NewPositionState(liquidity, feeGrowthInside0LastX128, feeGrowthInside1LastX128)
	require.Same(t, liquidity, p.Liquidity)
	require.Same(t, feeGrowthInside0LastX128, p.FeeGrowthInside0LastX128)
	require.Same(t, feeGrowthInside1LastX128, p.FeeGrowthInside1LastX128)
}

func TestNewFeeGrowthInside(t *testing.T) {
	feeGrowthInside0X128 := big.NewInt(77)
	feeGrowthInside1X128 := big.NewInt(88)

	f := NewFeeGrowthInside(feeGrowthInside0X128, feeGrowthInside1X128)
	require.Same(t, feeGrowthInside0X128, f.FeeGrowthInside0X128)
	require.Same(t, feeGrowthInside1X128, f.FeeGrowthInside1X128)
}

func TestCalculatePositionKeyMatchesGolden(t *testing.T) {
	var saltSeven [32]byte
	saltSeven[31] = 7
	var saltFortyTwo [32]byte
	saltFortyTwo[31] = 0x2a
	var saltAllFF [32]byte
	for i := range saltAllFF {
		saltAllFF[i] = 0xff
	}

	cases := []struct {
		name      string
		owner     *types.Address
		tickLower *big.Int
		tickUpper *big.Int
		salt      [32]byte
		want      string
	}{
		{"negativeAndPositiveTicks", testSender, big.NewInt(-600), big.NewInt(600), saltSeven, "0x61a237c56f3a308261c2d9b5974952f42fdc16df84d6cd9c48d1f21459eadfa7"},
		{"int24MinMaxZeroOwner", types.MustNewAddressFromHex("0x0000000000000000000000000000000000000000"), big.NewInt(-8388608), big.NewInt(8388607), [32]byte{}, "0xc6b2825c26e2ee04dc84369a0afe87fd461164512ac7ed191fe5508a18c4b41a"},
		{"positionManagerFullRange", types.MustNewAddressFromHex("0xAc631556d3d4019C95769033B5E719dD77124BAc"), big.NewInt(-887220), big.NewInt(887220), saltFortyTwo, "0x6b0da1990b666de465aab7a849f57fa235923db6a3b9f4db1dc319d77887895c"},
		{"minusOneAndZeroAllFF", types.MustNewAddressFromHex("0xffffffffffffffffffffffffffffffffffffffff"), big.NewInt(-1), big.NewInt(0), saltAllFF, "0xd6067463ef2caad8dded4a28db884bbd17e1b61dbe4095c5b9b07d766cbf2c85"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			key, err := CalculatePositionKey(c.owner, c.tickLower, c.tickUpper, c.salt)
			require.NoError(t, err)
			require.Equal(t, c.want, encoding.Hex.EncodePrefixed(key[:]))
		})
	}
}

func TestCalculatePositionKeyIsSensitiveToEveryField(t *testing.T) {
	var salt [32]byte
	salt[31] = 1
	base, err := CalculatePositionKey(testSender, big.NewInt(-60), big.NewInt(60), salt)
	require.NoError(t, err)

	otherOwner, err := CalculatePositionKey(testCurrency0, big.NewInt(-60), big.NewInt(60), salt)
	require.NoError(t, err)
	require.NotEqual(t, base, otherOwner)

	otherLower, err := CalculatePositionKey(testSender, big.NewInt(-120), big.NewInt(60), salt)
	require.NoError(t, err)
	require.NotEqual(t, base, otherLower)

	otherUpper, err := CalculatePositionKey(testSender, big.NewInt(-60), big.NewInt(120), salt)
	require.NoError(t, err)
	require.NotEqual(t, base, otherUpper)

	swapped, err := CalculatePositionKey(testSender, big.NewInt(60), big.NewInt(-60), salt)
	require.NoError(t, err)
	require.NotEqual(t, base, swapped)

	var otherSalt [32]byte
	otherSalt[0] = 1
	saltChanged, err := CalculatePositionKey(testSender, big.NewInt(-60), big.NewInt(60), otherSalt)
	require.NoError(t, err)
	require.NotEqual(t, base, saltChanged)

	again, err := CalculatePositionKey(testSender, big.NewInt(-60), big.NewInt(60), salt)
	require.NoError(t, err)
	require.Equal(t, base, again)
}

func TestCalculatePositionKeyDoesNotMutateInputs(t *testing.T) {
	tickLower := big.NewInt(-600)
	tickUpper := big.NewInt(600)
	_, err := CalculatePositionKey(testSender, tickLower, tickUpper, [32]byte{})
	require.NoError(t, err)
	require.Zero(t, big.NewInt(-600).Cmp(tickLower))
	require.Zero(t, big.NewInt(600).Cmp(tickUpper))
}

func TestCalculatePositionKeyRejectsInvalidInputs(t *testing.T) {
	cases := []struct {
		name      string
		owner     *types.Address
		tickLower *big.Int
		tickUpper *big.Int
		want      string
	}{
		{"nilOwner", nil, big.NewInt(-60), big.NewInt(60), "owner must not be nil"},
		{"nilTickLower", testSender, nil, big.NewInt(60), "out of int24 range"},
		{"nilTickUpper", testSender, big.NewInt(-60), nil, "out of int24 range"},
		{"tickLowerBelowInt24", testSender, big.NewInt(-8388609), big.NewInt(60), "out of int24 range"},
		{"tickUpperAboveInt24", testSender, big.NewInt(-60), big.NewInt(8388608), "out of int24 range"},
		{"tickFarOutOfRange", testSender, new(big.Int).Lsh(big.NewInt(1), 100), big.NewInt(60), "out of int24 range"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := CalculatePositionKey(c.owner, c.tickLower, c.tickUpper, [32]byte{})
			require.ErrorContains(t, err, c.want)
			require.ErrorContains(t, err, "uniswapv4: calculatePositionKey")
		})
	}
}

func TestStateViewRequestsValidate(t *testing.T) {
	sv := testPoolManager
	owner := testSender
	var poolID PoolID
	poolID[0] = 0xab
	tick := big.NewInt(-60)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"poolManager nil", (*PoolManagerRequest)(nil).Validate(), "invalid request: request is nil"},
		{"poolManager empty", (&PoolManagerRequest{}).Validate(), "invalid request: contract is nil"},
		{"poolManager valid", (&PoolManagerRequest{Contract: sv}).Validate(), ""},
		{"getSlot0 nil", (*GetSlot0Request)(nil).Validate(), "invalid request: request is nil"},
		{"getSlot0 empty", (&GetSlot0Request{}).Validate(), "invalid request: contract is nil"},
		{"getSlot0 zero poolId", (&GetSlot0Request{Contract: sv}).Validate(), ""},
		{"getSlot0 valid", (&GetSlot0Request{Contract: sv, PoolID: poolID}).Validate(), ""},
		{"getLiquidity nil", (*GetLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getLiquidity empty", (&GetLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"getLiquidity valid", (&GetLiquidityRequest{Contract: sv, PoolID: poolID}).Validate(), ""},
		{"getFeeGrowthGlobals nil", (*GetFeeGrowthGlobalsRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getFeeGrowthGlobals empty", (&GetFeeGrowthGlobalsRequest{}).Validate(), "invalid request: contract is nil"},
		{"getFeeGrowthGlobals valid", (&GetFeeGrowthGlobalsRequest{Contract: sv, PoolID: poolID}).Validate(), ""},
		{"getFeeGrowthInside nil", (*GetFeeGrowthInsideRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getFeeGrowthInside empty", (&GetFeeGrowthInsideRequest{}).Validate(), "invalid request: contract is nil"},
		{"getFeeGrowthInside nil contract", (&GetFeeGrowthInsideRequest{TickLower: tick, TickUpper: tick}).Validate(), "invalid request: contract is nil"},
		{"getFeeGrowthInside nil tickLower", (&GetFeeGrowthInsideRequest{Contract: sv, TickUpper: tick}).Validate(), "invalid request: tickLower is nil"},
		{"getFeeGrowthInside nil tickUpper", (&GetFeeGrowthInsideRequest{Contract: sv, TickLower: tick}).Validate(), "invalid request: tickUpper is nil"},
		{"getFeeGrowthInside valid", (&GetFeeGrowthInsideRequest{Contract: sv, PoolID: poolID, TickLower: tick, TickUpper: big.NewInt(60)}).Validate(), ""},
		{"getFeeGrowthInside reversed ticks", (&GetFeeGrowthInsideRequest{Contract: sv, TickLower: big.NewInt(60), TickUpper: tick}).Validate(), ""},
		{"getTickInfo nil", (*GetTickInfoRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getTickInfo empty", (&GetTickInfoRequest{}).Validate(), "invalid request: contract is nil"},
		{"getTickInfo nil tick", (&GetTickInfoRequest{Contract: sv}).Validate(), "invalid request: tick is nil"},
		{"getTickInfo valid", (&GetTickInfoRequest{Contract: sv, PoolID: poolID, Tick: tick}).Validate(), ""},
		{"getTickLiquidity nil", (*GetTickLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getTickLiquidity empty", (&GetTickLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"getTickLiquidity nil tick", (&GetTickLiquidityRequest{Contract: sv}).Validate(), "invalid request: tick is nil"},
		{"getTickLiquidity valid", (&GetTickLiquidityRequest{Contract: sv, Tick: big.NewInt(0)}).Validate(), ""},
		{"getTickFeeGrowthOutside nil", (*GetTickFeeGrowthOutsideRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getTickFeeGrowthOutside empty", (&GetTickFeeGrowthOutsideRequest{}).Validate(), "invalid request: contract is nil"},
		{"getTickFeeGrowthOutside nil tick", (&GetTickFeeGrowthOutsideRequest{Contract: sv}).Validate(), "invalid request: tick is nil"},
		{"getTickFeeGrowthOutside valid", (&GetTickFeeGrowthOutsideRequest{Contract: sv, Tick: tick}).Validate(), ""},
		{"getTickBitmap nil", (*GetTickBitmapRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getTickBitmap empty", (&GetTickBitmapRequest{}).Validate(), "invalid request: contract is nil"},
		{"getTickBitmap negative word", (&GetTickBitmapRequest{Contract: sv, WordPosition: -32768}).Validate(), ""},
		{"getPositionInfo nil", (*GetPositionInfoRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getPositionInfo empty", (&GetPositionInfoRequest{}).Validate(), "invalid request: contract is nil"},
		{"getPositionInfo nil contract", (&GetPositionInfoRequest{Owner: owner, TickLower: tick, TickUpper: tick}).Validate(), "invalid request: contract is nil"},
		{"getPositionInfo nil owner", (&GetPositionInfoRequest{Contract: sv, TickLower: tick, TickUpper: tick}).Validate(), "invalid request: owner is nil"},
		{"getPositionInfo nil tickLower", (&GetPositionInfoRequest{Contract: sv, Owner: owner, TickUpper: tick}).Validate(), "invalid request: tickLower is nil"},
		{"getPositionInfo nil tickUpper", (&GetPositionInfoRequest{Contract: sv, Owner: owner, TickLower: tick}).Validate(), "invalid request: tickUpper is nil"},
		{"getPositionInfo valid", (&GetPositionInfoRequest{Contract: sv, PoolID: poolID, Owner: owner, TickLower: tick, TickUpper: big.NewInt(60), Salt: [32]byte{0x01}}).Validate(), ""},
		{"getPositionInfo zero owner", (&GetPositionInfoRequest{Contract: sv, Owner: types.ZeroAddress(), TickLower: tick, TickUpper: tick}).Validate(), ""},
		{"getPositionInfoByID nil", (*GetPositionInfoByIDRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getPositionInfoByID empty", (&GetPositionInfoByIDRequest{}).Validate(), "invalid request: contract is nil"},
		{"getPositionInfoByID valid", (&GetPositionInfoByIDRequest{Contract: sv, PoolID: poolID, PositionID: [32]byte{0x02}}).Validate(), ""},
		{"stateViewGetPositionLiquidity nil", (*StateViewGetPositionLiquidityRequest)(nil).Validate(), "invalid request: request is nil"},
		{"stateViewGetPositionLiquidity empty", (&StateViewGetPositionLiquidityRequest{}).Validate(), "invalid request: contract is nil"},
		{"stateViewGetPositionLiquidity valid", (&StateViewGetPositionLiquidityRequest{Contract: sv, PoolID: poolID}).Validate(), ""},
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

func TestNewStateViewRequestsSetFields(t *testing.T) {
	sv := testPoolManager
	owner := testSender
	poolID := PoolID{0x11}
	positionID := [32]byte{0x22}
	salt := [32]byte{0x33}

	require.Equal(t, &PoolManagerRequest{Contract: sv}, NewPoolManagerRequest(sv))
	require.Equal(t, &GetSlot0Request{Contract: sv, PoolID: poolID}, NewGetSlot0Request(sv, poolID))
	require.Equal(t, &GetLiquidityRequest{Contract: sv, PoolID: poolID}, NewGetLiquidityRequest(sv, poolID))
	require.Equal(t, &GetFeeGrowthGlobalsRequest{Contract: sv, PoolID: poolID}, NewGetFeeGrowthGlobalsRequest(sv, poolID))
	require.Equal(t, &GetFeeGrowthInsideRequest{Contract: sv, PoolID: poolID, TickLower: big.NewInt(-600), TickUpper: big.NewInt(600)}, NewGetFeeGrowthInsideRequest(sv, poolID, big.NewInt(-600), big.NewInt(600)))
	require.Equal(t, &GetTickInfoRequest{Contract: sv, PoolID: poolID, Tick: big.NewInt(-887220)}, NewGetTickInfoRequest(sv, poolID, big.NewInt(-887220)))
	require.Equal(t, &GetTickLiquidityRequest{Contract: sv, PoolID: poolID, Tick: big.NewInt(60)}, NewGetTickLiquidityRequest(sv, poolID, big.NewInt(60)))
	require.Equal(t, &GetTickFeeGrowthOutsideRequest{Contract: sv, PoolID: poolID, Tick: big.NewInt(120)}, NewGetTickFeeGrowthOutsideRequest(sv, poolID, big.NewInt(120)))
	require.Equal(t, &GetTickBitmapRequest{Contract: sv, PoolID: poolID, WordPosition: -58}, NewGetTickBitmapRequest(sv, poolID, -58))
	require.Equal(t, &GetPositionInfoRequest{Contract: sv, PoolID: poolID, Owner: owner, TickLower: big.NewInt(-60), TickUpper: big.NewInt(60), Salt: salt}, NewGetPositionInfoRequest(sv, poolID, owner, big.NewInt(-60), big.NewInt(60), salt))
	require.Equal(t, &GetPositionInfoByIDRequest{Contract: sv, PoolID: poolID, PositionID: positionID}, NewGetPositionInfoByIDRequest(sv, poolID, positionID))
	require.Equal(t, &StateViewGetPositionLiquidityRequest{Contract: sv, PoolID: poolID, PositionID: positionID}, NewStateViewGetPositionLiquidityRequest(sv, poolID, positionID))

	tickLower := big.NewInt(-60)
	req := NewGetPositionInfoRequest(sv, poolID, owner, tickLower, big.NewInt(60), salt)
	require.Same(t, owner, req.Owner)
	require.Same(t, tickLower, req.TickLower)
}
