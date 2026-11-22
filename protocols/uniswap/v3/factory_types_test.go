package v3

import (
	"math/big"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/ninja-protocol-labs/go-evmkit/bind"
	"github.com/ninja-protocol-labs/go-evmkit/core/types"
)

func TestNewOwnerChangedEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	oldOwner := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	newOwner := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")

	oc := NewOwnerChangedEvent(contract, oldOwner, newOwner)
	require.Same(t, contract, oc.Contract)
	require.Same(t, oldOwner, oc.OldOwner)
	require.Same(t, newOwner, oc.NewOwner)
}

func TestNewPoolCreatedEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	token0 := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	token1 := types.MustNewAddressFromHex("0xbc4ca0eda7647a8ab7c2061c2e118a18a936f13d")
	pool := types.MustNewAddressFromHex("0xb4e16d0168e52d35cacd2c6185b44281ec28c9dc")
	fee := big.NewInt(3000)
	tickSpacing := big.NewInt(60)

	pc := NewPoolCreatedEvent(contract, token0, token1, fee, tickSpacing, pool)
	require.Same(t, contract, pc.Contract)
	require.Same(t, token0, pc.Token0)
	require.Same(t, token1, pc.Token1)
	require.Same(t, fee, pc.Fee)
	require.Same(t, tickSpacing, pc.TickSpacing)
	require.Same(t, pool, pc.Pool)
}

func TestNewFeeAmountEnabledEvent(t *testing.T) {
	contract := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	fee := big.NewInt(500)
	tickSpacing := big.NewInt(10)

	fae := NewFeeAmountEnabledEvent(contract, fee, tickSpacing)
	require.Same(t, contract, fae.Contract)
	require.Same(t, fee, fae.Fee)
	require.Same(t, tickSpacing, fae.TickSpacing)
}

