package v4

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

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