func TestFactoryRequestsValidate(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")
	fee := big.NewInt(3000)

	for _, tc := range []struct {
		name string
		err  error
		want string
	}{
		{"owner nil", (*OwnerRequest)(nil).Validate(), "invalid request: request is nil"},
		{"owner empty", (&OwnerRequest{}).Validate(), "invalid request: contract is nil"},
		{"owner valid", (&OwnerRequest{Contract: factory}).Validate(), ""},
		{"feeAmountTickSpacing nil", (*FeeAmountTickSpacingRequest)(nil).Validate(), "invalid request: request is nil"},
		{"feeAmountTickSpacing empty", (&FeeAmountTickSpacingRequest{}).Validate(), "invalid request: contract is nil"},
		{"feeAmountTickSpacing nil contract", (&FeeAmountTickSpacingRequest{Fee: fee}).Validate(), "invalid request: contract is nil"},
		{"feeAmountTickSpacing nil fee", (&FeeAmountTickSpacingRequest{Contract: factory}).Validate(), "invalid request: fee is nil"},
		{"feeAmountTickSpacing valid", (&FeeAmountTickSpacingRequest{Contract: factory, Fee: fee}).Validate(), ""},
		{"getPool nil", (*GetPoolRequest)(nil).Validate(), "invalid request: request is nil"},
		{"getPool empty", (&GetPoolRequest{}).Validate(), "invalid request: contract is nil"},
		{"getPool nil contract", (&GetPoolRequest{TokenA: a, TokenB: b, Fee: fee}).Validate(), "invalid request: contract is nil"},
		{"getPool nil tokenA", (&GetPoolRequest{Contract: factory, TokenB: b, Fee: fee}).Validate(), "invalid request: tokenA is nil"},
		{"getPool nil tokenB", (&GetPoolRequest{Contract: factory, TokenA: a, Fee: fee}).Validate(), "invalid request: tokenB is nil"},
		{"getPool nil fee", (&GetPoolRequest{Contract: factory, TokenA: a, TokenB: b}).Validate(), "invalid request: fee is nil"},
		{"getPool valid", (&GetPoolRequest{Contract: factory, TokenA: a, TokenB: b, Fee: fee}).Validate(), ""},
		{"createPool nil", (*CreatePoolRequest)(nil).Validate(), "invalid request: request is nil"},
		{"createPool empty", (&CreatePoolRequest{}).Validate(), "invalid request: contract is nil"},
		{"createPool nil contract", (&CreatePoolRequest{TokenA: a, TokenB: b, Fee: fee}).Validate(), "invalid request: contract is nil"},
		{"createPool nil tokenA", (&CreatePoolRequest{Contract: factory, TokenB: b, Fee: fee}).Validate(), "invalid request: tokenA is nil"},
		{"createPool nil tokenB", (&CreatePoolRequest{Contract: factory, TokenA: a, Fee: fee}).Validate(), "invalid request: tokenB is nil"},
		{"createPool nil fee", (&CreatePoolRequest{Contract: factory, TokenA: a, TokenB: b}).Validate(), "invalid request: fee is nil"},
		{"createPool valid", (&CreatePoolRequest{Contract: factory, TokenA: a, TokenB: b, Fee: fee}).Validate(), ""},
		{"setOwner nil", (*SetOwnerRequest)(nil).Validate(), "invalid request: request is nil"},
		{"setOwner empty", (&SetOwnerRequest{}).Validate(), "invalid request: contract is nil"},
		{"setOwner nil contract", (&SetOwnerRequest{Owner: a}).Validate(), "invalid request: contract is nil"},
		{"setOwner nil owner", (&SetOwnerRequest{Contract: factory}).Validate(), "invalid request: owner is nil"},
		{"setOwner valid", (&SetOwnerRequest{Contract: factory, Owner: a}).Validate(), ""},
		{"enableFeeAmount nil", (*EnableFeeAmountRequest)(nil).Validate(), "invalid request: request is nil"},
		{"enableFeeAmount empty", (&EnableFeeAmountRequest{}).Validate(), "invalid request: contract is nil"},
		{"enableFeeAmount nil contract", (&EnableFeeAmountRequest{Fee: fee, TickSpacing: big.NewInt(60)}).Validate(), "invalid request: contract is nil"},
		{"enableFeeAmount nil fee", (&EnableFeeAmountRequest{Contract: factory, TickSpacing: big.NewInt(60)}).Validate(), "invalid request: fee is nil"},
		{"enableFeeAmount nil tickSpacing", (&EnableFeeAmountRequest{Contract: factory, Fee: fee}).Validate(), "invalid request: tickSpacing is nil"},
		{"enableFeeAmount valid", (&EnableFeeAmountRequest{Contract: factory, Fee: fee, TickSpacing: big.NewInt(60)}).Validate(), ""},
		{"getPool identical tokens", (&GetPoolRequest{Contract: factory, TokenA: a, TokenB: a, Fee: fee}).Validate(), ""},
		{"createPool zero token", (&CreatePoolRequest{Contract: factory, TokenA: a, TokenB: types.ZeroAddress(), Fee: fee}).Validate(), ""},
		{"enableFeeAmount negative tickSpacing", (&EnableFeeAmountRequest{Contract: factory, Fee: fee, TickSpacing: big.NewInt(-1)}).Validate(), ""},
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

func TestNewFactoryRequestsSetFields(t *testing.T) {
	factory := types.MustNewAddressFromHex("0x1F98431c8aD98523631AE4a59f267346ea31F984")
	a := types.MustNewAddressFromHex("0x6667c8dc9fbfec411e7c1ee2b24de960149f930f")
	b := types.MustNewAddressFromHex("0xba12222222228d8ba445958a75a0704d566bf2c8")

	require.Equal(t, &OwnerRequest{Contract: factory}, NewOwnerRequest(factory))
	require.Equal(t, &FeeAmountTickSpacingRequest{Contract: factory, Fee: big.NewInt(500)}, NewFeeAmountTickSpacingRequest(factory, big.NewInt(500)))
	require.Equal(t, &GetPoolRequest{Contract: factory, TokenA: a, TokenB: b, Fee: big.NewInt(3000)}, NewGetPoolRequest(factory, a, b, big.NewInt(3000)))
	require.Equal(t, &CreatePoolRequest{Contract: factory, TokenA: b, TokenB: a, Fee: big.NewInt(10000)}, NewCreatePoolRequest(factory, b, a, big.NewInt(10000)))
	require.Equal(t, &SetOwnerRequest{Contract: factory, Owner: a}, NewSetOwnerRequest(factory, a))
	require.Equal(t, &EnableFeeAmountRequest{Contract: factory, Fee: big.NewInt(100), TickSpacing: big.NewInt(1)}, NewEnableFeeAmountRequest(factory, big.NewInt(100), big.NewInt(1)))
}
